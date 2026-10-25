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

func TestIdentificationNativeRealization(t *testing.T) {
	matrix.RunRealization(t, features.LanguageID("en"), []matrix.RealizationCase{
		{
			ID: "I-EN-classification", MeaningID: "identification.classification.this",
			ProposalTexts: []string{"This is a cat."}, AllowedTexts: []string{"This is a cat."}, RequiredTexts: []string{"This is a cat."},
			AllowedAssumptionFields: []string{"tense", "person", "number"},
			ReferenceOrigin:         "authored interlingua fixture", OwnerAcceptance: "NOT_EVALUATED",
		},
		{
			ID: "I-EN-black", MeaningID: "identification.black.this",
			ProposalTexts: []string{"This cat is black."}, AllowedTexts: []string{"This cat is black."}, RequiredTexts: []string{"This cat is black."},
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
	realizer, exists := runtime.TargetRealizerFor(features.LanguageID("en"))
	if !exists || realizer == nil {
		t.Fatal("INFRA: English target realizer is unavailable")
	}
	candidate := semantic.MeaningCandidate{
		ID: "I-EN-black-unresolved", SemanticID: "I-EN-black-unresolved", Meaning: &fixture.Graph,
		Status: status.OutcomeOK, Evidence: semantic.EvidenceTrace{AlternativeID: "I-EN-black-unresolved", Outcome: status.OutcomeOK},
		Contexts: []semantic.MeaningContext{{Meaning: fixture.Graph, Context: contextCopy}},
	}
	result, err := realizer.Realize(t.Context(), candidate)
	if err != nil {
		t.Fatalf("target realization returned an error: %v", err)
	}
	if result.Outcome != status.OutcomeUnsupported && result.Outcome != status.OutcomeUnknown {
		t.Errorf("unresolved black reference must remain unrendered and unresolved, got %s: %+v", result.Outcome, result.Diagnostics)
	}
	for _, alternative := range result.Alternatives {
		if alternative.Text != "" {
			t.Errorf("unresolved black reference produced target text %q", alternative.Text)
		}
	}
}
