// Package interlingua contains independent A1.13 composition Meaning fixtures.
package interlingua

import (
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
)

type compositionExpansionSpec struct {
	index int
	// Clauses are predicate[!](role=referent,...)[@past], separated by semicolons.
	clauses string
	// Referents are label=concept:multiplicity, with mass/unknown omitted.
	referents string
	black     string
	request   bool
	gap       string
}

// This catalogue records only concepts, roles, counts, time, negation, and
// licensed child sense. Surface strings stay in the language case files.
var compositionExpansionSpecs = []compositionExpansionSpec{
	{1, "reading(reader=mother,material=book);drinking(drinker=father,liquid=water)", "mother=mother:one,book=book:one,father=father:one,water=water", "", false, ""},
	{2, "seeing(experiencer=dog,stimulus=cat);hearing(experiencer=bird,stimulus=doctor)", "dog=dog:one,cat=cat:one,bird=bird:one,doctor=doctor:one", "", false, ""},
	{3, "reading(reader=teacher,material=book);hearing(experiencer=teacher,stimulus=bird)", "teacher=teacher:one,book=book:one,bird=bird:one", "", false, ""},
	{4, "drinking(drinker=child,liquid=water);eating(eater=child,food=bread)", "child=child:one,water=water,bread=bread", "", false, ""},
	{5, "sleeping(sleeper=dog);running(runner=cat)", "dog=dog:one,cat=cat:one", "", false, ""},
	{6, "reading(reader=teacher,material=book)@past;seeing(experiencer=teacher,stimulus=dog)", "teacher=teacher:one,book=book:one,dog=dog:one", "", false, ""},
	{7, "seeing(experiencer=mother,stimulus=dog);hearing!(experiencer=mother,stimulus=cat)", "mother=mother:one,dog=dog:one,cat=cat:one", "", false, ""},
	{8, "seeing!(experiencer=teacher,stimulus=cat);hearing(experiencer=mother,stimulus=bird)", "teacher=teacher:one,cat=cat:one,mother=mother:one,bird=bird:one", "", false, ""},
	{9, "seeing!(experiencer=doctor,stimulus=dog);hearing!(experiencer=doctor,stimulus=cat)", "doctor=doctor:one,dog=dog:one,cat=cat:one", "", false, ""},
	{10, "seeing(experiencer=teacher,stimulus=dog);hearing(experiencer=teacher,stimulus=cat)", "teacher=teacher:one,dog=dog:one,cat=cat:one", "dog,cat", false, ""},
	{11, "seeing(experiencer=dog,stimulus=cat);hearing(experiencer=dog,stimulus=bird)", "dog=dog:one,cat=cat:one,bird=bird:one", "dog", false, ""},
	{12, "", "", "", false, "Meaning v1 lacks a heterogeneous coordinated-group representation that preserves the selected collective/distributive ambiguity; separate atomic predicates would not satisfy the group reading."},
	{13, "", "", "", false, "Meaning v1 lacks a heterogeneous group and unresolved collective/distributive event representation; separate atomic predicates would not satisfy the selected subject-group reading."},
	{14, "", "", "", false, "Meaning v1 has no disjunction operator for alternative propositions; a root conjunction would assert the wrong relation."},
	{15, "", "", "", false, "Meaning v1 hypothetical scope cannot encode the selected conditional implication between two propositions."},
	{16, "", "", "", false, "Meaning v1 has no causal relation binding the two propositions."},
	{17, "", "", "", false, "Meaning v1 has no ordered temporal relation between events; discourse order is not event sequence."},
	{18, "", "", "", false, "Meaning v1 temporal_frame records reference time per event but cannot encode overlap between events."},
	{19, "seeing!(experiencer=dog,stimulus=cat)", "dog=dog:one,cat=cat:one", "cat", false, ""},
	{20, "hearing!(experiencer=cat,stimulus=dog)", "cat=cat:one,dog=dog:one", "cat", false, ""},
	{21, "opening_object(agent=person,object=door)", "person=person:addressee,door=door:one", "", true, ""},
	{22, "", "", "", false, "Meaning v1 request act does not preserve explicit please/politeness; request force alone does not meet the selected full meaning."},
	{23, "opening_object!(agent=person,object=door)", "person=person:addressee,door=door:one", "", true, ""},
	{24, "opening_object(agent=person,object=door);closing_object(agent=person,object=window)", "person=person:addressee,door=door:one,window=window:one", "", true, ""},
	{25, "", "", "", false, "Meaning v1 request act does not preserve explicit please/politeness; neither a bare request nor one child sense represents the selected politeness and child-sense ambiguity."},
	{26, "giving(giver=person,recipient=speaker,theme=key)", "person=person:addressee,speaker=person:one:speaker,key=key_tool:one", "", true, ""},
	{27, "", "", "", false, "CORE has no posture-change event for moving into sitting; the existing sitting posture state is not equivalent."},
	{28, "reading(reader=person,material=book);looking(agent=person,target=bird)", "person=person:addressee,book=book:one,bird=bird:one", "", true, ""},
	{29, "", "", "", false, "Meaning v1 cannot preserve both epistemic possibility and permission readings (and the Polish ability reading) without asserting actual sleep."},
	{30, "", "", "", false, "CORE speaking has no content role to bind the embedded sleeping proposition; adding a bare sleep root would assert it."},
	{31, "seeing!(experiencer=dog,stimulus=cat);seeing!(experiencer=dog,stimulus=bird)", "dog=dog:one,cat=cat:one,bird=bird:one", "", false, ""},
	{32, "", "", "", false, "Meaning v1 has no contrast relation; a plain conjunction erases the selected contrast."},
	{33, "seeing(experiencer=mother,stimulus=mother)", "mother=mother:one", "", false, ""},
}

// CompositionExpansion returns one independent, complete composition graph.
// IDs are a1.13.comp.NNN.en|pl, optionally .offspring for row 4's licensed
// child_of_person reading. Unrepresentable full meanings intentionally return
// ok=false; their exact missing relation/operator is recorded in the spec.
func CompositionExpansion(id string) (IdentificationMeaning, bool, error) {
	parts := strings.Split(id, ".")
	if (len(parts) != 5 && len(parts) != 6) || parts[0] != "a1" || parts[1] != "13" || parts[2] != "comp" || (parts[4] != "en" && parts[4] != "pl") {
		return IdentificationMeaning{}, false, nil
	}
	var index int
	if _, err := fmt.Sscanf(parts[3], "%d", &index); err != nil || index < 1 || index > len(compositionExpansionSpecs) {
		return IdentificationMeaning{}, false, nil
	}
	if parts[3] != fmt.Sprintf("%03d", index) {
		return IdentificationMeaning{}, false, nil
	}
	variant := ""
	if len(parts) == 6 {
		variant = parts[5]
		if variant != "offspring" {
			return IdentificationMeaning{}, false, nil
		}
	}
	spec := compositionExpansionSpecs[index-1]
	if spec.index != index || spec.gap != "" || (variant != "" && index != 4) {
		return IdentificationMeaning{}, false, nil
	}
	return buildCompositionExpansion(spec, parts[4], variant)
}

// CompositionExpansionIDs lists each unique represented interpretation.
func CompositionExpansionIDs() []string {
	var out []string
	for _, spec := range compositionExpansionSpecs {
		if spec.gap != "" {
			continue
		}
		for _, lang := range []string{"en", "pl"} {
			base := fmt.Sprintf("a1.13.comp.%03d.%s", spec.index, lang)
			out = append(out, base)
			if spec.index == 4 {
				out = append(out, base+".offspring")
			}
		}
	}
	return out
}

// CompositionExpansionGap reports the precise missing relation/operator for
// an intentionally unrepresented full interpretation.
func CompositionExpansionGap(id string) (string, bool) {
	parts := strings.Split(id, ".")
	if len(parts) != 5 || parts[0] != "a1" || parts[1] != "13" || parts[2] != "comp" || (parts[4] != "en" && parts[4] != "pl") {
		return "", false
	}
	var index int
	if _, err := fmt.Sscanf(parts[3], "%d", &index); err != nil || index < 1 || index > len(compositionExpansionSpecs) || parts[3] != fmt.Sprintf("%03d", index) {
		return "", false
	}
	gap := compositionExpansionSpecs[index-1].gap
	return gap, gap != ""
}

func buildCompositionExpansion(spec compositionExpansionSpec, language, variant string) (IdentificationMeaning, bool, error) {
	refs := map[string]meaning.Node{}
	refOrder := make([]string, 0)
	for _, item := range strings.Split(spec.referents, ",") {
		fields := strings.SplitN(item, "=", 2)
		if len(fields) != 2 {
			return IdentificationMeaning{}, false, fmt.Errorf("composition row %d: malformed referent %q", spec.index, item)
		}
		details := strings.Split(fields[1], ":")
		concept := details[0]
		if concept == "child" && variant == "offspring" {
			concept = "child_of_person"
		}
		cid, err := ids.NewConceptID(concept)
		if err != nil {
			return IdentificationMeaning{}, false, fmt.Errorf("composition row %d referent: %w", spec.index, err)
		}
		id := meaning.NodeID("r" + fields[0])
		node, err := meaning.NewReferent(id, cid)
		if err != nil {
			return IdentificationMeaning{}, false, err
		}
		if len(details) > 1 {
			switch details[1] {
			case "one":
				node.Multiplicity = meaning.MultiplicityOne
			case "many":
				node.Multiplicity = meaning.MultiplicityMany
			}
		}
		if language == "pl" && (spec.index == 21 || spec.index == 23 || spec.index == 24) && fields[0] == "door" {
			// Polish plural-only drzwi does not reveal source referent count.
			node.Multiplicity = ""
		}
		refs[fields[0]] = node
		refOrder = append(refOrder, fields[0])
	}

	nodes := make([]meaning.Node, 0, len(refs)+8)
	roots := make([]meaning.NodeID, 0, 4)
	context := meaning.RequestContext{ID: "a1-13-composition"}
	if spec.request {
		for _, label := range refOrder {
			node := refs[label]
			if node.Concept != "concept:person" {
				continue
			}
			anchor := "addressee"
			if label == "speaker" {
				anchor = "speaker"
			}
			if label == "person" && language == "en" {
				node.Multiplicity = ""
				refs[label] = node
			}
			if label == "person" && language == "pl" {
				node.Multiplicity = meaning.MultiplicityOne
				refs[label] = node
			}
			context.Mentions = append(context.Mentions, meaning.Mention{ID: meaning.MentionID("m_" + label), NodeID: node.ID, Anchor: &meaning.DeicticAnchor{Kind: "participant", Value: anchor}})
		}
	}

	clauses := strings.Split(spec.clauses, ";")
	for clauseIndex, raw := range clauses {
		text := raw
		frame := meaning.TemporalFramePresent
		if strings.HasSuffix(text, "@past") {
			text = strings.TrimSuffix(text, "@past")
			frame = meaning.TemporalFramePast
		}
		negative := strings.Contains(text, "!")
		text = strings.ReplaceAll(text, "!", "")
		open := strings.IndexByte(text, '(')
		if open <= 0 || !strings.HasSuffix(text, ")") {
			return IdentificationMeaning{}, false, fmt.Errorf("composition row %d: malformed clause %q", spec.index, raw)
		}
		predicateName := text[:open]
		predicate, err := ids.NewConceptID(predicateName)
		if err != nil {
			return IdentificationMeaning{}, false, err
		}
		participants := []meaning.Participant{}
		arguments := strings.TrimSuffix(text[open+1:], ")")
		if arguments != "" {
			for _, argument := range strings.Split(arguments, ",") {
				pair := strings.SplitN(argument, "=", 2)
				if len(pair) != 2 {
					return IdentificationMeaning{}, false, fmt.Errorf("composition row %d: malformed argument %q", spec.index, argument)
				}
				target, ok := refs[pair[1]]
				if !ok {
					return IdentificationMeaning{}, false, fmt.Errorf("composition row %d: missing referent %q", spec.index, pair[1])
				}
				participants = append(participants, meaning.Participant{Role: meaning.Role(pair[0]), Target: target.ID})
			}
		}
		eventID := meaning.NodeID(fmt.Sprintf("e%d", clauseIndex+1))
		event, err := meaning.NewEvent(eventID, predicate, participants)
		if err != nil {
			return IdentificationMeaning{}, false, err
		}
		if !spec.request {
			event.TemporalFrame = frame
		}
		nodes = append(nodes, event)
		target := eventID
		if negative {
			negation, err := meaning.NewNegation(meaning.NodeID(fmt.Sprintf("n%d", clauseIndex+1)), eventID)
			if err != nil {
				return IdentificationMeaning{}, false, err
			}
			nodes = append(nodes, negation)
			target = negation.ID
		}
		if spec.request {
			roots = append(roots, target)
		}
	}

	for _, label := range refOrder {
		ref := refs[label]
		nodes = append(nodes, ref)
		if strings.Contains(","+spec.black+",", ","+label+",") {
			blackID, err := ids.NewConceptID("black")
			if err != nil {
				return IdentificationMeaning{}, false, err
			}
			state, err := meaning.NewState(meaning.NodeID("black_"+label), blackID, []meaning.Participant{{Role: "bearer", Target: ref.ID}})
			if err != nil {
				return IdentificationMeaning{}, false, err
			}
			nodes = append(nodes, state)
			roots = append(roots, state.ID)
		}
	}
	if spec.request {
		content := make([]meaning.NodeID, 0, len(roots))
		for _, root := range roots {
			content = append(content, root)
		}
		request, err := meaning.NewDiscourse("request", meaning.ActRequest, content)
		if err != nil {
			return IdentificationMeaning{}, false, err
		}
		nodes = append(nodes, request)
		roots = []meaning.NodeID{"request"}
	} else {
		for i := range clauses {
			target := meaning.NodeID(fmt.Sprintf("e%d", i+1))
			if strings.Contains(clauses[i], "!") {
				target = meaning.NodeID(fmt.Sprintf("n%d", i+1))
			}
			roots = append(roots, target)
		}
	}
	for _, label := range refOrder {
		ref := refs[label]
		if ref.Concept == "concept:person" && !spec.request {
			context.Mentions = append(context.Mentions, meaning.Mention{ID: meaning.MentionID("m_" + strings.TrimPrefix(string(ref.ID), "r")), NodeID: ref.ID, Anchor: &meaning.DeicticAnchor{Kind: "participant", Value: "unresolved"}})
		}
	}
	graph, err := meaning.Build(roots, nodes, meaning.DefaultBudget())
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("composition row %d build: %w", spec.index, err)
	}
	if err := meaning.ValidateReferences(graph, context, meaning.DefaultBudget()); err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("composition row %d references: %w", spec.index, err)
	}
	return IdentificationMeaning{Graph: graph, Context: context}, true, nil
}
