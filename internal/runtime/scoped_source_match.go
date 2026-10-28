package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

// ScopedSourceReadingPolicy validates language-local morphology, case, and
// feature constraints for one slot reading. Returning false rejects that
// reading; errors abort matching without retaining partial alternatives.
type ScopedSourceReadingPolicy func(context.Context, construction.ScopedClauseVariant, construction.ScopedClausePart, analyzer.ReadingCandidate, analyzer.SenseCandidate) (bool, error)

// ScopedSourceOptions supplies the language tokenizer, finite DSE budget, and
// the mandatory language-local slot validator.
type ScopedSourceOptions struct {
	Context context.Context
	Tokens  text.TokenizeOptions
	Budget  semantic.DSEBudget
	Policy  ScopedSourceReadingPolicy
	// AllowFinalPeriod mirrors SourcePipeline.AnalyzeSyntax: one eligible
	// terminal period may be absent from the lexical lattice and source plan.
	AllowFinalPeriod bool
}

// ScopedSourceBinding is one occurrence of a binder, retaining its own token,
// morphology reading, lexeme, and lexical sense evidence.
type ScopedSourceBinding struct {
	Binder     string               `json:"binder"`
	Case       string               `json:"case,omitempty"`
	TokenIndex int                  `json:"token_index"`
	Span       text.Span            `json:"span"`
	Reading    morph.Reading        `json:"reading"`
	Lexeme     lexicon.MorphLexeme  `json:"lexeme"`
	Sense      lexicon.LexicalSense `json:"sense"`
}

// ScopedSourceAlternative preserves a complete surface-plan and all of its
// independently selected slot readings. It does not construct references or AST.
type ScopedSourceAlternative struct {
	VariantID     string                `json:"variant_id"`
	Negation      string                `json:"negation"`
	SubjectBinder string                `json:"subject_binder"`
	Bindings      []ScopedSourceBinding `json:"bindings"`
}

// ScopedSourceMatch retains every complete interpretation within the budget.
type ScopedSourceMatch struct {
	Status       status.OutcomeKind        `json:"status"`
	Complete     bool                      `json:"complete"`
	Alternatives []ScopedSourceAlternative `json:"alternatives"`
	Work         int                       `json:"work"`
	OutputBytes  int                       `json:"output_bytes"`
	Diagnostics  []status.Diagnostic       `json:"diagnostics,omitempty"`
}

// MatchScopedSource matches tokenized literals and slot occurrences against
// the full lexical lattice. Any bound exhaustion or cancellation fails closed.
func MatchScopedSource(plan construction.ScopedClausePlan, lattice analyzer.LexicalLattice, source string, options ScopedSourceOptions) (ScopedSourceMatch, error) {
	ctx := options.Context
	if ctx == nil {
		return scopedSourceFailure(status.OutcomeInvalid, "SCOPED_SOURCE_CONTEXT_REQUIRED", "a non-nil context is required"), nil
	}
	budget, err := semantic.NormalizeDSEBudget(options.Budget)
	if err != nil {
		return ScopedSourceMatch{}, err
	}
	if options.Policy == nil {
		return scopedSourceFailure(status.OutcomeUnsupported, "SCOPED_SOURCE_POLICY_REQUIRED", "language-local reading policy is required"), nil
	}
	if plan.Language == "" || lattice.Language != plan.Language || len(plan.Variants) == 0 {
		return scopedSourceFailure(status.OutcomeInvalid, "SCOPED_SOURCE_LANGUAGE_MISMATCH", "plan and lexical lattice must use the same language"), nil
	}
	if lattice.Outcome == status.OutcomeResourceLimit {
		return scopedSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "lexical lattice reports a resource limit"), nil
	}
	if lattice.Outcome == status.OutcomeInvalid || lattice.Outcome == status.OutcomeError {
		return scopedSourceFailure(lattice.Outcome, "SCOPED_SOURCE_LATTICE_FAILED", "lexical lattice reports an invalid or failed analysis"), nil
	}
	scanSource := source
	strippedFinalPeriod := options.AllowFinalPeriod && strings.HasSuffix(source, ".") && len(source) > 1 && !strings.ContainsAny(source[:len(source)-1], ".?!")
	if strippedFinalPeriod {
		scanSource = source[:len(source)-1]
	}
	tokens, err := text.Tokenize(scanSource, options.Tokens)
	if err != nil {
		return ScopedSourceMatch{}, err
	}
	sourceHasFinalPeriod := len(tokens) > 0 && tokens[len(tokens)-1].Surface == "."
	latticeHasFinalPeriod := len(lattice.Tokens) > 0 && lattice.Tokens[len(lattice.Tokens)-1].Token.Surface == "."
	allowPlanFinalPeriod := !sourceHasFinalPeriod && !latticeHasFinalPeriod
	if len(tokens) != len(lattice.Tokens) {
		if lattice.Outcome == status.OutcomeIncomplete {
			if !scopedLiteralPlanMatches(plan, tokens, options.Tokens, allowPlanFinalPeriod) {
				return scopedSourceNoMatch(), nil
			}
			return scopedSourceFailure(status.OutcomeIncomplete, "SCOPED_SOURCE_TOKEN_COVERAGE", "the matching scoped source has incomplete lexical lattice token coverage"), nil
		}
		return scopedSourceFailure(status.OutcomeIncomplete, "SCOPED_SOURCE_TOKEN_COVERAGE", "lexical lattice does not cover every source token"), nil
	}
	for i := range tokens {
		if lattice.Tokens[i].Index != i || lattice.Tokens[i].Token.Span != tokens[i].Span || lattice.Tokens[i].Token.Surface != tokens[i].Surface {
			if lattice.Outcome == status.OutcomeIncomplete {
				if !scopedLiteralPlanMatches(plan, tokens, options.Tokens, allowPlanFinalPeriod) {
					return scopedSourceNoMatch(), nil
				}
				return scopedSourceFailure(status.OutcomeIncomplete, "SCOPED_SOURCE_TOKEN_COVERAGE", "the matching scoped source has inconsistent incomplete lattice token positions"), nil
			}
			return scopedSourceFailure(status.OutcomeIncomplete, "SCOPED_SOURCE_TOKEN_COVERAGE", "lexical lattice token positions differ from tokenized source"), nil
		}
	}
	result := ScopedSourceMatch{Status: status.OutcomeUnsupported, Complete: true, Alternatives: []ScopedSourceAlternative{}}
	incomplete := false
	unknownReading := false
	matchedLiteralVariant := false
	for _, variant := range plan.Variants {
		if err := ctx.Err(); err != nil {
			return scopedSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "scoped source matching was canceled"), nil
		}
		if result.Work >= budget.MaxWork {
			return scopedSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "scoped source matching work limit exceeded"), nil
		}
		result.Work++
		alts, matched, literalMatched, work, failure := matchScopedVariant(ctx, variant, tokens, lattice.Tokens, options, budget, result.Work, allowPlanFinalPeriod)
		result.Work = work
		if failure != nil {
			if failure == errScopedResource {
				return scopedSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "scoped source matching work limit exceeded"), nil
			}
			if failure == errScopedUnknown {
				if literalMatched {
					matchedLiteralVariant = true
					incomplete = true
					unknownReading = true
					continue
				}
				return scopedSourceFailure(status.OutcomeIncomplete, "SCOPED_SOURCE_UNKNOWN_READING", "a slot contains unknown lexical analysis"), nil
			}
			if errors.Is(failure, context.Canceled) || errors.Is(failure, context.DeadlineExceeded) {
				return scopedSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "scoped source matching was canceled"), nil
			}
			return ScopedSourceMatch{}, failure
		}
		if !matched {
			if literalMatched {
				matchedLiteralVariant = true
			}
			continue
		}
		matchedLiteralVariant = true
		for _, alt := range alts {
			if len(result.Alternatives) >= budget.MaxAlternatives {
				return scopedSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "scoped source alternative limit exceeded"), nil
			}
			result.Alternatives = append(result.Alternatives, alt)
			encoded, marshalErr := json.Marshal(result.Alternatives)
			if marshalErr != nil {
				return ScopedSourceMatch{}, marshalErr
			}
			if len(encoded) > budget.MaxOutputBytes {
				return scopedSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "scoped source output limit exceeded"), nil
			}
			result.OutputBytes = len(encoded)
		}
	}
	if lattice.Outcome == status.OutcomeIncomplete && matchedLiteralVariant {
		incomplete = true
		result.Diagnostics = append(result.Diagnostics, status.Diagnostic{Code: "SCOPED_SOURCE_LATTICE_INCOMPLETE", Message: "the matching source variant has incomplete lexical analysis; omitted readings may change its interpretation", Severity: status.SeverityWarning})
	}
	if incomplete {
		result.Status = status.OutcomeIncomplete
		result.Complete = false
		if unknownReading {
			result.Diagnostics = append(result.Diagnostics, status.Diagnostic{Code: "SCOPED_SOURCE_UNKNOWN_READING", Message: "one or more required lexical slots contain unknown or incomplete lexical analysis", Severity: status.SeverityWarning})
		}
		return result, nil
	}
	if len(result.Alternatives) == 0 {
		result.Status = status.OutcomeUnsupported
		result.Complete = false
		result.Diagnostics = []status.Diagnostic{{Code: "SCOPED_SOURCE_NO_MATCH", Message: "no complete scoped clause variant matched", Severity: status.SeverityInfo}}
		return result, nil
	}
	result.Status = status.OutcomeOK
	if len(result.Alternatives) > 1 {
		result.Status = status.OutcomeAmbiguous
	}
	return result, nil
}

func matchScopedVariant(ctx context.Context, variant construction.ScopedClauseVariant, sourceTokens []text.Token, lattice []analyzer.TokenNode, options ScopedSourceOptions, budget semantic.DSEBudget, work int, allowPlanFinalPeriod bool) ([]ScopedSourceAlternative, bool, bool, int, error) {
	// Keep the plan's interleaved part order while expanding literals to tokens.
	ordered := make([]construction.ScopedClausePart, 0)
	for _, p := range variant.Parts {
		if p.Literal != "" {
			ts, err := text.Tokenize(p.Literal, options.Tokens)
			if err != nil {
				return nil, false, false, work, err
			}
			for _, t := range ts {
				ordered = append(ordered, construction.ScopedClausePart{Literal: t.Normalized})
			}
		} else {
			ordered = append(ordered, p)
		}
	}
	if allowPlanFinalPeriod && len(ordered) > 0 && ordered[len(ordered)-1].Literal == "." {
		ordered = ordered[:len(ordered)-1]
	}
	if len(ordered) != len(sourceTokens) {
		return nil, false, false, work, nil
	}
	// Validate every literal before looking at any slot. A later mismatch means
	// this variant is irrelevant and cannot make an unrelated reading fatal.
	for i, part := range ordered {
		if part.Literal == "" {
			continue
		}
		if work >= budget.MaxWork {
			return nil, false, false, work, errScopedResource
		}
		work++
		if normalizedToken(sourceTokens[i], options.Tokens) != normalizedLiteral(part.Literal, options.Tokens) {
			return nil, false, false, work, nil
		}
	}
	literalMatched := true
	bindings := make([]ScopedSourceBinding, 0)
	alternatives := make([]ScopedSourceAlternative, 0)
	localOutputBytes := 0
	var walk func(int, map[string]ScopedSourceBinding) error
	walk = func(pos int, prior map[string]ScopedSourceBinding) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		if work >= budget.MaxWork {
			return errScopedResource
		}
		work++
		if pos == len(ordered) {
			alt := ScopedSourceAlternative{VariantID: variant.ID, Negation: variant.Negation, SubjectBinder: variant.SubjectBinder, Bindings: append([]ScopedSourceBinding(nil), bindings...)}
			if len(alternatives) >= budget.MaxAlternatives {
				return errScopedResource
			}
			encoded, err := json.Marshal(alt)
			if err != nil {
				return err
			}
			localOutputBytes += len(encoded)
			if localOutputBytes > budget.MaxOutputBytes {
				return errScopedResource
			}
			alternatives = append(alternatives, alt)
			return nil
		}
		part, tok := ordered[pos], sourceTokens[pos]
		if part.Literal != "" {
			return walk(pos+1, prior)
		}
		node := lattice[pos]
		if tok.Kind != text.TokenWord {
			return nil
		}
		if node.Outcome == status.OutcomeResourceLimit {
			return errScopedResource
		}
		if (node.Outcome != status.OutcomeOK && node.Outcome != status.OutcomeAmbiguous) || len(node.Readings) == 0 {
			return errScopedUnknown
		}
		for _, candidate := range node.Readings {
			if work >= budget.MaxWork {
				return errScopedResource
			}
			work++
			if candidate.State != analyzer.CandidateKnown || !candidate.HasLexeme || len(candidate.Senses) == 0 || candidate.Reading.LexemeID == "" {
				return errScopedUnknown
			}
			for _, sense := range candidate.Senses {
				if work >= budget.MaxWork {
					return errScopedResource
				}
				work++
				if sense.Sense.ID.String() == "" || sense.Sense.LexemeID.String() != candidate.Lexeme.ID.String() || candidate.Reading.LexemeID != candidate.Lexeme.ID.String() {
					continue
				}
				accept, err := options.Policy(ctx, variant, part, candidate, sense)
				if err != nil {
					return err
				}
				if !accept {
					continue
				}
				if previous, ok := prior[part.Slot]; ok && (previous.Lexeme.ID.String() != candidate.Lexeme.ID.String() || previous.Sense.ID.String() != sense.Sense.ID.String()) {
					continue
				}
				binding := ScopedSourceBinding{Binder: part.Slot, Case: part.Case, TokenIndex: node.Index, Span: tok.Span, Reading: candidate.Reading, Lexeme: candidate.Lexeme, Sense: sense.Sense}
				bindings = append(bindings, binding)
				next := cloneScopedBindings(prior)
				if part.Slot != "event" {
					next[part.Slot] = binding
				}
				if err := walk(pos+1, next); err != nil {
					return err
				}
				bindings = bindings[:len(bindings)-1]
			}
		}
		return nil
	}
	err := walk(0, map[string]ScopedSourceBinding{})
	if err != nil {
		return nil, false, literalMatched, work, err
	}
	return alternatives, len(alternatives) > 0, literalMatched, work, nil
}

var errScopedResource = fmt.Errorf("scoped source work limit exceeded")
var errScopedUnknown = fmt.Errorf("unknown scoped source reading")

func cloneScopedBindings(in map[string]ScopedSourceBinding) map[string]ScopedSourceBinding {
	out := make(map[string]ScopedSourceBinding, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
func normalizedToken(t text.Token, opts text.TokenizeOptions) string {
	if opts.Profile != nil && opts.Profile.CaseFold {
		if t.Folded != "" {
			return t.Folded
		}
	}
	return t.Normalized
}
func normalizedLiteral(value string, opts text.TokenizeOptions) string {
	ts, err := text.Tokenize(value, opts)
	if err != nil || len(ts) != 1 {
		return ""
	}
	return normalizedToken(ts[0], opts)
}
func scopedSourceFailure(kind status.OutcomeKind, code, message string) ScopedSourceMatch {
	return ScopedSourceMatch{Status: kind, Complete: false, Alternatives: []ScopedSourceAlternative{}, Diagnostics: []status.Diagnostic{{Code: code, Message: message, Severity: status.SeverityError}}}
}

func scopedSourceNoMatch() ScopedSourceMatch {
	return ScopedSourceMatch{Status: status.OutcomeUnsupported, Complete: false, Alternatives: []ScopedSourceAlternative{}, Diagnostics: []status.Diagnostic{{Code: "SCOPED_SOURCE_NO_MATCH", Message: "no complete scoped clause variant matched", Severity: status.SeverityInfo}}}
}

func scopedLiteralPlanMatches(plan construction.ScopedClausePlan, tokens []text.Token, options text.TokenizeOptions, allowFinalPeriod bool) bool {
	for _, variant := range plan.Variants {
		ordered := make([]construction.ScopedClausePart, 0, len(variant.Parts))
		valid := true
		for _, part := range variant.Parts {
			if part.Literal != "" {
				literalTokens, err := text.Tokenize(part.Literal, options)
				if err != nil {
					valid = false
					break
				}
				for _, token := range literalTokens {
					ordered = append(ordered, construction.ScopedClausePart{Literal: token.Normalized})
				}
			} else {
				ordered = append(ordered, part)
			}
		}
		if !valid {
			continue
		}
		if allowFinalPeriod && len(ordered) > 0 && ordered[len(ordered)-1].Literal == "." {
			ordered = ordered[:len(ordered)-1]
		}
		if len(ordered) != len(tokens) {
			continue
		}
		matched := true
		for index, part := range ordered {
			if part.Literal != "" && normalizedToken(tokens[index], options) != normalizedLiteral(part.Literal, options) {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}
