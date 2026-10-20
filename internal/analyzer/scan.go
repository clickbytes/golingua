package analyzer

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/morph"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/internal/text"
)

// validLanguageID mirrors the canonical lowercase ASCII tag contract used by
// feature profiles, without exposing that package's validation helper.
func validLanguageID(language features.LanguageID) bool {
	value := string(language)
	if len(value) == 0 || len(value) > 35 || strings.ToLower(value) != value {
		return false
	}
	parts := strings.Split(value, "-")
	if len(parts[0]) < 2 || len(parts[0]) > 8 {
		return false
	}
	for index, part := range parts {
		if len(part) == 0 || len(part) > 8 {
			return false
		}
		for _, character := range part {
			if (character < 'a' || character > 'z') && !(index > 0 && character >= '0' && character <= '9') {
				return false
			}
		}
	}
	return true
}

// BuildLattice scans bounded input and retains every morphology reading,
// lexical sense, and semantic binding that fits the caller's alternative
// budget. A capped scan returns its partial lattice with OutcomeIncomplete.
func BuildLattice(
	input string,
	language features.LanguageID,
	tokenOptions text.TokenizeOptions,
	morphology *morph.InverseIndex,
	lexical *lexicon.Index,
	mwePlans []construction.CompiledMWEPlan,
	budget Budget,
) (LexicalLattice, error) {
	if err := budget.Validate(); err != nil {
		return LexicalLattice{}, err
	}
	if !validLanguageID(language) {
		return LexicalLattice{}, latticeInvalid("language", "requested language ID must be a canonical lowercase ASCII tag")
	}
	if morphology == nil {
		return LexicalLattice{}, latticeUnavailable("MORPHOLOGY_UNAVAILABLE", "morphology", "morphology evidence was not supplied")
	}
	if lexical == nil {
		return LexicalLattice{}, latticeUnavailable("LEXICON_UNAVAILABLE", "lexicon", "lexical evidence was not supplied")
	}
	if tokenOptions.MaxInputBytes == 0 || tokenOptions.MaxInputBytes > budget.MaxInputBytes {
		tokenOptions.MaxInputBytes = budget.MaxInputBytes
	}
	if tokenOptions.MaxInputBytes < 0 {
		return LexicalLattice{}, &text.TextError{Kind: text.TextErrorInvalidLimit, Limit: tokenOptions.MaxInputBytes, Detail: "maximum input bytes must be positive"}
	}
	if tokenOptions.MaxTokens < 0 {
		return LexicalLattice{}, &text.TextError{Kind: text.TextErrorInvalidLimit, Limit: tokenOptions.MaxTokens, Detail: "maximum token count must be positive"}
	}
	tokenLimit := budget.MaxTokens
	if tokenOptions.MaxTokens > 0 && tokenOptions.MaxTokens < tokenLimit {
		tokenLimit = tokenOptions.MaxTokens
	}
	inputLimit := tokenOptions.MaxInputBytes
	scanInput := input
	inputLimitHit := len(input) > inputLimit
	if inputLimitHit {
		cut := inputLimit
		for cut > 0 && !utf8.RuneStart(input[cut]) {
			cut--
		}
		scanInput = input[:cut]
	}
	// A token needs at least one source byte. Tokenize up to this finite upper
	// bound, then retain the configured prefix so token-budget exhaustion can
	// return an incomplete lattice with the first omitted source span.
	tokenOptions.MaxTokens = inputLimit
	tokens, err := text.Tokenize(scanInput, tokenOptions)
	if err != nil {
		return LexicalLattice{}, err
	}
	if inputLimitHit && len(tokens) != 0 && tokens[len(tokens)-1].Span.End == len(scanInput) && !nextRuneIsDelimiter(input, len(scanInput), tokenOptions.Profile) {
		// The final token may have been cut mid-surface. Do not invent a
		// morphology candidate from that incomplete surface.
		tokens = tokens[:len(tokens)-1]
	}
	tokenLimitHit := len(tokens) > tokenLimit
	var omittedSpan text.Span
	if tokenLimitHit {
		omittedSpan = tokens[tokenLimit].Span
		tokens = tokens[:tokenLimit]
	}
	lattice := LexicalLattice{
		Language: language,
		Tokens:   make([]TokenNode, len(tokens)),
		Outcome:  status.OutcomeOK,
	}
	for index, token := range tokens {
		lattice.Tokens[index] = TokenNode{Index: index, Token: token, Outcome: status.OutcomeOK}
	}
	if inputLimitHit {
		lattice.Outcome = status.OutcomeIncomplete
		appendLatticeDiagnostic(&lattice, status.Diagnostic{
			Code: "RESOURCE_LIMIT", Path: fmt.Sprintf("input.bytes[%d:%d]", len(scanInput), len(input)),
			Message:  "input scan stopped at the configured byte budget; the remaining source bytes were omitted",
			Severity: status.SeverityWarning,
		})
	}
	if tokenLimitHit {
		lattice.Outcome = status.OutcomeIncomplete
		appendLatticeDiagnostic(&lattice, status.Diagnostic{
			Code: "RESOURCE_LIMIT", Path: spanPath(tokenLimit, omittedSpan),
			Message:  "token scan stopped at the configured token budget; this and later token spans were omitted",
			Severity: status.SeverityWarning,
		})
	}

	used := 0
	candidateBudgetHit := false
	for tokenIndex := range lattice.Tokens {
		node := &lattice.Tokens[tokenIndex]
		if node.Token.Kind != text.TokenWord {
			continue
		}

		readings, analyzeErr := morphology.Analyze(node.Token.Surface)
		if analyzeErr != nil {
			var morphFailure *morph.MorphError
			if errors.As(analyzeErr, &morphFailure) {
				result := morphFailure.Result()
				if result.Status == status.OutcomeUnknown && hasDiagnosticCode(result.Diagnostics, "UNKNOWN_SURFACE_FORM") {
					node.Outcome = status.OutcomeUnknown
					setLatticeOutcome(&lattice, status.OutcomeUnknown)
					appendLatticeDiagnostic(&lattice, status.Diagnostic{
						Code: "UNKNOWN_SURFACE", Path: spanPath(tokenIndex, node.Token.Span),
						Message: "no licensed morphology reading matches this token surface", Severity: status.SeverityWarning,
					})
					continue
				}
				return LexicalLattice{}, NewLatticeError(cloneStatusResult(result))
			}
			return LexicalLattice{}, analyzeErr
		}
		if len(readings) == 0 {
			node.Outcome = status.OutcomeUnknown
			setLatticeOutcome(&lattice, status.OutcomeUnknown)
			appendLatticeDiagnostic(&lattice, status.Diagnostic{
				Code: "UNKNOWN_SURFACE", Path: spanPath(tokenIndex, node.Token.Span),
				Message: "no licensed morphology reading matches this token surface", Severity: status.SeverityWarning,
			})
			continue
		}

		for _, reading := range readings {
			if reading.Features.Language != language {
				return LexicalLattice{}, latticeInvalid(
					"morphology.readings.language",
					fmt.Sprintf("reading language %q does not match requested language %q", reading.Features.Language, language),
				)
			}
			if used >= budget.MaxCandidates {
				markCandidateLimit(&lattice, tokenIndex, node.Token.Span, used, budget.MaxCandidates)
				candidateBudgetHit = true
				break
			}

			lexemeID, parseErr := ids.ParseLexemeID(reading.LexemeID)
			if parseErr != nil {
				return LexicalLattice{}, latticeInvalid("morphology.readings.lexeme_id", "morphology reading contains an invalid lexeme ID")
			}
			candidate := ReadingCandidate{Reading: reading}
			lexeme, found := lexical.ByLexemeID(lexemeID)
			if !found {
				// This is an explicit positive-index miss. It does not claim that
				// the surface itself is an unsupported lexical item.
				candidate.State = CandidateUnknownLexeme
				used++
				node.Readings = append(node.Readings, candidate)
				continue
			}
			if features.LanguageID(lexeme.Language) != language {
				return LexicalLattice{}, latticeInvalid(
					"lexicon.lexemes.language",
					fmt.Sprintf("lexeme language %q does not match requested language %q", lexeme.Language, language),
				)
			}
			candidate.Lexeme = lexeme
			candidate.HasLexeme = true
			candidate.State = CandidateKnown
			used++ // One unit for the morphology reading.

			senses := lexical.SensesForLexeme(lexemeID)
			if len(senses) == 0 {
				// The explicit index query above establishes that this known
				// lexeme has no sense records in the supplied snapshot.
				candidate.State = CandidateUnknownSense
				node.Readings = append(node.Readings, candidate)
				continue
			}
			for _, sense := range senses {
				if used >= budget.MaxCandidates {
					markCandidateLimit(&lattice, tokenIndex, node.Token.Span, used, budget.MaxCandidates)
					candidateBudgetHit = true
					break
				}
				used++ // One unit for each sense alternative.
				senseCandidate := SenseCandidate{Sense: sense}
				for _, binding := range lexical.BindingsForSense(sense.ID) {
					if used >= budget.MaxCandidates {
						candidate.Senses = append(candidate.Senses, senseCandidate)
						markCandidateLimit(&lattice, tokenIndex, node.Token.Span, used, budget.MaxCandidates)
						candidateBudgetHit = true
						break
					}
					senseCandidate.Bindings = append(senseCandidate.Bindings, binding)
					used++ // Bindings are retained alternatives too.
				}
				if candidateBudgetHit {
					break
				}
				candidate.Senses = append(candidate.Senses, senseCandidate)
			}
			if candidateBudgetHit {
				node.Readings = append(node.Readings, candidate)
				break
			}
			node.Readings = append(node.Readings, candidate)
		}
		if candidateBudgetHit {
			for remaining := tokenIndex + 1; remaining < len(lattice.Tokens); remaining++ {
				if lattice.Tokens[remaining].Token.Kind == text.TokenWord {
					lattice.Tokens[remaining].Outcome = status.OutcomeIncomplete
				}
			}
			break
		}
	}

	if err := matchMWEs(&lattice, mwePlans, lexical, budget, &used); err != nil {
		return LexicalLattice{}, err
	}
	return lattice, nil
}

func nextRuneIsDelimiter(input string, offset int, profile *text.TokenizerProfile) bool {
	if offset >= len(input) {
		return false
	}
	next, _ := utf8.DecodeRuneInString(input[offset:])
	if !unicode.IsSpace(next) {
		return false
	}
	return profile == nil || profile.WordRune == nil || !profile.WordRune(next)
}

func hasDiagnosticCode(diagnostics []status.Diagnostic, code string) bool {
	for _, diagnostic := range diagnostics {
		if diagnostic.Code == code {
			return true
		}
	}
	return false
}

func setLatticeOutcome(lattice *LexicalLattice, outcome status.OutcomeKind) {
	if lattice.Outcome == status.OutcomeIncomplete {
		return
	}
	if outcome == status.OutcomeIncomplete || lattice.Outcome == status.OutcomeOK {
		lattice.Outcome = outcome
	}
}

func appendLatticeDiagnostic(lattice *LexicalLattice, diagnostic status.Diagnostic) {
	if len(lattice.Diagnostics) < status.MaxDiagnostics {
		lattice.Diagnostics = append(lattice.Diagnostics, diagnostic)
		return
	}
	// Keep the diagnostic list valid while recording that later token-level
	// details are represented by the affected nodes and spans.
	last := &lattice.Diagnostics[status.MaxDiagnostics-1]
	if diagnostic.Code == "RESOURCE_LIMIT" {
		if last.Code == "RESOURCE_LIMIT" {
			diagnostic = status.Diagnostic{
				Code: "RESOURCE_LIMIT", Path: "input/tokens/candidates/mwe",
				Message:  "one or more scan, candidate, or MWE limits were reached; inspect partial token and MWE spans",
				Severity: status.SeverityWarning,
			}
		}
		*last = diagnostic
		return
	}
	if last.Code != "ADDITIONAL_SCAN_DIAGNOSTICS" && last.Code != "RESOURCE_LIMIT" {
		*last = status.Diagnostic{
			Code: "ADDITIONAL_SCAN_DIAGNOSTICS", Path: "tokens",
			Message:  "additional token findings are omitted; inspect token outcomes and spans",
			Severity: status.SeverityWarning,
		}
	}
}

func markCandidateLimit(lattice *LexicalLattice, tokenIndex int, span text.Span, used, limit int) {
	lattice.Outcome = status.OutcomeIncomplete
	lattice.Tokens[tokenIndex].Outcome = status.OutcomeIncomplete
	appendLatticeDiagnostic(lattice, status.Diagnostic{
		Code: "RESOURCE_LIMIT", Path: spanPath(tokenIndex, span),
		Message:  fmt.Sprintf("candidate alternatives stopped at budget %d (already used %d)", limit, used),
		Severity: status.SeverityWarning,
	})
}

func spanPath(tokenIndex int, span text.Span) string {
	return fmt.Sprintf("tokens[%d].span[%d:%d]", tokenIndex, span.Start, span.End)
}

func latticeUnavailable(code, path, message string) *LatticeError {
	return NewLatticeError(status.Result{
		Status: status.OutcomeUnknown,
		Diagnostics: []status.Diagnostic{{
			Code: code, Path: path, Message: message, Severity: status.SeverityError,
		}},
	})
}

func latticeInvalid(path, message string) *LatticeError {
	return NewLatticeError(status.Result{
		Status: status.OutcomeInvalid,
		Diagnostics: []status.Diagnostic{{
			Code: "INVALID_LATTICE_INPUT", Path: path, Message: message, Severity: status.SeverityError,
		}},
	})
}
