package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

// RelationSourceReadingPolicy validates one language-local reading of one
// compiled-plan slot. It does not generate morphology or interpret symbols.
type RelationSourceReadingPolicy func(context.Context, engine.RelationPathVariant, engine.RelationPathPart, analyzer.ReadingCandidate, analyzer.SenseCandidate) (bool, error)

// RelationSourceOptions bounds matching and supplies language-local token and
// reading policies. AllowFinalPeriod mirrors SourcePipeline.AnalyzeSyntax.
type RelationSourceOptions struct {
	Context          context.Context
	Tokens           text.TokenizeOptions
	Budget           semantic.DSEBudget
	Policy           RelationSourceReadingPolicy
	AllowFinalPeriod bool
}

type RelationSourcePartEvidence struct {
	PartIndex           int                   `json:"part_index"`
	Step                string                `json:"step"`
	Slot                string                `json:"slot,omitempty"`
	Literal             string                `json:"literal,omitempty"`
	Port                string                `json:"port,omitempty"`
	PortPosition        string                `json:"port_position,omitempty"`
	GrammaticalFunction string                `json:"grammatical_function,omitempty"`
	TokenIndex          int                   `json:"token_index"`
	Span                text.Span             `json:"span"`
	Surface             string                `json:"surface"`
	Reading             *morph.Reading        `json:"reading,omitempty"`
	Lexeme              *lexicon.MorphLexeme  `json:"lexeme,omitempty"`
	Sense               *lexicon.LexicalSense `json:"sense,omitempty"`
}

type RelationSourceAlternative struct {
	VariantID string                       `json:"variant_id"`
	Parts     []RelationSourcePartEvidence `json:"parts"`
}

type RelationSourceMatch struct {
	Status       status.OutcomeKind          `json:"status"`
	Complete     bool                        `json:"complete"`
	Alternatives []RelationSourceAlternative `json:"alternatives"`
	Work         int                         `json:"work"`
	OutputBytes  int                         `json:"output_bytes"`
	Diagnostics  []status.Diagnostic         `json:"diagnostics,omitempty"`
}

type relationPlanToken struct {
	partIndex int
	part      engine.RelationPathPart
	literal   string
	verb      bool
}

type relationSourceReading struct {
	reading morph.Reading
	lexeme  lexicon.MorphLexeme
	sense   lexicon.LexicalSense
}

// MatchRelationPathSource matches a strict compiled relation plan against the
// complete lexical lattice. It retains every compatible pair of independent
// verb readings that identifies the same lexeme and sense.
func MatchRelationPathSource(plan engine.RelationPathPlan, lattice analyzer.LexicalLattice, source string, options RelationSourceOptions) (RelationSourceMatch, error) {
	ctx := options.Context
	if ctx == nil {
		return relationSourceFailure(status.OutcomeInvalid, "RELATION_SOURCE_CONTEXT_REQUIRED", "a non-nil context is required"), nil
	}
	budget, err := semantic.NormalizeDSEBudget(options.Budget)
	if err != nil {
		return RelationSourceMatch{}, err
	}
	if options.Policy == nil {
		return relationSourceFailure(status.OutcomeUnsupported, "RELATION_SOURCE_POLICY_REQUIRED", "a language-local reading policy is required"), nil
	}
	if lattice.Language == "" {
		return relationSourceFailure(status.OutcomeInvalid, "RELATION_SOURCE_LANGUAGE_MISSING", "lexical lattice must declare its language"), nil
	}
	if lattice.Outcome == status.OutcomeResourceLimit {
		return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "lexical lattice reports a resource limit"), nil
	}
	if lattice.Outcome == status.OutcomeInvalid || lattice.Outcome == status.OutcomeError {
		return relationSourceFailure(lattice.Outcome, "RELATION_SOURCE_LATTICE_FAILED", "lexical lattice reports an invalid or failed analysis"), nil
	}
	encodedPlan, err := json.Marshal(plan)
	if err != nil || len(encodedPlan) > engine.MaxRelationPlanBytes {
		return relationSourceFailure(status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "compiled relation path plan exceeds its finite representation limit"), nil
	}
	strictPlan, err := engine.DecodeRelationPathPlan(encodedPlan, string(lattice.Language))
	if err != nil {
		return relationSourceFailure(status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "compiled relation path plan failed strict decoding: "+err.Error()), nil
	}
	scanSource := source
	strippedFinalPeriod := options.AllowFinalPeriod && strings.HasSuffix(source, ".") && len(source) > 1 && !strings.ContainsAny(source[:len(source)-1], ".?!")
	if strippedFinalPeriod {
		scanSource = source[:len(source)-1]
	}
	tokens, err := text.Tokenize(scanSource, options.Tokens)
	if err != nil {
		return RelationSourceMatch{}, err
	}
	sourceHasFinalPeriod := len(tokens) > 0 && tokens[len(tokens)-1].Surface == "."
	latticeHasFinalPeriod := len(lattice.Tokens) > 0 && lattice.Tokens[len(lattice.Tokens)-1].Token.Surface == "."
	allowPlanFinalPeriod := !sourceHasFinalPeriod && !latticeHasFinalPeriod
	if len(tokens) != len(lattice.Tokens) {
		if lattice.Outcome == status.OutcomeIncomplete {
			if !relationLiteralPlanMatches(strictPlan, tokens, options.Tokens, allowPlanFinalPeriod) {
				return relationSourceNoMatch("no compiled relation path variant covered the complete source token sequence"), nil
			}
			return relationSourceFailure(status.OutcomeIncomplete, "RELATION_SOURCE_TOKEN_COVERAGE", "the matching relation source has incomplete lexical lattice token coverage"), nil
		}
		return relationSourceFailure(status.OutcomeIncomplete, "RELATION_SOURCE_TOKEN_COVERAGE", "lexical lattice does not cover every token in the original source"), nil
	}
	for i := range tokens {
		if lattice.Tokens[i].Index != i || lattice.Tokens[i].Token.Span != tokens[i].Span || lattice.Tokens[i].Token.Surface != tokens[i].Surface {
			if lattice.Outcome == status.OutcomeIncomplete {
				if !relationLiteralPlanMatches(strictPlan, tokens, options.Tokens, allowPlanFinalPeriod) {
					return relationSourceNoMatch("no compiled relation path variant covered the complete source token sequence"), nil
				}
				return relationSourceFailure(status.OutcomeIncomplete, "RELATION_SOURCE_TOKEN_COVERAGE", "the matching relation source has inconsistent incomplete lattice token positions"), nil
			}
			return relationSourceFailure(status.OutcomeIncomplete, "RELATION_SOURCE_TOKEN_COVERAGE", "lexical lattice token positions differ from tokenized original source"), nil
		}
	}
	result := RelationSourceMatch{Status: status.OutcomeUnsupported, Complete: true, Alternatives: []RelationSourceAlternative{}, Diagnostics: []status.Diagnostic{}}
	matchedLiteralPlan, incomplete := false, false
	for _, variant := range strictPlan.Variants {
		if err := ctx.Err(); err != nil {
			return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source matching was canceled"), nil
		}
		if !relationSourceCharge(&result, budget.MaxWork, 1) {
			return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source matching work limit exceeded"), nil
		}
		if !relationSourceCharge(&result, budget.MaxWork, len(variant.Parts)) {
			return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source plan expansion work limit exceeded"), nil
		}
		ordered, err := expandRelationVariant(variant, options.Tokens)
		if err != nil {
			return RelationSourceMatch{}, err
		}
		if !relationSourceCharge(&result, budget.MaxWork, len(ordered)) {
			return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source token expansion work limit exceeded"), nil
		}
		if allowPlanFinalPeriod && len(ordered) > 0 && ordered[len(ordered)-1].literal == "." && !ordered[len(ordered)-1].verb && ordered[len(ordered)-1].part.Slot == "" {
			ordered = ordered[:len(ordered)-1]
		}
		if len(ordered) != len(tokens) {
			continue
		}
		literalMatch := true
		for index, item := range ordered {
			if err := ctx.Err(); err != nil {
				return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source matching was canceled"), nil
			}
			if item.verb {
				continue
			}
			if !relationSourceCharge(&result, budget.MaxWork, 1) {
				return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source matching work limit exceeded"), nil
			}
			if normalizedToken(tokens[index], options.Tokens) != normalizedLiteral(item.literal, options.Tokens) {
				literalMatch = false
				break
			}
		}
		if !literalMatch {
			continue
		}
		matchedLiteralPlan = true
		readingLists := make([][]relationSourceReading, 0, 2)
		verbPositions := make([]int, 0, 2)
		variantIncomplete := false
		for index, item := range ordered {
			if !item.verb {
				continue
			}
			verbPositions = append(verbPositions, index)
			readings, unknown, work, policyErr := relationSourceReadings(ctx, lattice.Tokens[index], variant, item.part, options.Policy, budget.MaxWork-result.Work)
			result.Work += work
			if policyErr != nil {
				if errors.Is(policyErr, errRelationSourceResource) {
					return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source reading scan exceeded its finite work limit"), nil
				}
				if errors.Is(policyErr, context.Canceled) || errors.Is(policyErr, context.DeadlineExceeded) || ctx.Err() != nil {
					return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source matching was canceled"), nil
				}
				return RelationSourceMatch{}, policyErr
			}
			if result.Work > budget.MaxWork {
				return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source reading scan exceeded its finite work limit"), nil
			}
			if unknown {
				variantIncomplete = true
			}
			readingLists = append(readingLists, readings)
		}
		if len(verbPositions) != 2 || len(readingLists) != 2 {
			return relationSourceFailure(status.OutcomeInvalid, "RELATION_PATH_PLAN_INVALID", "strictly decoded relation plan did not provide exactly two verb positions"), nil
		}
		if variantIncomplete {
			incomplete = true
			relationSourceDiagnostic(&result, status.Diagnostic{Code: "RELATION_SOURCE_UNKNOWN_READING", Path: "variant." + variant.ID, Message: "one or both verb slots contain unknown or incomplete lexical analysis", Severity: status.SeverityWarning})
		}
		for _, first := range readingLists[0] {
			for _, second := range readingLists[1] {
				if err := ctx.Err(); err != nil {
					return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source matching was canceled"), nil
				}
				if !relationSourceCharge(&result, budget.MaxWork, 1) {
					return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source combination work limit exceeded"), nil
				}
				if first.lexeme.ID != second.lexeme.ID || first.lexeme.Language != second.lexeme.Language || first.lexeme.Lemma != second.lexeme.Lemma || first.lexeme.POS != second.lexeme.POS || first.sense.ID != second.sense.ID || first.sense.LexemeID != second.sense.LexemeID {
					continue
				}
				alternative := RelationSourceAlternative{VariantID: variant.ID, Parts: make([]RelationSourcePartEvidence, 0, len(ordered))}
				firstVerb := true
				for planIndex, item := range ordered {
					token := tokens[planIndex]
					latticeNode := lattice.Tokens[planIndex]
					partEvidence := RelationSourcePartEvidence{
						PartIndex: item.partIndex, Step: item.part.Step, Slot: item.part.Slot,
						Literal: item.part.Literal, Port: item.part.Port, PortPosition: item.part.Port,
						GrammaticalFunction: item.part.GrammaticalFunction, TokenIndex: latticeNode.Index,
						Span: token.Span, Surface: token.Surface,
					}
					if item.verb {
						chosen := second
						if firstVerb {
							chosen, firstVerb = first, false
						}
						partEvidence.Reading = cloneRelationSourceReading(chosen.reading)
						lexeme := cloneRelationSourceLexeme(chosen.lexeme)
						sense := cloneRelationSourceSense(chosen.sense)
						partEvidence.Lexeme, partEvidence.Sense = &lexeme, &sense
					}
					alternative.Parts = append(alternative.Parts, partEvidence)
				}
				result.Alternatives = append(result.Alternatives, alternative)
				if len(result.Alternatives) > budget.MaxAlternatives {
					return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source alternative limit exceeded"), nil
				}
				encoded, marshalErr := json.Marshal(result.Alternatives)
				if marshalErr != nil {
					return RelationSourceMatch{}, marshalErr
				}
				if len(encoded) > budget.MaxOutputBytes {
					return relationSourceFailure(status.OutcomeResourceLimit, "RESOURCE_LIMIT", "relation source output limit exceeded"), nil
				}
				result.OutputBytes = len(encoded)
			}
		}
	}
	if !matchedLiteralPlan {
		result.Complete = false
		result.Diagnostics = append(result.Diagnostics, status.Diagnostic{Code: "RELATION_SOURCE_NO_MATCH", Message: "no compiled relation path variant covered the complete source token sequence", Severity: status.SeverityInfo})
		return result, nil
	}
	if lattice.Outcome == status.OutcomeIncomplete {
		incomplete = true
		relationSourceDiagnostic(&result, status.Diagnostic{Code: "RELATION_SOURCE_LATTICE_INCOMPLETE", Message: "the lexical lattice is incomplete, so source readings may have been omitted", Severity: status.SeverityWarning})
	}
	if incomplete {
		result.Status, result.Complete = status.OutcomeIncomplete, false
		return result, nil
	}
	if len(result.Alternatives) == 0 {
		result.Status, result.Complete = status.OutcomeUnsupported, false
		result.Diagnostics = append(result.Diagnostics, status.Diagnostic{Code: "RELATION_SOURCE_NO_MATCH", Message: "no pair of known verb readings shared one exact lexical sense", Severity: status.SeverityInfo})
		return result, nil
	}
	result.Status = status.OutcomeOK
	if len(result.Alternatives) > 1 {
		result.Status = status.OutcomeAmbiguous
	}
	return result, nil
}

func relationSourceNoMatch(message string) RelationSourceMatch {
	return RelationSourceMatch{Status: status.OutcomeUnsupported, Complete: false, Alternatives: []RelationSourceAlternative{}, Diagnostics: []status.Diagnostic{{Code: "RELATION_SOURCE_NO_MATCH", Message: message, Severity: status.SeverityInfo}}}
}

func relationLiteralPlanMatches(plan engine.RelationPathPlan, tokens []text.Token, options text.TokenizeOptions, allowFinalPeriod bool) bool {
	for _, variant := range plan.Variants {
		ordered, err := expandRelationVariant(variant, options)
		if err != nil {
			continue
		}
		if allowFinalPeriod && len(ordered) > 0 && ordered[len(ordered)-1].literal == "." && !ordered[len(ordered)-1].verb && ordered[len(ordered)-1].part.Slot == "" {
			ordered = ordered[:len(ordered)-1]
		}
		if len(ordered) != len(tokens) {
			continue
		}
		matched := true
		for index, item := range ordered {
			if !item.verb && normalizedToken(tokens[index], options) != normalizedLiteral(item.literal, options) {
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

func expandRelationVariant(variant engine.RelationPathVariant, tokenizeOptions text.TokenizeOptions) ([]relationPlanToken, error) {
	ordered := make([]relationPlanToken, 0, len(variant.Parts))
	for partIndex, part := range variant.Parts {
		if part.Slot == "verb" {
			ordered = append(ordered, relationPlanToken{partIndex: partIndex, part: part, verb: true})
			continue
		}
		tokens, err := text.Tokenize(part.Literal, tokenizeOptions)
		if err != nil {
			return nil, err
		}
		for _, token := range tokens {
			ordered = append(ordered, relationPlanToken{partIndex: partIndex, part: part, literal: token.Normalized})
		}
	}
	return ordered, nil
}

func relationSourceReadings(ctx context.Context, node analyzer.TokenNode, variant engine.RelationPathVariant, part engine.RelationPathPart, policy RelationSourceReadingPolicy, maxWork int) ([]relationSourceReading, bool, int, error) {
	if maxWork < 1 {
		return nil, false, 1, errRelationSourceResource
	}
	work := 0
	if err := ctx.Err(); err != nil {
		return nil, false, work, err
	}
	if node.Outcome == status.OutcomeResourceLimit {
		return nil, false, work, errRelationSourceResource
	}
	if (node.Outcome != status.OutcomeOK && node.Outcome != status.OutcomeAmbiguous) || len(node.Readings) == 0 || node.Token.Kind != text.TokenWord {
		return nil, true, work, nil
	}
	readings := make([]relationSourceReading, 0)
	unknown := false
	for _, candidate := range node.Readings {
		work++
		if work > maxWork {
			return nil, false, work, errRelationSourceResource
		}
		if candidate.State != analyzer.CandidateKnown || !candidate.HasLexeme || candidate.Reading.LexemeID == "" || candidate.Reading.LexemeID != candidate.Lexeme.ID.String() || len(candidate.Senses) == 0 {
			unknown = true
			continue
		}
		for _, sense := range candidate.Senses {
			work++
			if work > maxWork {
				return nil, false, work, errRelationSourceResource
			}
			if sense.Sense.ID.String() == "" || sense.Sense.LexemeID != candidate.Lexeme.ID {
				unknown = true
				continue
			}
			accept, err := policy(ctx, variant, part, candidate, sense)
			if err != nil {
				return nil, false, work, err
			}
			if accept {
				readings = append(readings, relationSourceReading{reading: candidate.Reading, lexeme: candidate.Lexeme, sense: sense.Sense})
			}
		}
	}
	return readings, unknown, work, nil
}

var errRelationSourceResource = fmt.Errorf("relation source finite work limit exceeded")

func relationSourceCharge(result *RelationSourceMatch, maxWork, amount int) bool {
	if amount < 0 || result.Work > maxWork-amount {
		return false
	}
	result.Work += amount
	return true
}

func relationSourceFailure(kind status.OutcomeKind, code, message string) RelationSourceMatch {
	return RelationSourceMatch{Status: kind, Complete: false, Alternatives: []RelationSourceAlternative{}, Diagnostics: []status.Diagnostic{{Code: code, Message: message, Severity: status.SeverityError}}}
}

func relationSourceDiagnostic(result *RelationSourceMatch, diagnostic status.Diagnostic) {
	if len(result.Diagnostics) < status.MaxDiagnostics {
		result.Diagnostics = append(result.Diagnostics, diagnostic)
	}
}

func cloneRelationSourceReading(reading morph.Reading) *morph.Reading {
	copy := reading
	copy.Features.Values = append(copy.Features.Values[:0:0], reading.Features.Values...)
	for index := range copy.Features.Values {
		if variable := copy.Features.Values[index].Variable; variable != nil {
			variableCopy := *variable
			copy.Features.Values[index].Variable = &variableCopy
		}
	}
	return &copy
}

func cloneRelationSourceLexeme(lexeme lexicon.MorphLexeme) lexicon.MorphLexeme {
	copy := lexeme
	copy.Properties = append(copy.Properties[:0:0], lexeme.Properties...)
	return copy
}

func cloneRelationSourceSense(sense lexicon.LexicalSense) lexicon.LexicalSense {
	copy := sense
	copy.Usage = append(copy.Usage[:0:0], sense.Usage...)
	copy.ValencyIDs = append(copy.ValencyIDs[:0:0], sense.ValencyIDs...)
	return copy
}
