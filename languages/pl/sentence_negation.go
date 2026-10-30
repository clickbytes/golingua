package pl

import (
	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/lexicon"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/realizer"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/text"
	"github.com/clickbytes/golingua/internal/valency"
	"strings"
)

type polishPresentNegationRoute struct {
	concept          string
	sense            string
	lexeme           string
	family           string
	variant          string
	subjectRole      meaning.Role
	objectRole       meaning.Role
	frame            string
	derived          bool
	personal         bool
	variantsByFamily map[string]map[string]bool
	roles            []meaning.Role
	slotsByFamily    map[string]map[string]construction.SlotCategory
	kind             meaning.NodeKind
}

var polishPresentNegationRoutes = []polishPresentNegationRoute{
	{concept: "concept:buying", sense: "sense:pl.kupowac.buy", lexeme: "lexeme:pl.kupowac.v", family: "construction:pl.transitive_negative", variant: "construction:pl.transitive_negative.svo", subjectRole: "buyer", kind: meaning.KindEvent},
	{concept: "concept:buying", sense: "sense:pl.kupic.buy", lexeme: "lexeme:pl.kupic.v", family: "construction:pl.transitive_negative", variant: "construction:pl.transitive_negative.svo", subjectRole: "buyer", kind: meaning.KindEvent},
	{concept: "concept:giving", sense: "sense:pl.dawac.transfer", lexeme: "lexeme:pl.dawac.v", family: "construction:pl.ditransitive_negative", variant: "construction:pl.ditransitive_negative.svo", subjectRole: "giver", kind: meaning.KindEvent},
	{concept: "concept:giving", sense: "sense:pl.dac.transfer", lexeme: "lexeme:pl.dac.v", family: "construction:pl.ditransitive_negative", variant: "construction:pl.ditransitive_negative.svo", subjectRole: "giver", kind: meaning.KindEvent},
	{concept: "concept:looking", sense: "sense:pl.patrzec.directed", lexeme: "lexeme:pl.patrzec.v", family: "construction:pl.oblique_negative", variant: "construction:pl.oblique_negative.svo", subjectRole: "agent", kind: meaning.KindEvent},
	{concept: "concept:class_membership", sense: "sense:pl.byc.class_membership", lexeme: "lexeme:pl.byc.v", family: "construction:pl.identification_negative", variant: "construction:pl.identification_negative.to", subjectRole: "member", kind: meaning.KindState},
}

// polishNegationRouteResolver derives intransitive and transitive-negative
// routes from exact compiled variants, approved valency bindings, and lexical
// transfer bindings supplied to the source/target pipeline builders.
type polishNegationRouteResolver struct {
	intransitive []polishPresentNegationRoute
	transitive   []polishPresentNegationRoute
}

func newPolishNegationRouteResolver(lexiconIndex *lexicon.Index, valencyIndex *valency.Index, registry *construction.Registry, packages []construction.CompiledConstructionPackage) polishNegationRouteResolver {
	resolver := polishNegationRouteResolver{}
	if lexiconIndex == nil || valencyIndex == nil || registry == nil {
		return resolver
	}
	familyID, err := ids.ParseConstructionID("construction:pl.intransitive")
	if err != nil {
		return resolver
	}
	family, exists := registry.ByID(familyID)
	if !exists {
		return resolver
	}
	variants := make(map[string]construction.Variant)
	for _, pkg := range packages {
		if pkg.Package.Family.ID != familyID {
			continue
		}
		for _, variant := range pkg.Package.Variants {
			variants[variant.ID.String()] = variant
		}
	}
	frames := make(map[string]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[string(frame.ID)] = frame
	}
	for _, binding := range registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.FamilyID != familyID {
			continue
		}
		variant, variantExists := variants[binding.VariantID.String()]
		if !variantExists || !polishVariantHasRequiredNegationPlan(family.Family, variant) {
			continue
		}
		frame, frameExists := frames[string(binding.ValencyFrameID)]
		if !frameExists || frame.SenseID != binding.SenseID || !polishFrameAllowsFamily(frame, familyID) {
			continue
		}
		var subjectRole meaning.Role
		for _, argument := range frame.Arguments {
			if argument.ID == "subject" && argument.Required && argument.GrammaticalFunction == "subject" && argument.Case == "nominative" {
				subjectRole = argument.Role
				break
			}
		}
		if subjectRole == "" {
			continue
		}
		sense, senseExists := lexiconIndex.BySenseID(binding.SenseID)
		if !senseExists {
			continue
		}
		lexeme, lexemeExists := lexiconIndex.ByLexemeID(sense.LexemeID)
		if !lexemeExists || lexeme.Language != "pl" || lexeme.POS != "verb" {
			continue
		}
		concept := ""
		ambiguousConcept := false
		for _, semanticBinding := range lexiconIndex.BindingsForSense(binding.SenseID) {
			if !semanticBinding.TransferAuthorized() {
				continue
			}
			if concept != "" && concept != semanticBinding.CoreRef.String() {
				ambiguousConcept = true
				break
			}
			concept = semanticBinding.CoreRef.String()
		}
		if concept == "" || ambiguousConcept {
			continue
		}
		resolver.intransitive = append(resolver.intransitive, polishPresentNegationRoute{
			concept: concept, sense: binding.SenseID.String(), lexeme: sense.LexemeID.String(),
			family: familyID.String(), variant: binding.VariantID.String(), subjectRole: subjectRole, kind: meaning.KindEvent,
		})
	}
	resolver.transitive = append(
		derivePolishTransitiveNegativeRoutes(lexiconIndex, valencyIndex, registry, packages),
		derivePolishPersonalNegativeRoutes(lexiconIndex, valencyIndex, registry, packages)...,
	)
	return resolver
}

func derivePolishPersonalNegativeRoutes(lexiconIndex *lexicon.Index, valencyIndex *valency.Index, registry *construction.Registry, packages []construction.CompiledConstructionPackage) []polishPresentNegationRoute {
	if lexiconIndex == nil || valencyIndex == nil || registry == nil {
		return nil
	}
	familyID, err := ids.ParseConstructionID("construction:pl.transitive_personal")
	if err != nil {
		return nil
	}
	record, exists := registry.ByID(familyID)
	if !exists || string(record.Family.Language) != "pl" || !polishFamilyHasNegation(record.Family) {
		return nil
	}
	variants := make(map[string]construction.Variant)
	for _, pkg := range packages {
		if pkg.Package.Family.ID != familyID {
			continue
		}
		for _, variant := range pkg.Package.Variants {
			variants[variant.ID.String()] = variant
		}
	}
	frames := make(map[string]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[string(frame.ID)] = frame
	}
	var routes []polishPresentNegationRoute
	for _, binding := range registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.FamilyID != familyID {
			continue
		}
		frame, found := frames[string(binding.ValencyFrameID)]
		if !found || frame.SenseID != binding.SenseID || !polishFrameAllowsFamily(frame, familyID) {
			continue
		}
		subjectRole, objectRole, slots, frameOK := polishPersonalNegativeFrame(record.Family, frame)
		if !frameOK {
			continue
		}
		variant, found := variants[binding.VariantID.String()]
		if !found || variant.Validate(record.Family) != nil || variant.DirectionPolicy == construction.DirectionRealize ||
			!polishVariantHasRequiredNegationPlan(record.Family, variant) ||
			!polishSequenceHasSlot(variant.Sequence, "verb") || !polishSequenceHasSlot(variant.Sequence, "object") {
			continue
		}
		sense, found := lexiconIndex.BySenseID(binding.SenseID)
		if !found {
			continue
		}
		lexeme, found := lexiconIndex.ByLexemeID(sense.LexemeID)
		if !found || lexeme.Language != "pl" || lexeme.POS != "verb" {
			continue
		}
		var concept string
		for _, transfer := range lexiconIndex.BindingsForSense(binding.SenseID) {
			if !transfer.TransferAuthorized() {
				continue
			}
			if concept != "" && concept != transfer.CoreRef.String() {
				concept = ""
				break
			}
			concept = transfer.CoreRef.String()
		}
		if concept == "" {
			continue
		}
		routes = append(routes, polishPresentNegationRoute{
			concept: concept, sense: binding.SenseID.String(), lexeme: sense.LexemeID.String(),
			family: familyID.String(), variant: binding.VariantID.String(), subjectRole: subjectRole,
			objectRole: objectRole, frame: string(binding.ValencyFrameID), derived: true, personal: true,
			variantsByFamily: map[string]map[string]bool{familyID.String(): {binding.VariantID.String(): true}},
			roles:            []meaning.Role{subjectRole, objectRole}, slotsByFamily: map[string]map[string]construction.SlotCategory{familyID.String(): slots},
			kind: meaning.KindState,
		})
	}
	return mergePolishDerivedTransitiveRoutes(routes)
}

func polishPersonalNegativeFrame(family construction.ConstructionFamily, frame valency.ValencyFrame) (meaning.Role, meaning.Role, map[string]construction.SlotCategory, bool) {
	slots := make(map[string]construction.SlotCategory, len(family.SlotSpecs))
	for _, spec := range family.SlotSpecs {
		slots[spec.ID] = spec.Category
		if spec.ID != "subject" && spec.ID != "verb" && spec.ID != "object" {
			return "", "", nil, false
		}
		if spec.ID == "subject" {
			if spec.Category != construction.CategoryNP || spec.Optionality != construction.SlotOptional || spec.Cardinality.Maximum != 1 {
				return "", "", nil, false
			}
			continue
		}
		if spec.Optionality != construction.SlotRequired || spec.Cardinality.Minimum != 1 || spec.Cardinality.Maximum != 1 {
			return "", "", nil, false
		}
	}
	if len(slots) != 3 || slots["verb"] != construction.CategoryVP || slots["object"] != construction.CategoryNP ||
		len(frame.SemanticRoles) != 2 || len(frame.Arguments) != 2 {
		return "", "", nil, false
	}
	var subjectRole, objectRole meaning.Role
	for _, argument := range frame.Arguments {
		if !argument.Required || argument.Role == "" {
			return "", "", nil, false
		}
		switch argument.ID {
		case "subject":
			if argument.GrammaticalFunction != "subject" || argument.Case != "nominative" || subjectRole != "" {
				return "", "", nil, false
			}
			subjectRole = argument.Role
		case "object":
			if argument.GrammaticalFunction != "object" || argument.Case != "genitive" || objectRole != "" {
				return "", "", nil, false
			}
			objectRole = argument.Role
		default:
			return "", "", nil, false
		}
	}
	roles := map[meaning.Role]bool{frame.SemanticRoles[0]: true, frame.SemanticRoles[1]: true}
	if subjectRole == "" || objectRole == "" || subjectRole == objectRole || len(roles) != 2 || !roles[subjectRole] || !roles[objectRole] {
		return "", "", nil, false
	}
	return subjectRole, objectRole, map[string]construction.SlotCategory{"verb": construction.CategoryVP, "object": construction.CategoryNP}, true
}

func derivePolishTransitiveNegativeRoutes(lexiconIndex *lexicon.Index, valencyIndex *valency.Index, registry *construction.Registry, packages []construction.CompiledConstructionPackage) []polishPresentNegationRoute {
	if lexiconIndex == nil || valencyIndex == nil || registry == nil {
		return nil
	}
	familyID, err := ids.ParseConstructionID("construction:pl.transitive_negative")
	if err != nil {
		return nil
	}
	record, exists := registry.ByID(familyID)
	if !exists || string(record.Family.Language) != "pl" || !polishFamilyHasNegation(record.Family) || !polishFamilyHasHardNumberAgreement(record.Family) {
		return nil
	}
	variants := make(map[string]construction.Variant)
	for _, pkg := range packages {
		if pkg.Package.Family.ID != familyID {
			continue
		}
		for _, variant := range pkg.Package.Variants {
			variants[variant.ID.String()] = variant
		}
	}
	frames := make(map[string]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[string(frame.ID)] = frame
	}
	var routes []polishPresentNegationRoute
	for _, binding := range registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.FamilyID != familyID {
			continue
		}
		if polishStaticTransitiveException(binding.SenseID) {
			continue
		}
		frame, found := frames[string(binding.ValencyFrameID)]
		if !found || frame.SenseID != binding.SenseID || !polishFrameAllowsFamily(frame, familyID) {
			continue
		}
		roles, subjectRole, objectRole, slotCategories, valid := polishTransitiveNegativeFrame(record.Family, frame)
		if !valid {
			continue
		}
		variant, found := variants[binding.VariantID.String()]
		if !found || variant.Validate(record.Family) != nil || variant.DirectionPolicy == construction.DirectionAnalyze ||
			!polishVariantHasRequiredNegationPlan(record.Family, variant) {
			continue
		}
		for _, slot := range []string{"subject", "verb", "object"} {
			if !polishSequenceHasSlot(variant.Sequence, slot) {
				valid = false
			}
		}
		if !valid {
			continue
		}
		sense, found := lexiconIndex.BySenseID(binding.SenseID)
		if !found {
			continue
		}
		lexeme, found := lexiconIndex.ByLexemeID(sense.LexemeID)
		if !found || lexeme.Language != "pl" || lexeme.POS != "verb" {
			continue
		}
		bindings := lexiconIndex.BindingsForSense(binding.SenseID)
		concept := ""
		for _, transfer := range bindings {
			if !transfer.TransferAuthorized() {
				continue
			}
			if concept != "" && concept != transfer.CoreRef.String() {
				concept = ""
				break
			}
			concept = transfer.CoreRef.String()
		}
		if concept == "" {
			continue
		}
		routes = append(routes, polishPresentNegationRoute{
			concept: concept, sense: binding.SenseID.String(), lexeme: sense.LexemeID.String(),
			family: familyID.String(), variant: binding.VariantID.String(), subjectRole: subjectRole,
			objectRole: objectRole, frame: string(binding.ValencyFrameID), derived: true,
			variantsByFamily: map[string]map[string]bool{familyID.String(): {binding.VariantID.String(): true}}, roles: roles,
			slotsByFamily: map[string]map[string]construction.SlotCategory{familyID.String(): slotCategories}, kind: meaning.KindEvent,
		})
	}
	routes = mergePolishDerivedTransitiveRoutes(routes)
	return addPolishAttributedSeeingNegativeVariants(routes, valencyIndex, registry, packages)
}

func polishFamilyHasNegation(family construction.ConstructionFamily) bool {
	for _, operator := range family.Operators {
		if operator == "negate" {
			for _, capability := range family.Capabilities {
				if capability == "negation-negated" {
					return true
				}
			}
		}
	}
	return false
}

func polishFamilyHasHardNumberAgreement(family construction.ConstructionFamily) bool {
	for _, constraint := range family.Constraints {
		if constraint.Kind != construction.ConstraintAgreement || constraint.Strength != construction.ConstraintHard ||
			constraint.ControllerSlot != "subject" || constraint.TargetSlot != "verb" {
			continue
		}
		for _, feature := range constraint.Features {
			if feature == "number" {
				return true
			}
		}
	}
	return false
}

func polishStaticTransitiveException(sense ids.SenseID) bool {
	for _, route := range polishPresentNegationRoutes {
		if route.sense == sense.String() && route.family == "construction:pl.transitive_negative" && route.concept == "concept:buying" {
			return true
		}
	}
	return false
}

func polishTransitiveNegativeFrame(family construction.ConstructionFamily, frame valency.ValencyFrame) ([]meaning.Role, meaning.Role, meaning.Role, map[string]construction.SlotCategory, bool) {
	roles := make([]meaning.Role, 0, len(family.SlotSpecs))
	slots := make(map[string]construction.SlotCategory, len(family.SlotSpecs))
	var subjectRole, objectRole meaning.Role
	for _, spec := range family.SlotSpecs {
		if spec.Optionality != construction.SlotRequired || spec.Cardinality.Minimum == 0 {
			return nil, "", "", nil, false
		}
		slots[spec.ID] = spec.Category
	}
	if slots["subject"] != construction.CategoryNP || slots["verb"] != construction.CategoryVP || slots["object"] != construction.CategoryNP {
		return nil, "", "", nil, false
	}
	for _, argument := range frame.Arguments {
		if argument.ID != "subject" && argument.ID != "object" {
			if argument.Required {
				return nil, "", "", nil, false
			}
			continue
		}
		if !argument.Required || argument.Role == "" {
			return nil, "", "", nil, false
		}
		if argument.ID == "subject" {
			if argument.GrammaticalFunction != "subject" || argument.Case != "nominative" || subjectRole != "" {
				return nil, "", "", nil, false
			}
			subjectRole = argument.Role
		} else {
			if argument.GrammaticalFunction != "object" || argument.Case != "genitive" || objectRole != "" {
				return nil, "", "", nil, false
			}
			objectRole = argument.Role
		}
		roles = append(roles, argument.Role)
	}
	return roles, subjectRole, objectRole, slots, subjectRole != "" && objectRole != "" && len(roles) == 2 && len(slots) == 3
}

func polishSequenceHasSlot(sequence []construction.SequenceToken, slot string) bool {
	for _, token := range sequence {
		if token.Kind == construction.TokenSlot && token.SlotID == slot {
			return true
		}
	}
	return false
}

func mergePolishDerivedTransitiveRoutes(routes []polishPresentNegationRoute) []polishPresentNegationRoute {
	merged := make([]polishPresentNegationRoute, 0, len(routes))
	positions := make(map[string]int)
	for _, route := range routes {
		key := route.sense + "\x00" + route.frame
		position, exists := positions[key]
		if !exists {
			positions[key] = len(merged)
			merged = append(merged, route)
			continue
		}
		current := &merged[position]
		for family, variants := range route.variantsByFamily {
			if current.variantsByFamily[family] == nil {
				current.variantsByFamily[family] = make(map[string]bool)
			}
			for variant := range variants {
				current.variantsByFamily[family][variant] = true
			}
		}
	}
	return merged
}

// addPolishAttributedSeeingNegativeVariants retains the existing AP-object
// realization as its own approved family. Its binding is admitted only when
// its frame has the same nominative subject and genitive object roles as the
// derived transitive route and its compiled variants contain every slot.
func addPolishAttributedSeeingNegativeVariants(routes []polishPresentNegationRoute, valencyIndex *valency.Index, registry *construction.Registry, packages []construction.CompiledConstructionPackage) []polishPresentNegationRoute {
	if valencyIndex == nil || registry == nil {
		return routes
	}
	familyID, err := ids.ParseConstructionID("construction:pl.attributed_seeing_object_negative")
	if err != nil {
		return routes
	}
	record, exists := registry.ByID(familyID)
	if !exists || !polishFamilyHasNegation(record.Family) {
		return routes
	}
	variants := make(map[string]construction.Variant)
	for _, pkg := range packages {
		if pkg.Package.Family.ID != familyID {
			continue
		}
		for _, variant := range pkg.Package.Variants {
			variants[variant.ID.String()] = variant
		}
	}
	frames := make(map[string]valency.ValencyFrame)
	for _, frame := range valencyIndex.Frames() {
		frames[string(frame.ID)] = frame
	}
	for _, binding := range registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.FamilyID != familyID {
			continue
		}
		frame, found := frames[string(binding.ValencyFrameID)]
		if !found || frame.SenseID != binding.SenseID || !polishFrameAllowsFamily(frame, familyID) {
			continue
		}
		var route *polishPresentNegationRoute
		for index := range routes {
			if routes[index].sense == binding.SenseID.String() {
				if route != nil {
					route = nil
					break
				}
				route = &routes[index]
			}
		}
		if route == nil || !polishAttributedFrameMatchesRoute(frame, route) || !polishAttributedFamilyMatchesRoute(record.Family, route) {
			continue
		}
		variant, found := variants[binding.VariantID.String()]
		if !found || variant.Validate(record.Family) != nil || variant.DirectionPolicy == construction.DirectionAnalyze ||
			!polishVariantHasRequiredNegationPlan(record.Family, variant) {
			continue
		}
		categories := make(map[string]construction.SlotCategory, len(record.Family.SlotSpecs))
		for _, slot := range record.Family.SlotSpecs {
			categories[slot.ID] = slot.Category
		}
		if route.variantsByFamily[familyID.String()] == nil {
			route.variantsByFamily[familyID.String()] = make(map[string]bool)
			route.slotsByFamily[familyID.String()] = categories
		}
		route.variantsByFamily[familyID.String()][binding.VariantID.String()] = true
	}
	return routes
}

func polishAttributedFrameMatchesRoute(frame valency.ValencyFrame, route *polishPresentNegationRoute) bool {
	var subject, object valency.ArgumentSlot
	for _, argument := range frame.Arguments {
		if argument.ID == "subject" {
			subject = argument
		} else if argument.ID == "object" {
			object = argument
		} else if argument.Required {
			return false
		}
	}
	return subject.Required && subject.GrammaticalFunction == "subject" && subject.Case == "nominative" && subject.Role == route.subjectRole &&
		object.Required && object.GrammaticalFunction == "object" && object.Case == "genitive" && object.Role == route.objectRole
}

func polishAttributedFamilyMatchesRoute(family construction.ConstructionFamily, route *polishPresentNegationRoute) bool {
	var subject, object, verb bool
	var modifier bool
	for _, slot := range family.SlotSpecs {
		if slot.Optionality != construction.SlotRequired || slot.Cardinality.Minimum == 0 {
			continue
		}
		switch slot.ID {
		case "subject":
			subject = slot.Category == construction.CategoryNP && slot.ArgumentBinding != nil && slot.ArgumentBinding.SemanticRole == route.subjectRole && slot.Case == "nominative"
		case "object":
			object = slot.Category == construction.CategoryNP && slot.ArgumentBinding != nil && slot.ArgumentBinding.SemanticRole == route.objectRole && slot.Case == "genitive"
		case "verb":
			verb = slot.Category == construction.CategoryVP && slot.IsMainPredicate()
		default:
			if slot.Category == construction.CategoryAP && slot.ModifierBinding != nil && slot.ModifierBinding.TargetSlot == "object" {
				modifier = true
			}
		}
	}
	return subject && object && verb && modifier && polishFamilyHasHardNumberAgreement(family)
}

func polishVariantHasRequiredNegationPlan(family construction.ConstructionFamily, variant construction.Variant) bool {
	if variant.FamilyID != family.ID || variant.Validate(family) != nil {
		return false
	}
	requiredSlots := make(map[string]bool)
	for _, spec := range family.SlotSpecs {
		if spec.Optionality == construction.SlotRequired && spec.Cardinality.Minimum > 0 {
			requiredSlots[spec.ID] = false
		}
	}
	negationTerminal := false
	for _, token := range variant.Sequence {
		if token.Kind == construction.TokenSlot {
			if _, required := requiredSlots[token.SlotID]; required {
				requiredSlots[token.SlotID] = true
			}
		}
		if token.Kind == construction.TokenTerminal && token.Terminal == "nie" {
			negationTerminal = true
		}
	}
	if !negationTerminal {
		return false
	}
	for _, present := range requiredSlots {
		if !present {
			return false
		}
	}
	return true
}

func polishFrameAllowsFamily(frame valency.ValencyFrame, familyID ids.ConstructionID) bool {
	for _, allowed := range frame.AllowedFamilyIDs {
		if allowed == familyID {
			return true
		}
	}
	return false
}

func (resolver polishNegationRouteResolver) routeForSense(sense ids.SenseID) (polishPresentNegationRoute, bool) {
	if route, found := polishPresentNegationRouteForSense(sense); found {
		return route, true
	}
	var selected polishPresentNegationRoute
	for _, route := range resolver.transitive {
		if route.sense != sense.String() {
			continue
		}
		if selected.sense != "" && (selected.frame != route.frame || selected.subjectRole != route.subjectRole || selected.objectRole != route.objectRole) {
			return polishPresentNegationRoute{}, false
		}
		if selected.sense == "" {
			selected = route
		} else {
			for family, variants := range route.variantsByFamily {
				if selected.variantsByFamily[family] == nil {
					selected.variantsByFamily[family] = make(map[string]bool)
				}
				for variant := range variants {
					selected.variantsByFamily[family][variant] = true
				}
			}
		}
	}
	if selected.sense != "" {
		return selected, true
	}
	for _, route := range resolver.intransitive {
		if route.sense == sense.String() {
			return route, true
		}
	}
	return polishPresentNegationRoute{}, false
}

func (resolver polishNegationRouteResolver) routeForEvent(event meaning.Node) (polishPresentNegationRoute, bool) {
	staticRoute, staticFound := polishPresentNegationRouteForEvent(event)
	var selected polishPresentNegationRoute
	if staticFound {
		selected = staticRoute
	}
	for _, route := range resolver.transitive {
		if route.concept != string(event.Concept) || event.Kind != route.kind || !polishEventHasRoles(event, route.roles) {
			continue
		}
		if selected.sense != "" && selected.sense != route.sense {
			return polishPresentNegationRoute{}, false
		}
		selected = route
	}
	if selected.sense != "" {
		return selected, true
	}
	for _, route := range resolver.intransitive {
		if route.concept != string(event.Concept) || event.Kind != route.kind {
			continue
		}
		if len(event.Participants) != 1 || event.Participants[0].Role != route.subjectRole {
			continue
		}
		if selected.sense != "" && selected.sense != route.sense {
			return polishPresentNegationRoute{}, false
		}
		selected = route
	}
	if selected.sense == "" {
		return polishPresentNegationRoute{}, false
	}
	return selected, true
}

func polishEventHasRoles(event meaning.Node, roles []meaning.Role) bool {
	if len(roles) == 0 || len(event.Participants) != len(roles) {
		return false
	}
	seen := make(map[meaning.Role]bool, len(event.Participants))
	for _, participant := range event.Participants {
		if seen[participant.Role] {
			return false
		}
		seen[participant.Role] = true
	}
	for _, role := range roles {
		if !seen[role] {
			return false
		}
	}
	return true
}

func (resolver polishNegationRouteResolver) negatedPresent(source semantic.MeaningCandidate) (meaning.Node, meaning.Node, polishPresentNegationRoute, bool) {
	if source.Meaning == nil {
		return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
	}
	var scope meaning.Node
	for _, node := range source.Meaning.Nodes {
		if node.Kind != meaning.KindScope {
			continue
		}
		if scope.ID != "" || node.Operator != meaning.OperatorNegation {
			return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
		}
		scope = node
	}
	if scope.ID == "" {
		return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID != scope.Target {
			continue
		}
		route, found := resolver.routeForEvent(node)
		if found {
			return scope, node, route, true
		}
	}
	return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
}

func polishPresentNegationRouteForSense(sense ids.SenseID) (polishPresentNegationRoute, bool) {
	for _, route := range polishPresentNegationRoutes {
		if route.sense == sense.String() {
			return route, true
		}
	}
	for _, spec := range polishAttributionSpecs {
		if spec.sense == sense.String() {
			return polishPresentNegationRoute{concept: spec.concept, sense: spec.sense, lexeme: spec.lexeme, family: "construction:pl.attribution_negative", subjectRole: "bearer", kind: meaning.KindState}, true
		}
	}
	return polishPresentNegationRoute{}, false
}

func polishGivingNegationSenseMatches(route polishPresentNegationRoute, sense string) bool {
	return route.concept == "concept:giving" && (sense == "sense:pl.dac.transfer" || sense == "sense:pl.dawac.transfer")
}

func polishBuyingNegationSenseMatches(route polishPresentNegationRoute, sense string) bool {
	return route.concept == "concept:buying" && (sense == "sense:pl.kupic.buy" || sense == "sense:pl.kupowac.buy")
}

func polishPresentNegationRouteForEvent(event meaning.Node) (polishPresentNegationRoute, bool) {
	for _, route := range polishPresentNegationRoutes {
		if route.concept == string(event.Concept) && event.Kind == route.kind {
			return route, true
		}
	}
	if event.Kind == meaning.KindState {
		for _, spec := range polishAttributionSpecs {
			if spec.concept == string(event.Concept) {
				return polishPresentNegationRoute{concept: spec.concept, sense: spec.sense, lexeme: spec.lexeme, family: "construction:pl.attribution_negative", subjectRole: "bearer", kind: meaning.KindState}, true
			}
		}
	}
	return polishPresentNegationRoute{}, false
}

func polishPresentNegationRouteForFamily(family ids.ConstructionID) (polishPresentNegationRoute, bool) {
	for _, route := range polishPresentNegationRoutes {
		if route.family == family.String() {
			return route, true
		}
	}
	if family.String() == "construction:pl.attribution_negative" && len(polishAttributionSpecs) > 0 {
		spec := polishAttributionSpecs[0]
		return polishPresentNegationRoute{concept: spec.concept, sense: spec.sense, lexeme: spec.lexeme, family: family.String(), subjectRole: "bearer", kind: meaning.KindState}, true
	}
	return polishPresentNegationRoute{}, false
}

func polishPresentNegationVariantAllowed(route polishPresentNegationRoute, family, variant string) bool {
	if route.derived {
		return route.variantsByFamily[family][variant]
	}
	if route.kind == meaning.KindState && route.family == "construction:pl.attribution_negative" {
		return family == route.family && strings.HasPrefix(variant, route.family+".")
	}
	if route.concept == "concept:seeing" && family == "construction:pl.attributed_seeing_object_negative" {
		return variant == "construction:pl.attributed_seeing_object_negative.svo" || variant == "construction:pl.attributed_seeing_object_negative.ovs"
	}
	if family != route.family {
		return false
	}
	if variant == route.variant {
		return true
	}
	if route.concept == "concept:giving" && route.family == "construction:pl.ditransitive_negative" {
		return variant == route.variant+".period"
	}
	return (route.concept == "concept:seeing" || route.concept == "concept:hearing" || route.concept == "concept:reading") && route.family == "construction:pl.transitive_negative" &&
		variant == "construction:pl.transitive_negative.ovs"
}

func polishNegatedPresent(source semantic.MeaningCandidate) (meaning.Node, meaning.Node, polishPresentNegationRoute, bool) {
	if source.Meaning == nil {
		return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
	}
	var scope meaning.Node
	for _, node := range source.Meaning.Nodes {
		if node.Kind != meaning.KindScope {
			continue
		}
		if scope.ID != "" || node.Operator != meaning.OperatorNegation {
			return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
		}
		scope = node
	}
	if scope.ID == "" {
		return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID != scope.Target {
			continue
		}
		route, found := polishPresentNegationRouteForEvent(node)
		if found {
			return scope, node, route, true
		}
	}
	return meaning.Node{}, meaning.Node{}, polishPresentNegationRoute{}, false
}

func polishNegatedPresentCandidate(candidate parser.PhraseCandidate, route polishPresentNegationRoute) bool {
	if route.kind == meaning.KindState && route.family == "construction:pl.attribution_negative" {
		return candidate.Category == construction.CategoryAP && candidate.LexemeID.String() == route.lexeme &&
			candidate.Features["degree"] == "positive" && (candidate.Features["number"] == "singular" || candidate.Features["number"] == "plural")
	}
	if candidate.Category != construction.CategoryVP || candidate.LexemeID.String() != route.lexeme || candidate.Features["verb_form"] != "finite" {
		return false
	}
	person := string(candidate.Features["person"])
	if person != "third" && !(route.personal && (person == "first" || person == "second")) {
		return false
	}
	if candidate.Features["tense"] == "past" {
		if route.derived {
			gender := string(candidate.Features["gender"])
			return (candidate.Features["aspect"] == "imperfective" || candidate.Features["aspect"] == "perfective") &&
				(gender == "masculine_personal" || gender == "masculine_animal" || gender == "masculine_inanimate" || gender == "feminine" || gender == "neuter" || gender == "non_masculine_personal")
		}
		if route.family != "construction:pl.intransitive" && !polishPastNegativeConceptAllowed(route.concept) {
			return false
		}
		if route.concept == "concept:giving" {
			expectedAspect := "imperfective"
			if route.sense == "sense:pl.dac.transfer" {
				expectedAspect = "perfective"
			}
			if string(candidate.Features["aspect"]) != expectedAspect {
				return false
			}
		} else if route.concept == "concept:buying" {
			expectedAspect := "imperfective"
			if route.sense == "sense:pl.kupic.buy" {
				expectedAspect = "perfective"
			}
			if string(candidate.Features["aspect"]) != expectedAspect {
				return false
			}
		} else if candidate.Features["aspect"] != "imperfective" {
			return false
		}
		gender := string(candidate.Features["gender"])
		return gender == "masculine_personal" || gender == "feminine" || gender == "neuter" || gender == "non_masculine_personal"
	}
	if candidate.Features["tense"] != "present" {
		return false
	}
	if route.derived && candidate.Features["aspect"] != "imperfective" {
		return false
	}
	return candidate.Features["number"] == "singular" || candidate.Features["number"] == "plural"
}

func polishNegationNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, route polishPresentNegationRoute, candidates []realizer.TargetCandidate, routes polishNegationRouteResolver) (string, bool) {
	_, event, sourceRoute, found := routes.negatedPresent(source)
	if !found || event.ID != predicate.NodeID || sourceRoute.family != route.family || sourceRoute.sense != route.sense && !(route.concept == "concept:giving" && polishGivingNegationSenseMatches(route, sourceRoute.sense)) && !(route.concept == "concept:buying" && polishBuyingNegationSenseMatches(route, sourceRoute.sense)) {
		return "", false
	}
	referent, exists := polishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || route.concept == "concept:class_membership" && referent.Multiplicity == meaning.MultiplicityMany {
		return "", false
	}
	if !polishFrameEvidenceConsistent(source, event.ID, predicate.TransferBinding.CoreRef) ||
		!polishMultiplicityEvidenceConsistent(source, event.ID, route.subjectRole, referent, predicate.TransferBinding.CoreRef, true) {
		return "", false
	}
	return polishNumberForReferentCandidates(source, referent.ID, route.subjectRole, event.ID, predicate.TransferBinding.CoreRef, candidates)
}

func polishTerminalSpan(lattice analyzer.LexicalLattice, terminal string) (text.Span, bool) {
	var span text.Span
	found := false
	for _, token := range lattice.Tokens {
		if token.Token.Folded != terminal {
			continue
		}
		if found {
			return text.Span{}, false
		}
		span, found = token.Token.Span, true
	}
	return span, found
}
