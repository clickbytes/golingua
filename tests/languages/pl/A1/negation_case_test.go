package a1pl_test

import (
	"context"
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestNegativeCaseRealizationContrasts(t *testing.T) {
	book, ok, err := interlingua.Contrast("m5.neg.seeing.teacher-book")
	if err != nil || !ok {
		t.Fatalf("load negative seeing/book Meaning: ok=%t error=%v", ok, err)
	}
	birdHearing, ok, err := interlingua.NegationExpansion("a1.8.neg.019")
	if err != nil || !ok {
		t.Fatalf("load negative hearing/birds Meaning: ok=%t error=%v", ok, err)
	}
	naTarget, err := negatedLookingAtBook()
	if err != nil {
		t.Fatalf("build na+acc contrast Meaning: %v", err)
	}

	runtime := matrix.Runtime(t)
	realizer, exists := runtime.TargetRealizerFor(features.LanguageID("pl"))
	if !exists || realizer == nil {
		t.Fatal("INFRA: Polish target realizer unavailable")
	}
	for _, tc := range []struct {
		id      string
		meaning meaning.Meaning
		want    string
	}{
		{"negative-seeing-book", book, "Nauczyciel nie widzi książki."},
		{"negative-hearing-birds", birdHearing, "Matka nie słyszy ptaków."},
		{"negative-looking-na-accusative", naTarget, "Nauczyciel nie patrzy na książkę."},
	} {
		t.Run(tc.id, func(t *testing.T) {
			candidate := semantic.MeaningCandidate{
				ID: tc.id, SemanticID: tc.id, Meaning: &tc.meaning, Status: status.OutcomeOK,
				Evidence: semantic.EvidenceTrace{AlternativeID: tc.id, Outcome: status.OutcomeOK},
			}
			result, err := realizer.Realize(context.Background(), candidate)
			if err != nil {
				t.Fatalf("realize: %v", err)
			}
			t.Logf("outcome=%s alternatives=%+v", result.Outcome, result.Alternatives)
			for _, alternative := range result.Alternatives {
				if alternative.Text == tc.want {
					return
				}
			}
			t.Errorf("expected surface %q; got %+v", tc.want, result.Alternatives)
		})
	}
}

func TestNegativeSeeingBookAnalysisKeepsSingularStimulus(t *testing.T) {
	runtime := matrix.Runtime(t)
	analyzer, ok := runtime.SourceAnalyzerFor(features.LanguageID("pl"))
	if !ok || analyzer == nil {
		t.Fatal("INFRA: Polish source analyzer unavailable")
	}
	result, err := analyzer.Analyze(context.Background(), "Nauczyciel nie widzi książki.")
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}
	t.Logf("outcome=%s candidates=%+v", result.Outcome, result.Candidates)
	for _, candidate := range result.Candidates {
		if candidate.Meaning == nil {
			continue
		}
		bookID, seeingID, negative := meaning.NodeID(""), meaning.NodeID(""), false
		for _, node := range candidate.Meaning.Nodes {
			switch {
			case node.Kind == meaning.KindReferent && string(node.Concept) == "concept:book" && node.Multiplicity == meaning.MultiplicityOne:
				bookID = node.ID
			case node.Kind == meaning.KindEvent && string(node.Concept) == "concept:seeing":
				seeingID = node.ID
			case node.Kind == meaning.KindScope && node.Operator == meaning.OperatorNegation:
				negative = true
			}
		}
		stimulus := false
		for _, node := range candidate.Meaning.Nodes {
			if node.ID != seeingID {
				continue
			}
			for _, participant := range node.Participants {
				if participant.Role == "stimulus" && participant.Target == bookID {
					stimulus = true
				}
			}
		}
		if bookID != "" && seeingID != "" && stimulus && negative {
			return
		}
	}
	t.Errorf("no analysis candidate preserves singular book as the negated seeing stimulus")
}

func negatedLookingAtBook() (meaning.Meaning, error) {
	looking, err := ids.NewConceptID("looking")
	if err != nil {
		return meaning.Meaning{}, err
	}
	teacherConcept, err := ids.NewConceptID("teacher")
	if err != nil {
		return meaning.Meaning{}, err
	}
	bookConcept, err := ids.NewConceptID("book")
	if err != nil {
		return meaning.Meaning{}, err
	}
	event, err := meaning.NewEvent("e", looking, []meaning.Participant{{Role: "agent", Target: "t"}, {Role: "target", Target: "b"}})
	if err != nil {
		return meaning.Meaning{}, err
	}
	event.TemporalFrame = meaning.TemporalFramePresent
	teacher, err := meaning.NewReferent("t", teacherConcept)
	if err != nil {
		return meaning.Meaning{}, err
	}
	teacher.Multiplicity = meaning.MultiplicityOne
	book, err := meaning.NewReferent("b", bookConcept)
	if err != nil {
		return meaning.Meaning{}, err
	}
	book.Multiplicity = meaning.MultiplicityOne
	neg, err := meaning.NewNegation("neg", "e")
	if err != nil {
		return meaning.Meaning{}, err
	}
	return meaning.Build([]meaning.NodeID{"neg"}, []meaning.Node{neg, event, teacher, book}, meaning.DefaultBudget())
}
