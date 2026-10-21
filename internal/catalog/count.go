package catalog

import (
	"fmt"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/lexicon"
)

// Count reports only compiled CORE definitions, never labels, forms or senses.
func Count() (actual, target int, err error) {
	graph, err := Compile()
	if err != nil {
		return 0, TargetConcepts, err
	}
	return graph.Len(), TargetConcepts, nil
}

// CoverageStatus describes the strongest mapping status for one concept and
// one language. Pending candidates never become approved by counting them.
type CoverageStatus string

const (
	CoverageAbsent   CoverageStatus = "absent"
	CoveragePending  CoverageStatus = "pending"
	CoverageOther    CoverageStatus = "non_transferable"
	CoverageApproved CoverageStatus = "approved"
)

type LanguageCount struct {
	Lexemes          int `json:"lexemes"`
	Senses           int `json:"senses"`
	Bindings         int `json:"bindings"`
	PendingConcepts  int `json:"pending_concepts"`
	OtherConcepts    int `json:"non_transferable_concepts"`
	ApprovedConcepts int `json:"approved_concepts"`
	AbsentConcepts   int `json:"absent_concepts"`
}

type ConceptCoverage struct {
	ConceptID string         `json:"concept_id"`
	Polish    CoverageStatus `json:"pl"`
	English   CoverageStatus `json:"en"`
}

type Inventory struct {
	Concepts      int               `json:"concepts"`
	Target        int               `json:"target"`
	Polish        LanguageCount     `json:"pl"`
	English       LanguageCount     `json:"en"`
	PerConcept    []ConceptCoverage `json:"per_concept"`
	ApprovedPairs int               `json:"approved_pairs"`
}

// Inspect counts the exact compiled snapshots supplied by the caller. It has
// no language-plugin dependency and never selects a sense or approves data.
// PerConcept follows canonical CORE ID order from ConceptGraph.Concepts.
func Inspect(graph *core.ConceptGraph, polish, english *lexicon.Index) (Inventory, error) {
	if graph == nil || polish == nil || english == nil {
		return Inventory{}, fmt.Errorf("catalog inspect requires CORE, Polish and English compiled snapshots")
	}
	result := Inventory{Concepts: graph.Len(), Target: TargetConcepts}
	result.Polish.Lexemes, result.Polish.Senses, result.Polish.Bindings = polish.Len()
	result.English.Lexemes, result.English.Senses, result.English.Bindings = english.Len()
	result.PerConcept = make([]ConceptCoverage, 0, graph.Len())
	for _, concept := range graph.Concepts() {
		pl := strongestCoverage(polish.ByCoreConcept(concept.ID))
		en := strongestCoverage(english.ByCoreConcept(concept.ID))
		addCoverage(&result.Polish, pl)
		addCoverage(&result.English, en)
		if pl == CoverageApproved && en == CoverageApproved {
			result.ApprovedPairs++
		}
		result.PerConcept = append(result.PerConcept, ConceptCoverage{ConceptID: concept.ID.String(), Polish: pl, English: en})
	}
	return result, nil
}

func strongestCoverage(bindings []lexicon.SemanticBinding) CoverageStatus {
	status := CoverageAbsent
	for _, binding := range bindings {
		if binding.TransferAuthorized() {
			return CoverageApproved
		}
		if binding.Status == lexicon.BindingPending {
			status = CoveragePending
		} else if status == CoverageAbsent {
			status = CoverageOther
		}
	}
	return status
}

func addCoverage(count *LanguageCount, status CoverageStatus) {
	switch status {
	case CoverageApproved:
		count.ApprovedConcepts++
	case CoveragePending:
		count.PendingConcepts++
	case CoverageOther:
		count.OtherConcepts++
	default:
		count.AbsentConcepts++
	}
}
