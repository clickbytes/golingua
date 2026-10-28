package en

import (
	"strings"

	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/text"
)

type englishPresentNegationRoute struct {
	concept         string
	sense           string
	verb            string
	lexeme          string
	family          string
	singularVariant string
	pluralVariant   string
	subjectRole     meaning.Role
	requiresAt      bool
	statePredicate  bool
}

var englishIdentificationNegationRoute = englishPresentNegationRoute{
	concept: "concept:class_membership", sense: "sense:en.be.class_membership", verb: "is",
	lexeme: "lexeme:en.be.v", family: "construction:en.identification_negative",
	singularVariant: "construction:en.identification_negative.this", subjectRole: "member", statePredicate: true,
}

var englishPresentNegationRoutes = []englishPresentNegationRoute{
	{concept: "concept:seeing", sense: "sense:en.see.visual", verb: "see", lexeme: "lexeme:en.see.v", family: "construction:en.transitive_negative", singularVariant: "construction:en.transitive_negative.svo", pluralVariant: "construction:en.transitive_negative.svo.plural", subjectRole: "experiencer"},
	{concept: "concept:hearing", sense: "sense:en.hear.auditory", verb: "hear", lexeme: "lexeme:en.hear.v", family: "construction:en.transitive_negative", singularVariant: "construction:en.transitive_negative.svo", pluralVariant: "construction:en.transitive_negative.svo.plural", subjectRole: "experiencer"},
	{concept: "concept:looking", sense: "sense:en.look.directed", verb: "look", lexeme: "lexeme:en.look.v", family: "construction:en.oblique_negative", singularVariant: "construction:en.oblique_negative.svo", pluralVariant: "construction:en.oblique_negative.svo.plural", subjectRole: "agent", requiresAt: true},
	{concept: "concept:liking", sense: "sense:en.liking.event", verb: "like", lexeme: "lexeme:en.liking.v", family: "construction:en.transitive_negative", singularVariant: "construction:en.transitive_negative.svo", pluralVariant: "construction:en.transitive_negative.svo.plural", subjectRole: "experiencer"},
	{concept: "concept:watching", sense: "sense:en.watching.event", verb: "watch", lexeme: "lexeme:en.watching.v", family: "construction:en.transitive_negative", singularVariant: "construction:en.transitive_negative.svo", pluralVariant: "construction:en.transitive_negative.svo.plural", subjectRole: "viewer"},
}

func englishPresentNegationRouteForVerb(verb string) (englishPresentNegationRoute, bool) {
	for _, route := range englishPresentNegationRoutes {
		if route.verb == verb {
			return route, true
		}
	}
	return englishPresentNegationRoute{}, false
}

func englishPresentNegationRouteForSense(sense ids.SenseID) (englishPresentNegationRoute, bool) {
	for _, route := range englishPresentNegationRoutes {
		if route.sense == sense.String() {
			return route, true
		}
	}
	return englishPresentNegationRoute{}, false
}

func englishPresentNegationRouteForEvent(event meaning.Node) (englishPresentNegationRoute, bool) {
	for _, route := range englishPresentNegationRoutes {
		if route.concept == string(event.Concept) && event.Kind == meaning.KindEvent {
			return route, true
		}
	}
	return englishPresentNegationRoute{}, false
}

func englishPresentNegationRouteForPredicate(predicate meaning.Node) (englishPresentNegationRoute, bool) {
	if predicate.Kind == meaning.KindState && string(predicate.Concept) == englishIdentificationNegationRoute.concept {
		return englishIdentificationNegationRoute, true
	}
	return englishPresentNegationRouteForEvent(predicate)
}

func englishPresentNegationRouteForFamily(family ids.ConstructionID) (englishPresentNegationRoute, bool) {
	for _, route := range englishPresentNegationRoutes {
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
	base := route.variantForNumber(number)
	if base == "" {
		return false
	}
	if number == "singular" && strings.HasPrefix(variantID, route.pluralVariant) {
		return false
	}
	return variantID == base || strings.HasPrefix(variantID, base+".")
}

func englishPresentNegationInput(lattice analyzer.LexicalLattice) (englishPresentNegationRoute, string, text.Span, bool) {
	notSpan, hasNot := englishTerminalSpan(lattice, "not")
	if !hasNot {
		return englishPresentNegationRoute{}, "", text.Span{}, false
	}
	var matched englishPresentNegationRoute
	var number string
	found := false
	for index := 0; index+2 < len(lattice.Tokens); index++ {
		auxiliary := lattice.Tokens[index].Token.Folded
		candidateNumber := ""
		switch auxiliary {
		case "does":
			candidateNumber = "singular"
		case "do":
			candidateNumber = "plural"
		default:
			continue
		}
		if lattice.Tokens[index+1].Token.Folded != "not" {
			continue
		}
		route, ok := englishPresentNegationRouteForVerb(lattice.Tokens[index+2].Token.Folded)
		if !ok || route.requiresAt && (index+3 >= len(lattice.Tokens) || lattice.Tokens[index+3].Token.Folded != "at") || found {
			return englishPresentNegationRoute{}, "", text.Span{}, false
		}
		matched, number, found = route, candidateNumber, true
	}
	if !found {
		return englishPresentNegationRoute{}, "", text.Span{}, false
	}
	return matched, number, notSpan, true
}

func englishNegatedPredicateContext(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) (englishPresentNegationRoute, string, bool) {
	if candidate.Category != construction.CategoryVP || candidate.TokenStart < 2 || candidate.TokenStart >= len(lattice.Tokens) {
		return englishPresentNegationRoute{}, "", false
	}
	auxiliary := lattice.Tokens[candidate.TokenStart-2].Token.Folded
	if lattice.Tokens[candidate.TokenStart-1].Token.Folded != "not" {
		return englishPresentNegationRoute{}, "", false
	}
	number := ""
	switch auxiliary {
	case "does":
		number = "singular"
	case "do":
		number = "plural"
	default:
		return englishPresentNegationRoute{}, "", false
	}
	for _, route := range englishPresentNegationRoutes {
		if candidate.LexemeID.String() != route.lexeme {
			continue
		}
		if route.requiresAt && (candidate.TokenStart+1 >= len(lattice.Tokens) || lattice.Tokens[candidate.TokenStart+1].Token.Folded != "at") {
			return englishPresentNegationRoute{}, "", false
		}
		return route, number, true
	}
	return englishPresentNegationRoute{}, "", false
}

func englishNegationAuxiliaryCandidate(lattice analyzer.LexicalLattice, candidate parser.PhraseCandidate) bool {
	start := candidate.TokenStart
	if candidate.Category != construction.CategoryVP || start < 0 || start+1 >= len(lattice.Tokens) {
		return false
	}
	return (lattice.Tokens[start].Token.Folded == "does" || lattice.Tokens[start].Token.Folded == "do") &&
		lattice.Tokens[start+1].Token.Folded == "not"
}

func englishNegatedPresent(source semantic.MeaningCandidate) (meaning.Node, meaning.Node, englishPresentNegationRoute, bool) {
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
		route, found := englishPresentNegationRouteForPredicate(node)
		if found {
			return scope, node, route, true
		}
	}
	return meaning.Node{}, meaning.Node{}, englishPresentNegationRoute{}, false
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
