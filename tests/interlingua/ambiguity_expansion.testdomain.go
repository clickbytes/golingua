// Package interlingua contains independent A1.12 ambiguity Meaning fixtures.
package interlingua

import (
	"fmt"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"strconv"
	"strings"
)

type ambiguitySpec struct {
	index                 int
	predicate, roles, gap string
}

// This semantic catalogue stores predicates, roles, referent counts and
// licensed operators only; bilingual source strings remain in their cases.
var ambiguitySpecs = []ambiguitySpec{
	{index: 1, predicate: "seeing", roles: "experiencer=teacher:one,stimulus=bird:one", gap: ""},
	{index: 2, predicate: "hearing", roles: "experiencer=doctor:one,stimulus=dog:one", gap: ""},
	{index: 3, predicate: "reading", roles: "reader=mother:one,material=book:one", gap: ""},
	{index: 4, predicate: "seeing", roles: "experiencer=father:one,stimulus=car:one", gap: ""},
	{index: 5, predicate: "drinking", roles: "drinker=dog:one,liquid=water:mass", gap: ""},
	{index: 6, predicate: "eating", roles: "eater=cat:one,food=bread:mass", gap: ""},
	{index: 7, predicate: "seeing", roles: "experiencer=teacher:many,stimulus=bird:many", gap: ""},
	{index: 8, predicate: "hearing", roles: "experiencer=doctor:many,stimulus=cat:many", gap: ""},
	{index: 9, predicate: "reading", roles: "reader=doctor:many,material=book:one", gap: ""},
	{index: 10, predicate: "reading", roles: "reader=mother:many,material=message:one", gap: ""},
	{index: 11, predicate: "reading", roles: "reader=father:many,material=book:one", gap: ""},
	{index: 12, predicate: "reading", roles: "reader=child:many,material=message:one", gap: ""},
	{index: 13, predicate: "reading", roles: "reader=teacher:many,material=letter_document:one", gap: ""},
	{index: 14, predicate: "reading", roles: "reader=person:many:speaker_group,material=message:one", gap: ""},
	{index: 15, predicate: "seeing", roles: "experiencer=child:one,stimulus=cat:one", gap: ""},
	{index: 16, predicate: "seeing", roles: "experiencer=dog:one,stimulus=child:one", gap: ""},
	{index: 17, predicate: "hearing", roles: "experiencer=child:many,stimulus=dog:one", gap: ""},
	{index: 18, predicate: "seeing", roles: "experiencer=mother:one,stimulus=child:one", gap: ""},
	{index: 19, predicate: "giving", roles: "giver=teacher:one,recipient=child:one,theme=apple_fruit:one", gap: ""},
	{index: 20, predicate: "giving", roles: "giver=child:one,recipient=mother:one,theme=book:one", gap: ""},
	{index: 21, predicate: "seeing", roles: "experiencer=cat:many,stimulus=dog:many", gap: ""},
	{index: 22, predicate: "seeing", roles: "experiencer=dog:many,stimulus=cat:many", gap: ""},
	{index: 23, predicate: "seeing", roles: "experiencer=bird:many,stimulus=cat:many", gap: ""},
	{index: 24, predicate: "seeing", roles: "experiencer=cat:many,stimulus=bird:many", gap: ""},
	{index: 25, predicate: "hearing", roles: "experiencer=dog:many,stimulus=cat:many", gap: ""},
	{index: 26, predicate: "hearing", roles: "experiencer=cat:many,stimulus=dog:many", gap: ""},
	{index: 27, predicate: "seeing", roles: "experiencer=dog:one:proximal,stimulus=cat:one:distal", gap: ""},
	{index: 28, predicate: "seeing", roles: "experiencer=dog:one:distal,stimulus=cat:one:proximal", gap: ""},
	{index: 29, predicate: "seeing", roles: "experiencer=dog:many:proximal,stimulus=cat:many:distal", gap: ""},
	{index: 30, predicate: "seeing", roles: "experiencer=dog:many:distal,stimulus=cat:many:proximal", gap: ""},
	{index: 31, predicate: "hearing", roles: "experiencer=mother:one:proximal,stimulus=child:one:distal", gap: ""},
	{index: 32, predicate: "hearing", roles: "experiencer=child:one:distal,stimulus=mother:one:proximal", gap: ""},
	{index: 33, predicate: "", roles: "", gap: "Meaning has no explicit quantifier-variable binding or scope relation for every teacher to some dog varying per teacher."},
	{index: 34, predicate: "", roles: "", gap: "Meaning has no explicit quantifier-variable binding or scope relation for some teacher to every dog with a shared teacher."},
	{index: 35, predicate: "seeing", roles: "experiencer=teacher:one,stimulus=cat:many:exactly2", gap: ""},
	{index: 36, predicate: "seeing", roles: "experiencer=dog:one,stimulus=teacher:many:exactly2", gap: ""},
	{index: 37, predicate: "", roles: "", gap: "PL permits impersonal absence or pro-drop third-person negative have; EN selects absence. Meaning cannot represent both finite full readings without conflation."},
	{index: 38, predicate: "", roles: "", gap: "Meaning cannot bind the with-child accompaniment phrase specifically to mother versus observer without adding a relation."},
	{index: 39, predicate: "", roles: "", gap: "Meaning cannot bind the on-table location specifically to cat versus seeing event/observer without adding an unsupported relation."},
	{index: 40, predicate: "", roles: "", gap: "Meaning cannot bind the with-book association specifically to teacher versus observer without asserting ownership."},
	{index: 41, predicate: "seeing", roles: "experiencer=child:many:exactly2,stimulus=dog:one", gap: ""},
	{index: 42, predicate: "seeing", roles: "experiencer=teacher:one,stimulus=teacher:one", gap: ""},
	{index: 43, predicate: "", roles: "", gap: "Meaning cannot bind this child specifically to the stated mother; child_of_person alone does not encode which parent."},
}

// AmbiguityExpansion builds one licensed interpretation. IDs follow
// a1.12.amb.NNN.en|pl[.past][.offspring][.swap]. Past is English read rows
// 9-14 only; offspring selects child_of_person; swap is the alternative Polish
// role binding for syncretic plural rows 21-26. Unmarked child is age-child.
// Rows 33-34,37-40,43 intentionally have no full I graph.
func AmbiguityExpansion(id string) (IdentificationMeaning, bool, error) {
	p := strings.Split(id, ".")
	if len(p) < 5 || len(p) > 8 || p[0] != "a1" || p[1] != "12" || p[2] != "amb" || (p[4] != "en" && p[4] != "pl") {
		return IdentificationMeaning{}, false, nil
	}
	index, err := strconv.Atoi(p[3])
	if err != nil || index < 1 || index > 43 {
		return IdentificationMeaning{}, false, nil
	}
	var spec ambiguitySpec
	for _, s := range ambiguitySpecs {
		if s.index == index {
			spec = s
			break
		}
	}
	if spec.index == 0 || spec.gap != "" {
		return IdentificationMeaning{}, false, nil
	}
	past, offspring, swap := false, false, false
	for _, v := range p[5:] {
		switch v {
		case "past":
			if past {
				return IdentificationMeaning{}, false, nil
			}
			past = true
		case "offspring":
			if offspring {
				return IdentificationMeaning{}, false, nil
			}
			offspring = true
		case "swap":
			if swap {
				return IdentificationMeaning{}, false, nil
			}
			swap = true
		default:
			return IdentificationMeaning{}, false, nil
		}
	}
	canonical := make([]string, 0, 3)
	if past {
		canonical = append(canonical, "past")
	}
	if offspring {
		canonical = append(canonical, "offspring")
	}
	if swap {
		canonical = append(canonical, "swap")
	}
	if strings.Join(p[5:], ".") != strings.Join(canonical, ".") {
		return IdentificationMeaning{}, false, nil
	}
	hasChild := strings.Contains(spec.roles, "child:")
	if offspring && !hasChild {
		return IdentificationMeaning{}, false, nil
	}
	if past && (p[4] != "en" || index < 9 || index > 14) {
		return IdentificationMeaning{}, false, nil
	}
	if swap && (p[4] != "pl" || index < 21 || index > 26) {
		return IdentificationMeaning{}, false, nil
	}
	return buildAmbiguity(spec, p[4], past, offspring, swap)
}

// AmbiguityExpansionIDs is the canonical ordered list of licensed unique readings.
func AmbiguityExpansionIDs() []string {
	var out []string
	for _, s := range ambiguitySpecs {
		if s.gap != "" {
			continue
		}
		for _, lang := range []string{"en", "pl"} {
			base := fmt.Sprintf("a1.12.amb.%03d.%s", s.index, lang)
			frames := []string{""}
			if lang == "en" && s.index >= 9 && s.index <= 14 {
				frames = append(frames, ".past")
			}
			variants := []string{""}
			if strings.Contains(s.roles, "child:") {
				variants = append(variants, ".offspring")
			}
			for _, f := range frames {
				for _, v := range variants {
					out = append(out, base+f+v)
				}
			}
			if lang == "pl" && s.index >= 21 && s.index <= 26 {
				out = append(out, base+".swap")
			}
		}
	}
	return out
}

func buildAmbiguity(spec ambiguitySpec, language string, past, offspring, swap bool) (IdentificationMeaning, bool, error) {
	pred, err := ids.NewConceptID(spec.predicate)
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("ambiguity predicate: %w", err)
	}
	roles := strings.Split(spec.roles, ",")
	nodes := make([]meaning.Node, 0, len(roles)+2)
	parts := make([]meaning.Participant, 0, len(roles))
	mentions := make([]meaning.Mention, 0, 2)
	quantifierTarget := meaning.NodeID("")
	quantifierCount := 0
	for i, raw := range roles {
		pair := strings.SplitN(strings.TrimSpace(raw), "=", 2)
		if len(pair) != 2 {
			return IdentificationMeaning{}, false, fmt.Errorf("bad ambiguity role %q", raw)
		}
		d := strings.Split(pair[1], ":")
		concept := d[0]
		count := ""
		anchor := ""
		exact := false
		for _, attr := range d[1:] {
			switch attr {
			case "one", "many", "mass":
				count = attr
			case "proximal", "distal":
				anchor = attr
			case "speaker_group":
				anchor = "speaker_group"
			case "exactly2":
				exact = true
			}
		}
		if concept == "child" && offspring {
			concept = "child_of_person"
		}
		cid, e := ids.NewConceptID(concept)
		if e != nil {
			return IdentificationMeaning{}, false, fmt.Errorf("ambiguity referent concept: %w", e)
		}
		rid := meaning.NodeID(fmt.Sprintf("r%d", i+1))
		ref, e := meaning.NewReferent(rid, cid)
		if e != nil {
			return IdentificationMeaning{}, false, e
		}
		if count == "one" {
			ref.Multiplicity = meaning.MultiplicityOne
		} else if count == "many" || exact {
			ref.Multiplicity = meaning.MultiplicityMany
		}
		nodes = append(nodes, ref)
		parts = append(parts, meaning.Participant{Role: meaning.Role(pair[0]), Target: rid})
		if anchor != "" {
			val := anchor
			if language == "pl" && anchor == "proximal" {
				val = "demonstrative"
			}
			kind := "deixis"
			if anchor == "speaker_group" {
				kind = "participant"
				val = "speaker_inclusive_group"
			}
			mentions = append(mentions, meaning.Mention{ID: meaning.MentionID(fmt.Sprintf("m%d", i+1)), NodeID: rid, Anchor: &meaning.DeicticAnchor{Kind: kind, Value: val}})
		}
		if exact {
			quantifierTarget = rid
			quantifierCount = 2
		}
	}
	if swap {
		if len(parts) != 2 {
			return IdentificationMeaning{}, false, fmt.Errorf("swap row %d does not have two roles", spec.index)
		}
		parts[0].Target, parts[1].Target = parts[1].Target, parts[0].Target
	}
	event, err := meaning.NewEvent("pred", pred, parts)
	if err != nil {
		return IdentificationMeaning{}, false, err
	}
	event.TemporalFrame = meaning.TemporalFramePresent
	if past {
		event.TemporalFrame = meaning.TemporalFramePast
	}
	nodes = append([]meaning.Node{event}, nodes...)
	roots := []meaning.NodeID{"pred"}
	if quantifierTarget != "" {
		q, e := meaning.NewQuantifier("exactly_two", quantifierTarget, meaning.Quantifier{Kind: meaning.QuantifierExactly, Count: &quantifierCount})
		if e != nil {
			return IdentificationMeaning{}, false, e
		}
		nodes = append(nodes, q)
		roots = append(roots, q.ID)
	}
	graph, err := meaning.Build(roots, nodes, meaning.DefaultBudget())
	if err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("build ambiguity graph: %w", err)
	}
	ctx := meaning.RequestContext{ID: "a1-12-ambiguity", Mentions: mentions}
	if err := meaning.ValidateReferences(graph, ctx, meaning.DefaultBudget()); err != nil {
		return IdentificationMeaning{}, false, fmt.Errorf("validate ambiguity references: %w", err)
	}
	return IdentificationMeaning{Graph: graph, Context: ctx}, true, nil
}
