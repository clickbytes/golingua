package pl

import (
	"context"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
)

// scopedSourceReadingPolicy requires explicit Polish case, number, and gender
// for nouns, and the bounded present imperfective finite verb profile.
func ScopedSourceReadingPolicy(ctx context.Context, variant construction.ScopedClauseVariant, part construction.ScopedClausePart, candidate analyzer.ReadingCandidate, sense analyzer.SenseCandidate) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if !candidate.HasLexeme || candidate.State != analyzer.CandidateKnown || candidate.Reading.LexemeID == "" || candidate.Reading.LexemeID != candidate.Lexeme.ID.String() || candidate.Lexeme.POS == "" || candidate.Reading.POS != string(candidate.Lexeme.POS) || sense.Sense.ID.String() == "" || sense.Sense.LexemeID.String() != candidate.Lexeme.ID.String() {
		return false, nil
	}
	seenSense := false
	for _, listed := range candidate.Senses {
		if listed.Sense.ID == sense.Sense.ID && listed.Sense.LexemeID == sense.Sense.LexemeID {
			seenSense = true
			break
		}
	}
	if !seenSense {
		return false, nil
	}
	values := make(map[string]string, len(candidate.Reading.Features.Values))
	for _, f := range candidate.Reading.Features.Values {
		if _, exists := values[string(f.Feature)]; exists {
			return false, nil
		}
		if f.State != features.StateKnown || f.Value == "" || f.Variable != nil {
			return false, nil
		}
		values[string(f.Feature)] = string(f.Value)
	}
	get := func(id, want string) bool { return values[id] == want }
	switch part.Slot {
	case "x", "y":
		if string(candidate.Lexeme.POS) != "noun" || candidate.Reading.POS != "noun" || part.Case == "" || !get("case", part.Case) || !get("number", "singular") || values["gender"] == "" {
			return false, nil
		}
	case "event":
		if string(candidate.Lexeme.POS) != "verb" || candidate.Reading.POS != "verb" || !get("verb_form", "finite") || !get("tense", "present") || !get("person", "third") || !get("mood", "indicative") || !get("voice", "active") || !get("number", "singular") || !get("aspect", "imperfective") {
			return false, nil
		}
	default:
		return false, nil
	}
	for _, c := range variant.Constraints {
		if c.Binder != part.Slot {
			continue
		}
		feature := c.Feature
		if part.Slot == "event" && feature == "temporal_frame" {
			feature = "tense"
		}
		value, ok := values[feature]
		if !ok {
			return false, nil
		}
		allowed := false
		for _, v := range c.Values {
			if v == value {
				allowed = true
				break
			}
		}
		if !allowed {
			return false, nil
		}
	}
	return true, nil
}
