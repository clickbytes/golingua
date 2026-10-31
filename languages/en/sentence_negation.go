package en

import (
	"strings"

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
)

type englishPresentNegationRoute struct {
	concept          string
	sense            string
	verb             string
	lexeme           string
	family           string
	frame            string
	singularVariant  string
	pluralVariant    string
	subjectRole      meaning.Role
	objectRole       meaning.Role
	requiresAt       bool
	statePredicate   bool
	variantsByNumber map[string]map[string]bool
	pastVariants     map[string]bool
}

// englishNegationResolver carries the immutable compiled language data needed
// to derive predicate routes from approved bindings and their family plans.
type englishNegationResolver struct {
	registry *construction.Registry
	valency  *valency.Index
	lexicon  *lexicon.Index
	variants map[ids.ConstructionID]*construction.VariantIndex
	// Compiled once for this exact immutable index set. Never modified by a request.
	compiledRoutes []englishPresentNegationRoute
}

var englishIdentificationNegationRoute = englishPresentNegationRoute{
	concept: "concept:class_membership", sense: "sense:en.be.class_membership", verb: "is",
	lexeme: "lexeme:en.be.v", family: "construction:en.identification_negative",
	singularVariant: "construction:en.identification_negative.this", subjectRole: "member", statePredicate: true,
}

var englishPresentNegationRoutes = []englishPresentNegationRoute{
	{concept: "concept:giving", sense: "sense:en.give.transfer", verb: "give", lexeme: "lexeme:en.give.v", family: "construction:en.ditransitive_negative", frame: "en.give.transfer.negative", singularVariant: "construction:en.ditransitive_negative.present", pluralVariant: "construction:en.ditransitive_negative.present", subjectRole: "giver"},
	{concept: "concept:looking", sense: "sense:en.look.directed", verb: "look", lexeme: "lexeme:en.look.v", family: "construction:en.oblique_negative", singularVariant: "construction:en.oblique_negative.svo", pluralVariant: "construction:en.oblique_negative.svo.plural", subjectRole: "agent", requiresAt: true},
}

func (resolver englishNegationResolver) routeForVerb(verb string) (englishPresentNegationRoute, bool) {
	var matched englishPresentNegationRoute
	found := false
	for _, route := range resolver.routes() {
		if route.verb == verb {
			if found && negationRouteKey(matched) != negationRouteKey(route) {
				return englishPresentNegationRoute{}, false
			}
			matched, found = route, true
		}
	}
	return matched, found
}

func newEnglishNegationResolver(registry *construction.Registry, index *valency.Index, lexicalIndex *lexicon.Index, variants map[ids.ConstructionID]*construction.VariantIndex) englishNegationResolver {
	resolver := englishNegationResolver{registry: registry, valency: index, lexicon: lexicalIndex, variants: variants}
	resolver.compiledRoutes = append([]englishPresentNegationRoute(nil), englishPresentNegationRoutes...)
	resolver.compiledRoutes = append(resolver.compiledRoutes, resolver.approvedNegativeRoutes()...)
	return resolver
}

func (resolver englishNegationResolver) routes() []englishPresentNegationRoute {
	if resolver.compiledRoutes != nil {
		return resolver.compiledRoutes
	}
	// The zero value deliberately exposes only the existing static routes.
	// Production adapters compile their data-derived routes at construction.
	return englishPresentNegationRoutes
}

// approvedNegativeRoutes derives intransitive and transitive negative
// capability from compiled variants and exact approved sense/frame bindings.
func (resolver englishNegationResolver) approvedNegativeRoutes() []englishPresentNegationRoute {
	if resolver.registry == nil || resolver.valency == nil || resolver.lexicon == nil || resolver.variants == nil {
		return nil
	}
	frames := make(map[valency.FrameID]valency.ValencyFrame)
	for _, frame := range resolver.valency.Frames() {
		frames[frame.ID] = frame
	}
	var routes []englishPresentNegationRoute
	for _, record := range resolver.registry.Families() {
		family := record.Family
		familyID := family.ID.String()
		if string(family.Language) != "en" || familyID != "construction:en.intransitive" && familyID != "construction:en.transitive_negative" ||
			!familyHasOperator(family, construction.OperatorID("negate")) || !familyHasCapability(family, construction.Capability("negation-negated")) {
			continue
		}
		var predicate *construction.SlotSpec
		for index := range family.SlotSpecs {
			slot := &family.SlotSpecs[index]
			if slot.Category == construction.CategoryVP && slot.IsMainPredicate() {
				if predicate != nil {
					predicate = nil
					break
				}
				predicate = slot
			}
		}
		if predicate == nil {
			continue
		}
		variantIndex := resolver.variants[family.ID]
		if variantIndex == nil {
			continue
		}
		for _, binding := range resolver.registry.BindingsByFamilyID(family.ID) {
			if !binding.Approved || binding.FamilyID != family.ID || resolver.hasStaticRoute(binding.SenseID, family.ID) {
				continue
			}
			frame, exists := frames[binding.ValencyFrameID]
			if !exists || frame.SenseID != binding.SenseID || !valencyFrameAllowsFamily(frame, family.ID) {
				continue
			}
			var objectRole meaning.Role
			if familyID == "construction:en.transitive_negative" {
				subjectRole, frameObjectRole, valid := englishTransitiveFrameRoles(family, frame)
				if !valid || subjectRole == "" {
					continue
				}
				objectRole = frameObjectRole
			}
			subjectSlot, role, subjectFound := approvedFamilySubjectSlot(family, frame)
			if !subjectFound {
				continue
			}
			sense, exists := resolver.lexicon.BySenseID(binding.SenseID)
			if !exists {
				continue
			}
			lexeme, exists := resolver.lexicon.LookupLexeme(sense.LexemeID)
			if !exists || lexeme.Language != "en" || lexeme.POS != "verb" {
				continue
			}
			coreBindings := resolver.lexicon.TransferBindings(binding.SenseID)
			if len(coreBindings) != 1 || !coreBindings[0].TransferAuthorized() {
				continue
			}
			baseRoute := englishPresentNegationRoute{
				concept: coreBindings[0].CoreRef.String(), sense: binding.SenseID.String(), frame: string(binding.ValencyFrameID),
				verb: strings.ToLower(lexeme.Lemma), lexeme: sense.LexemeID.String(), family: family.ID.String(), subjectRole: role, objectRole: objectRole,
				variantsByNumber: map[string]map[string]bool{"singular": {}, "plural": {}}, pastVariants: map[string]bool{},
			}
			variantID, err := construction.ParseVariantID(binding.VariantID.String())
			if err != nil {
				continue
			}
			compiled, exists := variantIndex.ByID(variantID)
			if !exists {
				continue
			}
			variant := compiled.Variant
			if variant.DirectionPolicy == construction.DirectionAnalyze || !variantHasRequiredSlots(family, variant) ||
				!sequenceHasSlot(variant.Sequence, subjectSlot.ID) || !sequenceHasSlot(variant.Sequence, predicate.ID) {
				continue
			}
			if !sequenceHasTerminal(variant.Sequence, "not") {
				if familyID != "construction:en.transitive_negative" || variant.DirectionPolicy != construction.DirectionRealize ||
					!sequenceHasSlot(variant.Sequence, "object") || sequenceAuxiliary(variant.Sequence) != "" ||
					!resolver.approvedNegativeSourceVariantExists(family.ID, binding.SenseID, binding.ValencyFrameID, subjectSlot.ID, predicate.ID) {
					continue
				}
				routes = append(routes, baseRoute)
				continue
			}
			auxiliary := sequenceAuxiliary(variant.Sequence)
			if auxiliary != "does" && auxiliary != "do" && auxiliary != "did" {
				continue
			}
			conditionNumber, numberConsistent := variantSubjectNumberIfConsistent(variant, subjectSlot.ID)
			if !numberConsistent || !variantSubjectPersonThird(variant, subjectSlot.ID) {
				continue
			}
			if auxiliary == "did" {
				if variantTenseConflicts(variant, predicate.ID, "past") {
					continue
				}
				baseRoute.pastVariants[variantID.String()] = true
			} else {
				if variantTenseConflicts(variant, predicate.ID, "present") {
					continue
				}
				number := "singular"
				if auxiliary == "do" {
					number = "plural"
				}
				if conditionNumber != "" && conditionNumber != number {
					continue
				}
				baseRoute.variantsByNumber[number][variantID.String()] = true
				if number == "singular" {
					baseRoute.singularVariant = variantID.String()
				} else {
					baseRoute.pluralVariant = variantID.String()
				}
			}
			routes = append(routes, baseRoute)
		}
	}
	return mergeDerivedNegativeRoutes(routes)
}

func (resolver englishNegationResolver) approvedNegativeSourceVariantExists(family ids.ConstructionID, sense ids.SenseID, frame valency.FrameID, subjectSlot, predicateSlot string) bool {
	if resolver.registry == nil || resolver.variants == nil {
		return false
	}
	familyRecord, familyExists := resolver.registry.ByID(family)
	variantIndex := resolver.variants[family]
	if !familyExists || variantIndex == nil {
		return false
	}
	for _, binding := range resolver.registry.BindingsByFamilyID(family) {
		if !binding.Approved || binding.SenseID != sense || binding.ValencyFrameID != frame || binding.FamilyID != family {
			continue
		}
		variantID, err := construction.ParseVariantID(binding.VariantID.String())
		if err != nil {
			continue
		}
		compiled, exists := variantIndex.ByID(variantID)
		if !exists || compiled.Variant.DirectionPolicy == construction.DirectionRealize ||
			!variantHasRequiredSlots(familyRecord.Family, compiled.Variant) ||
			!sequenceHasSlot(compiled.Variant.Sequence, subjectSlot) || !sequenceHasSlot(compiled.Variant.Sequence, predicateSlot) ||
			!sequenceHasTerminal(compiled.Variant.Sequence, "not") {
			continue
		}
		return true
	}
	return false
}

func (resolver englishNegationResolver) hasStaticRoute(sense ids.SenseID, family ids.ConstructionID) bool {
	for _, route := range englishPresentNegationRoutes {
		if route.sense == sense.String() && route.family == family.String() {
			return true
		}
	}
	return false
}

func familyHasOperator(family construction.ConstructionFamily, operator construction.OperatorID) bool {
	for _, candidate := range family.Operators {
		if candidate == operator {
			return true
		}
	}
	return false
}

func familyHasCapability(family construction.ConstructionFamily, capability construction.Capability) bool {
	for _, candidate := range family.Capabilities {
		if candidate == capability {
			return true
		}
	}
	return false
}

func valencyFrameAllowsFamily(frame valency.ValencyFrame, family ids.ConstructionID) bool {
	for _, allowed := range frame.AllowedFamilyIDs {
		if allowed == family {
			return true
		}
	}
	return false
}

func approvedFamilySubjectSlot(family construction.ConstructionFamily, frame valency.ValencyFrame) (construction.SlotSpec, meaning.Role, bool) {
	var selected construction.SlotSpec
	var role meaning.Role
	for _, slot := range family.SlotSpecs {
		if slot.Category != construction.CategoryNP || slot.Optionality != construction.SlotRequired || slot.Cardinality.Minimum == 0 {
			continue
		}
		for _, argument := range frame.Arguments {
			if argument.ID != slot.ID || argument.GrammaticalFunction != "subject" || !argument.Required {
				continue
			}
			if role != "" {
				return construction.SlotSpec{}, "", false
			}
			selected, role = slot, argument.Role
		}
	}
	return selected, role, role != ""
}

func variantHasRequiredSlots(family construction.ConstructionFamily, variant construction.Variant) bool {
	if variant.FamilyID != family.ID || variant.Validate(family) != nil {
		return false
	}
	present := make(map[string]bool)
	for _, token := range variant.Sequence {
		if token.Kind == construction.TokenSlot {
			present[token.SlotID] = true
		}
	}
	for _, slot := range family.SlotSpecs {
		if slot.Optionality == construction.SlotRequired && slot.Cardinality.Minimum > 0 && !present[slot.ID] {
			return false
		}
	}
	return true
}

func englishAlternativeHasRequiredSlots(alternative parser.ParseAlternative, family construction.ConstructionFamily, candidates map[string]parser.PhraseCandidate) bool {
	slots := make(map[string]parser.SlotBinding, len(alternative.Slots))
	for _, slot := range alternative.Slots {
		slots[slot.SlotID] = slot
	}
	for _, spec := range family.SlotSpecs {
		slot, exists := slots[spec.ID]
		if !exists || spec.Optionality == construction.SlotRequired && len(slot.CandidateIDs) < spec.Cardinality.Minimum {
			return false
		}
		for _, candidateID := range slot.CandidateIDs {
			candidate, found := candidates[candidateID]
			if !found || candidate.Category != spec.Category {
				return false
			}
		}
	}
	return true
}

func englishGivingNegativeFrameCoverage(alternative parser.ParseAlternative, family construction.ConstructionFamily, index *valency.Index, candidates map[string]parser.PhraseCandidate) bool {
	if index == nil || alternative.FamilyID.String() != "construction:en.ditransitive_negative" ||
		alternative.Permission.SenseID.String() != "sense:en.give.transfer" || string(alternative.Permission.FrameID) != "en.give.transfer.negative" {
		return false
	}
	var frame *valency.ValencyFrame
	for _, item := range index.FramesBySenseID(alternative.Permission.SenseID) {
		if string(item.ID) == string(alternative.Permission.FrameID) {
			copy := item
			frame = &copy
			break
		}
	}
	if frame == nil || !valencyFrameAllowsFamily(*frame, alternative.FamilyID) || len(frame.SemanticRoles) != 3 || len(frame.Arguments) != 3 {
		return false
	}
	wantSlots := map[string]string{"subject": "giver", "recipient": "recipient", "theme": "theme"}
	wantRoles := map[string]bool{"giver": true, "recipient": true, "theme": true}
	roles := make(map[string]bool, len(wantRoles))
	for _, role := range frame.SemanticRoles {
		if !wantRoles[string(role)] || roles[string(role)] {
			return false
		}
		roles[string(role)] = true
	}
	if len(roles) != len(wantRoles) {
		return false
	}
	for _, argument := range frame.Arguments {
		role, exists := wantSlots[argument.ID]
		if !exists || string(argument.Role) != role || !argument.Required {
			return false
		}
	}
	seen := make(map[string]bool, len(wantSlots))
	for _, slot := range alternative.Slots {
		if slot.SlotID == "verb" {
			if slot.Role != "" || len(slot.CandidateIDs) == 0 {
				return false
			}
			for _, candidateID := range slot.CandidateIDs {
				candidate, exists := candidates[candidateID]
				if !exists || candidate.Category != construction.CategoryVP {
					return false
				}
			}
			continue
		}
		role, expected := wantSlots[slot.SlotID]
		if !expected || string(slot.Role) != role || seen[slot.SlotID] || len(slot.CandidateIDs) == 0 {
			return false
		}
		for _, candidateID := range slot.CandidateIDs {
			candidate, exists := candidates[candidateID]
			if !exists || candidate.Category != construction.CategoryNP {
				return false
			}
		}
		seen[slot.SlotID] = true
	}
	return len(seen) == len(wantSlots)
}

func variantSubjectNumberIfConsistent(variant construction.Variant, subjectSlot string) (string, bool) {
	number := ""
	for _, condition := range variant.Conditions {
		if condition.Kind != construction.ConditionFeatureEquals || condition.SlotID != subjectSlot || condition.Feature != "noun_number" {
			continue
		}
		value := string(condition.Value)
		if value != "singular" && value != "plural" || number != "" && number != value {
			return "", false
		}
		number = value
	}
	return number, true
}

func variantSubjectPersonThird(variant construction.Variant, subjectSlot string) bool {
	for _, condition := range variant.Conditions {
		if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == subjectSlot && condition.Feature == "person" && string(condition.Value) != "third" {
			return false
		}
	}
	return true
}

func variantTenseConflicts(variant construction.Variant, predicateSlot, expected string) bool {
	for _, condition := range variant.Conditions {
		if condition.Kind == construction.ConditionFeatureEquals && condition.SlotID == predicateSlot && condition.Feature == "tense" && string(condition.Value) != expected {
			return true
		}
	}
	return false
}

func sequenceHasSlot(sequence []construction.SequenceToken, slot string) bool {
	for _, token := range sequence {
		if token.Kind == construction.TokenSlot && token.SlotID == slot {
			return true
		}
	}
	return false
}

func sequenceHasTerminal(sequence []construction.SequenceToken, terminal string) bool {
	for _, token := range sequence {
		if token.Kind == construction.TokenTerminal && token.Terminal == terminal {
			return true
		}
	}
	return false
}

func sequenceAuxiliary(sequence []construction.SequenceToken) string {
	for _, token := range sequence {
		if token.Kind == construction.TokenTerminal && (token.Terminal == "does" || token.Terminal == "do" || token.Terminal == "did") {
			return token.Terminal
		}
	}
	return ""
}

func mergeDerivedNegativeRoutes(routes []englishPresentNegationRoute) []englishPresentNegationRoute {
	merged := make([]englishPresentNegationRoute, 0, len(routes))
	positions := map[string]int{}
	for _, route := range routes {
		key := route.sense + "\x00" + route.family + "\x00" + route.frame + "\x00" + string(route.subjectRole) + "\x00" + string(route.objectRole)
		position, exists := positions[key]
		if !exists {
			positions[key] = len(merged)
			merged = append(merged, route)
			continue
		}
		current := &merged[position]
		for number, variants := range route.variantsByNumber {
			if current.variantsByNumber[number] == nil {
				current.variantsByNumber[number] = map[string]bool{}
			}
			for id := range variants {
				current.variantsByNumber[number][id] = true
			}
		}
		for id := range route.pastVariants {
			current.pastVariants[id] = true
		}
		if current.singularVariant == "" {
			current.singularVariant = route.singularVariant
		}
		if current.pluralVariant == "" {
			current.pluralVariant = route.pluralVariant
		}
	}
	return merged
}

func (resolver englishNegationResolver) routeForSense(sense ids.SenseID) (englishPresentNegationRoute, bool) {
	var matched englishPresentNegationRoute
	found := false
	for _, route := range resolver.routes() {
		if route.sense == sense.String() {
			if found && negationRouteKey(matched) != negationRouteKey(route) {
				return englishPresentNegationRoute{}, false
			}
			matched, found = route, true
		}
	}
	if found {
		return matched, true
	}
	for _, spec := range englishAttributionSpecs {
		if spec.sense == sense.String() {
			return englishPresentNegationRoute{
				concept: spec.concept, sense: spec.sense, verb: "be", lexeme: spec.lexeme,
				family: "construction:en.attribution_negative", singularVariant: "construction:en.attribution_negative.singular",
				pluralVariant: "construction:en.attribution_negative.plural", subjectRole: "bearer", statePredicate: true,
			}, true
		}
	}
	return englishPresentNegationRoute{}, false
}

// Legacy callers outside the sentence pipeline retain the established
// non-intransitive route lookup. Pipeline builders use resolver methods with
// their immutable compiled registry and indices.
func englishPresentNegationRouteForSense(sense ids.SenseID) (englishPresentNegationRoute, bool) {
	return (englishNegationResolver{}).routeForSense(sense)
}

func (resolver englishNegationResolver) routeForEvent(event meaning.Node) (englishPresentNegationRoute, bool) {
	var matched englishPresentNegationRoute
	found := false
	for _, route := range resolver.routes() {
		if route.concept == string(event.Concept) && event.Kind == meaning.KindEvent {
			if found && negationRouteKey(matched) != negationRouteKey(route) {
				return englishPresentNegationRoute{}, false
			}
			matched, found = route, true
		}
	}
	return matched, found
}

func negationRouteKey(route englishPresentNegationRoute) string {
	return route.concept + "\x00" + route.sense + "\x00" + route.frame + "\x00" + route.family + "\x00" + string(route.subjectRole) + "\x00" + string(route.objectRole)
}

func (resolver englishNegationResolver) routeForPredicate(predicate meaning.Node) (englishPresentNegationRoute, bool) {
	if predicate.Kind == meaning.KindState && string(predicate.Concept) == englishIdentificationNegationRoute.concept {
		return englishIdentificationNegationRoute, true
	}
	if predicate.Kind == meaning.KindState {
		var selected englishPresentNegationRoute
		found := false
		for _, route := range resolver.routes() {
			if route.family != "construction:en.transitive_negative" || route.objectRole == "" || route.concept != string(predicate.Concept) ||
				!englishPredicateHasRoles(predicate, route.subjectRole, route.objectRole) {
				continue
			}
			if found && negationRouteKey(selected) != negationRouteKey(route) {
				return englishPresentNegationRoute{}, false
			}
			selected, found = route, true
		}
		if found {
			return selected, true
		}
		for _, spec := range englishAttributionSpecs {
			if spec.concept == string(predicate.Concept) {
				sense, err := ids.NewSenseID(strings.TrimPrefix(spec.sense, "sense:"))
				if err == nil {
					return resolver.routeForSense(sense)
				}
			}
		}
	}
	return resolver.routeForEvent(predicate)
}

func englishPredicateHasRoles(predicate meaning.Node, subjectRole, objectRole meaning.Role) bool {
	if len(predicate.Participants) != 2 || subjectRole == "" || objectRole == "" || subjectRole == objectRole {
		return false
	}
	roles := make(map[meaning.Role]bool, len(predicate.Participants))
	for _, participant := range predicate.Participants {
		if roles[participant.Role] {
			return false
		}
		roles[participant.Role] = true
	}
	return roles[subjectRole] && roles[objectRole]
}

func (resolver englishNegationResolver) routeForFamily(family ids.ConstructionID) (englishPresentNegationRoute, bool) {
	for _, route := range resolver.routes() {
		if route.family == family.String() {
			return route, true
		}
	}
	return englishPresentNegationRoute{}, false
}

func (route englishPresentNegationRoute) variantForNumber(number string) string {
	switch number {
	case "singular":
		return route.singularVariant
	case "plural":
		return route.pluralVariant
	default:
		return ""
	}
}

func (route englishPresentNegationRoute) variantMatchesNumber(variantID, number string) bool {
	if route.variantsByNumber != nil {
		return route.variantsByNumber[number][variantID]
	}
	if route.concept == "concept:closing_object" && number == "plural" && variantID == "construction:en.transitive_negative.svo.closing_object.present.plural.subject_the.object_the_plural" {
		return true
	}
	if route.concept == "concept:giving" {
		if number == "singular" {
			return strings.HasPrefix(variantID, route.family+".present.")
		}
		return number == "plural" && strings.HasPrefix(variantID, route.family+".present_plural.")
	}
	base := route.variantForNumber(number)
	if base == "" {
		return false
	}
	if strings.HasPrefix(variantID, route.family+".svo.past") || number == "singular" && strings.HasPrefix(variantID, route.pluralVariant) {
		return false
	}
	return variantID == base || strings.HasPrefix(variantID, base+".")
}

func (route englishPresentNegationRoute) pastVariantAllowed(variantID string) bool {
	if route.pastVariants != nil {
		return route.pastVariants[variantID]
	}
	if route.concept == "concept:giving" {
		return strings.HasPrefix(variantID, route.family+".past.")
	}
	return false
}

func englishStaticPastNegationRoute(route englishPresentNegationRoute) bool {
	if route.family == "construction:en.intransitive" || route.family == "construction:en.transitive_negative" && route.pastVariants != nil {
		return route.pastVariants != nil && len(route.pastVariants) > 0
	}
	switch route.concept {
	case "concept:giving":
		return true
	default:
		return false
	}
}

func (resolver englishNegationResolver) negationInput(lattice analyzer.LexicalLattice, candidates []parser.PhraseCandidate) (englishPresentNegationRoute, string, string, string, text.Span, bool) {
	notSpan, hasNot := englishTerminalSpan(lattice, "not")
	if !hasNot {
		return englishPresentNegationRoute{}, "", "", "", text.Span{}, false
	}
	var matched englishPresentNegationRoute
	var person string
	var number string
	var tense string
	found := false
	for index := 0; index+2 < len(lattice.Tokens); index++ {
		auxiliary := lattice.Tokens[index].Token.Folded
		candidateNumber := ""
		candidatePerson := "third"
		candidateTense := "present"
		switch auxiliary {
		case "does":
			candidatePerson = "third"
			candidateNumber = "singular"
		case "do":
			var agreementOK bool
			candidatePerson, candidateNumber, agreementOK = englishAuxiliarySubjectAgreement(candidates, index)
			if !agreementOK || !englishDoAuxiliaryMatches(auxiliary, candidatePerson, candidateNumber) {
				return englishPresentNegationRoute{}, "", "", "", text.Span{}, false
			}
		case "did":
			candidateTense = "past"
		default:
			continue
		}
		if lattice.Tokens[index+1].Token.Folded != "not" {
			continue
		}
		route, ok := resolver.routeForVerb(lattice.Tokens[index+2].Token.Folded)
		if !ok || candidateTense == "past" && !englishStaticPastNegationRoute(route) ||
			route.requiresAt && (index+3 >= len(lattice.Tokens) || lattice.Tokens[index+3].Token.Folded != "at") || found {
			return englishPresentNegationRoute{}, "", "", "", text.Span{}, false
		}
		matched, person, number, tense, found = route, candidatePerson, candidateNumber, candidateTense, true
	}
	if !found {
		return englishPresentNegationRoute{}, "", "", "", text.Span{}, false
	}
	return matched, person, number, tense, notSpan, true
}

func englishAuxiliarySubjectAgreement(candidates []parser.PhraseCandidate, auxiliaryIndex int) (string, string, bool) {
	person, number := "", ""
	for _, candidate := range candidates {
		if candidate.Category != construction.CategoryNP || candidate.TokenEnd != auxiliaryIndex || candidate.TokenStart >= auxiliaryIndex {
			continue
		}
		candidatePerson := string(candidate.Features["person"])
		if candidatePerson == "" {
			candidatePerson = string(candidate.Features["pronoun_person"])
		}
		if candidatePerson == "" {
			candidatePerson = "third"
		}
		candidateNumber := string(candidate.Features["number"])
		if candidateNumber == "" {
			candidateNumber = string(candidate.Features["noun_number"])
		}
		if candidateNumber == "" {
			candidateNumber = string(candidate.Features["pronoun_number"])
		}
		if candidatePerson != "first" && candidatePerson != "second" && candidatePerson != "third" ||
			candidateNumber != "" && candidateNumber != "singular" && candidateNumber != "plural" ||
			person != "" && person != candidatePerson || number != "" && candidateNumber != "" && number != candidateNumber {
			return "", "", false
		}
		person = candidatePerson
		if candidateNumber != "" {
			number = candidateNumber
		}
	}
	return person, number, person != ""
}

func englishDoAuxiliaryMatches(auxiliary, person, number string) bool {
	switch auxiliary {
	case "does":
		return person == "third" && number == "singular"
	case "do":
		return person == "first" || person == "second" || person == "third" && number == "plural"
	default:
		return false
	}
}

func (resolver englishNegationResolver) negatedPredicateContext(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate, candidates []parser.PhraseCandidate) (englishPresentNegationRoute, string, string, string, bool) {
	if candidate.Category != construction.CategoryVP || candidate.TokenStart < 2 || candidate.TokenStart >= len(lattice.Tokens) {
		return englishPresentNegationRoute{}, "", "", "", false
	}
	auxiliary := lattice.Tokens[candidate.TokenStart-2].Token.Folded
	if lattice.Tokens[candidate.TokenStart-1].Token.Folded != "not" {
		return englishPresentNegationRoute{}, "", "", "", false
	}
	auxiliaryIndex := candidate.TokenStart - 2
	number := ""
	person := "third"
	tense := "present"
	switch auxiliary {
	case "does":
		person = "third"
		number = "singular"
	case "do":
		var found bool
		person, number, found = englishAuxiliarySubjectAgreement(candidates, auxiliaryIndex)
		if !found || !englishDoAuxiliaryMatches(auxiliary, person, number) {
			return englishPresentNegationRoute{}, "", "", "", false
		}
	case "did":
		tense = "past"
		number = string(candidate.Features["number"])
		if number == "" {
			for _, subject := range candidates {
				if subject.Category != construction.CategoryNP || subject.TokenEnd != auxiliaryIndex || subject.TokenStart >= auxiliaryIndex || subject.Features["person"] != "third" {
					continue
				}
				subjectNumber := string(subject.Features["number"])
				if subjectNumber != "singular" && subjectNumber != "plural" {
					continue
				}
				if number != "" && number != subjectNumber {
					return englishPresentNegationRoute{}, "", "", "", false
				}
				number = subjectNumber
			}
		}
		if number != "singular" && number != "plural" {
			return englishPresentNegationRoute{}, "", "", "", false
		}
	default:
		return englishPresentNegationRoute{}, "", "", "", false
	}
	for _, route := range resolver.routes() {
		if candidate.LexemeID.String() != route.lexeme || len(candidate.SenseIDs) > 0 && !containsEnglishSenseName(candidate.SenseIDs, strings.TrimPrefix(route.sense, "sense:")) {
			continue
		}
		if tense == "past" && !englishStaticPastNegationRoute(route) {
			return englishPresentNegationRoute{}, "", "", "", false
		}
		if route.requiresAt && (candidate.TokenStart+1 >= len(lattice.Tokens) || lattice.Tokens[candidate.TokenStart+1].Token.Folded != "at") {
			return englishPresentNegationRoute{}, "", "", "", false
		}
		return route, person, number, tense, true
	}
	return englishPresentNegationRoute{}, "", "", "", false
}

func englishNegationAuxiliaryCandidate(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) bool {
	start := candidate.TokenStart
	if candidate.Category != construction.CategoryVP || start < 0 || start+1 >= len(lattice.Tokens) {
		return false
	}
	return (lattice.Tokens[start].Token.Folded == "does" || lattice.Tokens[start].Token.Folded == "do" || lattice.Tokens[start].Token.Folded == "did") &&
		lattice.Tokens[start+1].Token.Folded == "not"
}

func (resolver englishNegationResolver) negatedPresent(source semantic.MeaningCandidate) (meaning.Node, meaning.Node, englishPresentNegationRoute, bool) {
	if source.Meaning == nil {
		return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
	}
	var scope meaning.Node
	for _, node := range source.Meaning.Nodes {
		if node.Kind != meaning.KindScope {
			continue
		}
		if scope.ID != "" || node.Operator != meaning.OperatorNegation {
			return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
		}
		scope = node
	}
	if scope.ID == "" {
		return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
	}
	for _, node := range source.Meaning.Nodes {
		if node.ID != scope.Target {
			continue
		}
		if route, found := resolver.attributedNegativeSeeingRoute(source, node); found {
			return scope, node, route, true
		}
		route, found := resolver.routeForPredicate(node)
		if found {
			return scope, node, route, true
		}
	}
	return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
}

// attributedNegativeSeeingRoute resolves the dedicated object-attribute
// negative family from its approved sense/frame/variant bindings. It remains
// separate from lemma routing so ordinary and attributed "see" do not collide.
func (resolver englishNegationResolver) attributedNegativeSeeingRoute(source semantic.MeaningCandidate, event meaning.Node) (englishPresentNegationRoute, bool) {
	const familyText = "construction:en.attributed_seeing_object_negative"
	const senseText = "sense:en.see.visual"
	const frameText = "en.see.visual.attributed_negative"
	if source.Meaning == nil || event.Kind != meaning.KindEvent || event.Concept != "concept:seeing" ||
		resolver.registry == nil || resolver.valency == nil || resolver.lexicon == nil || resolver.variants == nil {
		return englishPresentNegationRoute{}, false
	}
	var experiencer, stimulus meaning.NodeID
	for _, participant := range event.Participants {
		switch participant.Role {
		case "experiencer":
			if experiencer != "" {
				return englishPresentNegationRoute{}, false
			}
			experiencer = participant.Target
		case "stimulus":
			if stimulus != "" {
				return englishPresentNegationRoute{}, false
			}
			stimulus = participant.Target
		default:
			return englishPresentNegationRoute{}, false
		}
	}
	if experiencer == "" || stimulus == "" {
		return englishPresentNegationRoute{}, false
	}
	objectAttribute := false
	for _, node := range source.Meaning.Nodes {
		if node.Kind == meaning.KindState && englishSeeingColorConcept(string(node.Concept)) && len(node.Participants) == 1 &&
			node.Participants[0].Role == "bearer" && node.Participants[0].Target == stimulus {
			objectAttribute = true
		}
	}
	if !objectAttribute {
		return englishPresentNegationRoute{}, false
	}
	familyID, familyErr := ids.NewConstructionID(strings.TrimPrefix(familyText, "construction:"))
	senseID, senseErr := ids.NewSenseID(strings.TrimPrefix(senseText, "sense:"))
	if familyErr != nil || senseErr != nil {
		return englishPresentNegationRoute{}, false
	}
	familyRecord, exists := resolver.registry.ByID(familyID)
	variantIndex := resolver.variants[familyID]
	if !exists || variantIndex == nil || !familyHasOperator(familyRecord.Family, construction.OperatorID("negate")) ||
		!familyHasCapability(familyRecord.Family, construction.Capability("negation-negated")) {
		return englishPresentNegationRoute{}, false
	}
	var frame *valency.ValencyFrame
	for _, candidate := range resolver.valency.FramesBySenseID(senseID) {
		if string(candidate.ID) == frameText {
			copy := candidate
			frame = &copy
			break
		}
	}
	if frame == nil || !valencyFrameAllowsFamily(*frame, familyID) {
		return englishPresentNegationRoute{}, false
	}
	subjectRole := meaning.Role("")
	subjectSlot, predicateSlot := "", ""
	for _, argument := range frame.Arguments {
		if !argument.Required {
			continue
		}
		switch argument.GrammaticalFunction {
		case "subject":
			if subjectRole != "" || argument.Role != "experiencer" {
				return englishPresentNegationRoute{}, false
			}
			subjectRole = argument.Role
			subjectSlot = argument.ID
		case "object":
			if argument.Role != "stimulus" {
				return englishPresentNegationRoute{}, false
			}
		default:
			return englishPresentNegationRoute{}, false
		}
	}
	for _, slot := range familyRecord.Family.SlotSpecs {
		if slot.Category == construction.CategoryVP && slot.IsMainPredicate() {
			if predicateSlot != "" {
				return englishPresentNegationRoute{}, false
			}
			predicateSlot = slot.ID
		}
	}
	if subjectRole == "" || subjectSlot == "" || predicateSlot == "" {
		return englishPresentNegationRoute{}, false
	}
	approvedVariants := map[string]bool{}
	for _, binding := range resolver.registry.BindingsByFamilyID(familyID) {
		if !binding.Approved || binding.SenseID != senseID || string(binding.ValencyFrameID) != frameText {
			continue
		}
		variantID, err := construction.ParseVariantID(binding.VariantID.String())
		if err != nil {
			continue
		}
		compiled, exists := variantIndex.ByID(variantID)
		if !exists || compiled.Variant.DirectionPolicy == construction.DirectionAnalyze || !variantHasRequiredSlots(familyRecord.Family, compiled.Variant) ||
			!sequenceHasSlot(compiled.Variant.Sequence, subjectSlot) || !sequenceHasSlot(compiled.Variant.Sequence, predicateSlot) ||
			(!sequenceHasTerminal(compiled.Variant.Sequence, "not") && !englishNegativePhraseRecipe(familyRecord.Family, compiled.Variant)) {
			continue
		}
		approvedVariants[binding.VariantID.String()] = true
	}
	if len(approvedVariants) == 0 {
		return englishPresentNegationRoute{}, false
	}
	lexicalSense, exists := resolver.lexicon.BySenseID(senseID)
	if !exists {
		return englishPresentNegationRoute{}, false
	}
	lexeme, exists := resolver.lexicon.LookupLexeme(lexicalSense.LexemeID)
	if !exists || lexeme.Language != "en" || lexeme.POS != "verb" {
		return englishPresentNegationRoute{}, false
	}
	coreBindings := resolver.lexicon.TransferBindings(senseID)
	if len(coreBindings) != 1 || !coreBindings[0].TransferAuthorized() || coreBindings[0].CoreRef.String() != string(event.Concept) {
		return englishPresentNegationRoute{}, false
	}
	return englishPresentNegationRoute{
		concept: string(event.Concept), sense: senseText, verb: strings.ToLower(lexeme.Lemma), lexeme: lexicalSense.LexemeID.String(),
		family: familyText, frame: frameText, singularVariant: familyText + ".svo", subjectRole: subjectRole,
		variantsByNumber: map[string]map[string]bool{"singular": approvedVariants, "plural": {}}, pastVariants: map[string]bool{},
	}, true
}

func englishNegatedPresent(source semantic.MeaningCandidate) (meaning.Node, meaning.Node, englishPresentNegationRoute, bool) {
	return (englishNegationResolver{}).negatedPresent(source)
}

func (resolver englishNegationResolver) negationNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, route englishPresentNegationRoute, candidates []realizer.TargetCandidate) (string, bool) {
	_, event, sourceRoute, found := resolver.negatedPresent(source)
	if !found || event.ID != predicate.NodeID || sourceRoute.family != route.family || sourceRoute.sense != route.sense {
		return "", false
	}
	referent, exists := englishReferentForRole(source, event.ID, route.subjectRole)
	if !exists || referent.Multiplicity != "" && !englishMultiplicityEvidenceConsistent(source, event.ID, route.subjectRole, referent, predicate.TransferBinding.CoreRef, true) {
		return "", false
	}
	number, _, ok := englishPredicateNumber(source, predicate, candidates)
	if !ok || !englishFrameEvidenceConsistent(source, event.ID, predicate.TransferBinding.CoreRef) {
		return "", false
	}
	return number, true
}

func englishNegationNumber(source semantic.MeaningCandidate, predicate realizer.TargetCandidate, route englishPresentNegationRoute, candidates []realizer.TargetCandidate) (string, bool) {
	return (englishNegationResolver{}).negationNumber(source, predicate, route, candidates)
}

func englishTerminalSpan(lattice analyzer.LexicalLattice, terminal string) (text.Span, bool) {
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
