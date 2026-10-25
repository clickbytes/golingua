package en

import (
	"github.com/clickbytes/golingua/internal/construction"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/parser"
	"github.com/clickbytes/golingua/internal/runtime"
	"github.com/clickbytes/golingua/internal/semantic"
)

func addEnglishPersonalReferences(e runtime.CompositionEvidence, candidates []parser.PhraseCandidate, forest parser.Forest) runtime.CompositionEvidence {
	subjectCandidates := map[string]bool{}
	for _, node := range forest.Nodes {
		if node.Category != "clause" {
			continue
		}
		for _, alt := range node.Alternatives {
			if alt.FamilyID.String() != "construction:en.transitive_personal" {
				continue
			}
			for _, slot := range alt.Slots {
				if slot.SlotID == "subject" {
					for _, id := range slot.CandidateIDs {
						subjectCandidates[id] = true
					}
				}
			}
		}
	}
	for _, c := range candidates {
		if c.Category != construction.CategoryNP || !subjectCandidates[c.ID] {
			continue
		}
		anchor := ""
		switch c.LexemeID.String() {
		case "lexeme:en.i.pron":
			anchor = "speaker"
		case "lexeme:en.we.pron":
			anchor = "speaker_inclusive_group"
		case "lexeme:en.you.pron":
			anchor = "addressee"
		}
		if anchor != "" {
			e.References = append(e.References, semantic.ReferenceEvidence{CandidateID: c.ID, Anchor: &meaning.DeicticAnchor{Kind: "participant", Value: anchor}})
		}
	}
	return e
}
