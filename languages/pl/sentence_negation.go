package pl

import (
	"github.com/clickbytes/golingua/internal/analyzer"
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/semantic"
	"github.com/clickbytes/golingua/internal/text"
)

type polishPresentNegationRoute struct {
	concept     string
	sense       string
	lexeme      string
	family      string
	variant     string
	subjectRole meaning.Role
	kind        meaning.NodeKind
}

var polishPresentNegationRoutes = []polishPresentNegationRoute{
	{concept: "concept:seeing", sense: "sense:pl.widziec.visual", lexeme: "lexeme:pl.widziec.v", family: "construction:pl.transitive_negative", variant: "construction:pl.transitive_negative.svo", subjectRole: "experiencer", kind: meaning.KindEvent},
	{concept: "concept:hearing", sense: "sense:pl.slyszec.auditory", lexeme: "lexeme:pl.slyszec.v", family: "construction:pl.transitive_negative", variant: "construction:pl.transitive_negative.svo", subjectRole: "experiencer", kind: meaning.KindEvent},
	{concept: "concept:looking", sense: "sense:pl.patrzec.directed", lexeme: "lexeme:pl.patrzec.v", family: "construction:pl.oblique_negative", variant: "construction:pl.oblique_negative.svo", subjectRole: "agent", kind: meaning.KindEvent},
	{concept: "concept:class_membership", sense: "sense:pl.byc.class_membership", lexeme: "lexeme:pl.byc.v", family: "construction:pl.identification_negative", variant: "construction:pl.identification_negative.to", subjectRole: "member", kind: meaning.KindState},
	{concept: "concept:liking", sense: "sense:pl.lubic.like", lexeme: "lexeme:pl.lubic.v", family: "construction:pl.transitive_negative", variant: "construction:pl.transitive_negative.svo", subjectRole: "experiencer", kind: meaning.KindEvent},
	{concept: "concept:watching", sense: "sense:pl.ogladac.view", lexeme: "lexeme:pl.ogladac.v", family: "construction:pl.transitive_negative", variant: "construction:pl.transitive_negative.svo", subjectRole: "viewer", kind: meaning.KindEvent},
}

func polishPresentNegationRouteForSense(sense ids.SenseID) (polishPresentNegationRoute, bool) {
	for _, route := range polishPresentNegationRoutes {
		if route.sense == sense.String() {
			return route, true
		}
	}
	return polishPresentNegationRoute{}, false
}

func polishPresentNegationRouteForEvent(event meaning.Node) (polishPresentNegationRoute, bool) {
	for _, route := range polishPresentNegationRoutes {
		if route.concept == string(event.Concept) && event.Kind == route.kind {
			return route, true
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
	return polishPresentNegationRoute{}, false
}

func polishPresentNegationVariantAllowed(route polishPresentNegationRoute, family, variant string) bool {
	if route.concept == "concept:seeing" && family == "construction:pl.attributed_seeing_object_negative" {
		return variant == "construction:pl.attributed_seeing_object_negative.svo" || variant == "construction:pl.attributed_seeing_object_negative.ovs"
	}
	if family != route.family {
		return false
	}
	if variant == route.variant {
		return true
	}
	return route.concept == "concept:seeing" && route.family == "construction:pl.transitive_negative" &&
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
	if candidate.Category != construction.CategoryVP || candidate.LexemeID.String() != route.lexeme ||
		candidate.Features["verb_form"] != "finite" || candidate.Features["tense"] != "present" ||
		candidate.Features["person"] != "third" {
		return false
	}
	return candidate.Features["number"] == "singular" || candidate.Features["number"] == "plural"
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
