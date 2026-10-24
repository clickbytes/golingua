package a1pl_test

import (
	"testing"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestIdentificationNativeRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("pl"), []matrix.RealizationCase{
		{
			ID: "I-PL-classification", MeaningID: "identification.classification.to",
			ProposalTexts: []string{"To jest kot."}, AllowedTexts: []string{"To jest kot."},
			AllowedAssumptionFields: []string{"tense", "person", "number"},
			ReferenceOrigin:         "authored interlingua fixture", OwnerAcceptance: "NOT_EVALUATED",
		},
		{
			ID: "I-PL-black", MeaningID: "identification.black.ten",
			ProposalTexts: []string{"Ten kot jest czarny."}, AllowedTexts: []string{"Ten kot jest czarny."},
			AllowedAssumptionFields: []string{"tense", "person", "number"},
			ReferenceOrigin:         "authored interlingua fixture", OwnerAcceptance: "NOT_EVALUATED",
		},
	})
}

func TestIdentificationUnresolvedBlack(t *testing.T) {
	fixture, ok, err := interlingua.Identification("identification.black.this")
	if err != nil || !ok {
		t.Fatalf("load black Meaning fixture: ok=%t error=%v", ok, err)
	}
	contextCopy := fixture.Context
	contextCopy.Referents = append([]meaning.Referent(nil), fixture.Context.Referents...)
	contextCopy.Mentions = append([]meaning.Mention(nil), fixture.Context.Mentions...)
	if len(contextCopy.Mentions) != 1 || contextCopy.Mentions[0].Anchor == nil {
		t.Fatal("black fixture must contain one anchored mention")
	}
	anchor := *contextCopy.Mentions[0].Anchor
	anchor.Value = "unresolved"
	contextCopy.Mentions[0].Anchor = &anchor

	runtime := matrix.Runtime(t)
	realizer, exists := runtime.TargetRealizerFor(features.LanguageID("pl"))
	if !exists || realizer == nil {
		t.Fatal("INFRA: Polish target realizer is unavailable")
	}
	candidate := semantic.MeaningCandidate{
		ID: "I-PL-black-unresolved", SemanticID: "I-PL-black-unresolved", Meaning: &fixture.Graph,
		Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: "I-PL-black-unresolved", Outcome: status.OutcomeOK},
		Contexts: []semantic.MeaningContext{{Meaning: fixture.Graph, Context: contextCopy}},
	}
	result, err := realizer.Realize(t.Context(), candidate)
	if err != nil {
		t.Fatalf("target realization returned an error: %v", err)
	}
	// This negative control requires refusal to choose a demonstrative. An
	// unknown reference and an unsupported route are both refusals, never proof
	// that the missing reference mode has been realized successfully.
	if result.Outcome != status.OutcomeUnsupported && result.Outcome != status.OutcomeUnknown {
		t.Errorf("unresolved black reference must remain unrendered and unresolved, got %s: %+v", result.Outcome, result.Diagnostics)
	}
	for _, alternative := range result.Alternatives {
		if alternative.Text != "" {
			t.Errorf("unresolved black reference produced target text %q", alternative.Text)
		}
	}
}
