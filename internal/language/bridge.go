package language

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type Bridge struct {
	provider     LanguageProvider
	binding      Binding
	capabilities []VerifiedCapability
}

const MaxAnalysisCandidates = 1024
const MaxRealizedSurfaceBytes = 1 << 20

func (registry *Registry) Bridge(language features.LanguageID, senseID ids.SenseID, frameID string, familyID ids.ConstructionID) (Bridge, error) {
	pack, ok := registry.Language(language)
	if !ok {
		return Bridge{}, unknown("UNKNOWN_LANGUAGE_PACKAGE", "language", "language package is not registered")
	}
	provider, ok := registry.Provider(language)
	if !ok {
		return Bridge{}, unsupported("LANGUAGE_PROVIDER_UNAVAILABLE", "language", "language has no compiled Go provider")
	}
	for _, binding := range pack.Bindings {
		if binding.SenseID == senseID && binding.ValencyFrameID == frameID && binding.FamilyID == familyID {
			return Bridge{provider: provider, binding: binding, capabilities: pack.Capabilities}, nil
		}
	}
	return Bridge{}, unknown("UNKNOWN_LANGUAGE_BINDING", "binding", "no binding matches the requested sense, frame, and family")
}

func (bridge Bridge) Analyze(ctx context.Context, surface string) ([]AnalysisCandidate, error) {
	if bridge.provider == nil || !bridge.binding.Analyze || !supportsDirection(bridge.capabilities, true) {
		return nil, unsupported("ANALYZE_NOT_LICENSED", "binding.analyze", "binding and verified provider do not license analysis")
	}
	if ctx == nil {
		return nil, invalid("MISSING_ANALYSIS_CONTEXT", "context", "analysis requires a request context")
	}
	if len(surface) == 0 || len(surface) > MaxSurfaceBytes || !utf8.ValidString(surface) || strings.TrimSpace(surface) == "" {
		return nil, invalid("INVALID_ANALYSIS_SURFACE", "surface", "surface input must be non-empty valid UTF-8 within the package limit")
	}
	analyzer, ok := bridge.provider.(Analyzer)
	if !ok {
		return nil, unsupported("ANALYZER_IMPLEMENTATION_MISSING", "provider", "verified provider does not implement analysis")
	}
	candidates, err := analyzer.Analyze(ctx, surface)
	if err != nil {
		return nil, err
	}
	if len(candidates) > MaxAnalysisCandidates {
		return nil, resource("analysis_candidates", len(candidates), MaxAnalysisCandidates)
	}
	for _, candidate := range candidates {
		if candidate.SenseID != bridge.binding.SenseID || candidate.CoreRef != bridge.binding.CoreRef || candidate.ValencyFrameID != bridge.binding.ValencyFrameID || candidate.FamilyID != bridge.binding.FamilyID || !containsVariant(bridge.binding.VariantIDs, candidate.VariantID) {
			return nil, unsupported("ANALYSIS_CANDIDATE_OUTSIDE_BINDING", "analysis.candidates", "provider returned a candidate outside the approved typed binding")
		}
	}
	return append([]AnalysisCandidate(nil), candidates...), nil
}

func (bridge Bridge) Realize(ctx context.Context, graph meaning.Meaning, targetID meaning.NodeID, plan construction.CompiledConstructionPlan) (string, error) {
	if bridge.provider == nil || !bridge.binding.Realize || !supportsDirection(bridge.capabilities, false) {
		return "", unsupported("REALIZE_NOT_LICENSED", "binding.realize", "binding and verified provider do not license realization")
	}
	if ctx == nil {
		return "", invalid("MISSING_REALIZATION_CONTEXT", "context", "realization requires a request context")
	}
	if err := meaning.ValidateStructure(graph, meaning.DefaultBudget()); err != nil {
		return "", err
	}
	var target *meaning.Node
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == targetID {
			target = &graph.Nodes[i]
			break
		}
	}
	if target == nil {
		return "", unknown("UNKNOWN_REALIZATION_TARGET", "target_node_id", "selected Meaning target does not exist")
	}
	if target.Kind != meaning.KindEvent && target.Kind != meaning.KindState {
		return "", unsupported("REALIZATION_TARGET_UNSUPPORTED", "target_node_id", "realization binding requires an event or state target")
	}
	if target.Concept != meaning.ConceptRefFromID(bridge.binding.CoreRef) {
		return "", unsupported("REALIZATION_CORE_MISMATCH", "target_node_id", "selected Meaning concept does not match the approved binding CoreRef")
	}
	if !plan.ValidateIdentity() || plan.Language != string(bridge.provider.Language()) || plan.FamilyID != bridge.binding.FamilyID || plan.VariantID.String() == "" || !containsVariantString(bridge.binding.VariantIDs, plan.VariantID.String()) {
		return "", unsupported("REALIZATION_PLAN_OUTSIDE_BINDING", "plan", "plan family or variant is outside the approved binding")
	}
	licensed := false
	for _, permission := range plan.Permissions {
		if permission.SenseID == bridge.binding.SenseID && string(permission.FrameID) == bridge.binding.ValencyFrameID {
			licensed = true
			break
		}
	}
	if !licensed {
		return "", unsupported("REALIZATION_FRAME_OUTSIDE_BINDING", "plan.sense_permissions", "plan does not carry this exact sense/frame permission")
	}
	realizer, ok := bridge.provider.(Realizer)
	if !ok {
		return "", unsupported("REALIZER_IMPLEMENTATION_MISSING", "provider", "verified provider does not implement realization")
	}
	form, err := realizer.Realize(ctx, graph, targetID, plan)
	if err != nil {
		return "", err
	}
	if len(form) > MaxRealizedSurfaceBytes {
		return "", resource("realized_surface_bytes", len(form), MaxRealizedSurfaceBytes)
	}
	if form == "" {
		return "", incomplete("EMPTY_REALIZED_SURFACE", "realized_surface", "provider returned no realized surface")
	}
	if !utf8.ValidString(form) {
		return "", unsupported("INVALID_REALIZED_UTF8", "realized_surface", "provider returned invalid UTF-8")
	}
	return form, nil
}

const MaxSurfaceBytes = 16 << 10

func containsVariant(values []ids.ConstructionID, candidate ids.ConstructionID) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
}

func containsVariantString(values []ids.ConstructionID, candidate string) bool {
	for _, value := range values {
		if value.String() == candidate {
			return true
		}
	}
	return false
}
