package a1en_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestNegationContextPolarity(t *testing.T) {
	negative, negativeContext, ok, err := interlingua.Negation("negation.classification.this")
	if err != nil || !ok || negativeContext == nil {
		t.Fatalf("load negative classification fixture: ok=%t context=%t error=%v", ok, negativeContext != nil, err)
	}
	positive, positiveOK, err := interlingua.Identification("identification.classification.this")
	if err != nil || !positiveOK {
		t.Fatalf("load positive classification fixture: ok=%t error=%v", positiveOK, err)
	}
	runtime := matrix.Runtime(t)
	target, exists := runtime.TargetRealizerFor(features.LanguageID("en"))
	if !exists || target == nil {
		t.Fatal("INFRA: English target realizer is unavailable")
	}
	cases := []struct {
		name           string
		meaning        meaning.Meaning
		contextMeaning meaning.Meaning
		context        meaning.RequestContext
	}{
		{name: "negative-meaning-positive-context", meaning: negative, contextMeaning: positive.Graph, context: positive.Context},
		{name: "positive-meaning-negative-context", meaning: positive.Graph, contextMeaning: negative, context: negativeContext.Context},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := semantic.MeaningCandidate{
				ID: tc.name, SemanticID: tc.name, Meaning: &tc.meaning, Status: status.OutcomeOK,
				Evidence: semantic.EvidenceTrace{AlternativeID: tc.name, Outcome: status.OutcomeOK},
				Contexts: []semantic.MeaningContext{{Meaning: tc.contextMeaning, Context: tc.context}},
			}
			result, realizeErr := target.Realize(t.Context(), candidate)
			t.Logf("candidate=%s outcome=%s error=%v alternatives=%+v", tc.name, result.Outcome, realizeErr, result.Alternatives)
			if realizeErr != nil {
				return
			}
			for _, alternative := range result.Alternatives {
				if alternative.Text != "" {
					t.Errorf("mismatched graph/context polarity emitted text %q", alternative.Text)
				}
			}
			if result.Outcome != status.OutcomeUnsupported && result.Outcome != status.OutcomeUnknown {
				t.Errorf("mismatched graph/context polarity was not refused: outcome=%s", result.Outcome)
			}
		})
	}
}
