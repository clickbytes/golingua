package translation

import "github.com/clickbytes/golingua/tests/matrix"

var Cases = []matrix.TranslationCase{
	{CoverageTags: []string{"semantic_predicate", "participant_roles", "valency"}, ID: "A1-AC-001-EN-PL", From: "en", To: "pl", AnalysisID: "A1-AC-001-EN-PL:analyze", RealizationID: "A1-AC-001-EN-PL:realize"},
	{CoverageTags: []string{"semantic_predicate", "participant_roles", "valency"}, ID: "A1-AC-002-PL-EN", From: "pl", To: "en", AnalysisID: "A1-AC-002-PL-EN:analyze", RealizationID: "A1-AC-002-PL-EN:realize"},
	{CoverageTags: []string{"semantic_predicate", "participant_roles", "valency"}, ID: "A1-AC-003-EN-PL", From: "en", To: "pl", AnalysisID: "A1-AC-003-EN-PL:analyze", RealizationID: "A1-AC-003-EN-PL:realize"},
	{CoverageTags: []string{"semantic_predicate", "participant_roles", "valency"}, ID: "A1-AC-004-PL-EN", From: "pl", To: "en", AnalysisID: "A1-AC-004-PL-EN:analyze", RealizationID: "A1-AC-004-PL-EN:realize"},
	{CoverageTags: []string{"semantic_classification", "reference", "determiner_ambiguity"}, ID: "A1-ID-001-EN-PL", From: "en", To: "pl", AnalysisID: "A1-ID-001-EN-PL:analyze", RealizationID: "A1-ID-001-EN-PL:realize"},
	{CoverageTags: []string{"semantic_classification", "reference", "determiner_ambiguity"}, ID: "A1-ID-002-PL-EN", From: "pl", To: "en", AnalysisID: "A1-ID-002-PL-EN:analyze", RealizationID: "A1-ID-002-PL-EN:realize"},
	{CoverageTags: []string{"semantic_property_black", "reference"}, ID: "A1-ID-003-EN-PL", From: "en", To: "pl", AnalysisID: "A1-ID-003-EN-PL:analyze", RealizationID: "A1-ID-003-EN-PL:realize"},
	{CoverageTags: []string{"semantic_property_black", "reference"}, ID: "A1-ID-004-PL-EN", From: "pl", To: "en", AnalysisID: "A1-ID-004-PL-EN:analyze", RealizationID: "A1-ID-004-PL-EN:realize"},
	{CoverageTags: []string{"semantic_location", "locative"}, ID: "A1-LQ-001-EN-PL", From: "en", To: "pl", AnalysisID: "A1-LQ-001-EN-PL:analyze", RealizationID: "A1-LQ-001-EN-PL:realize"},
	{CoverageTags: []string{"semantic_location", "locative", "case"}, ID: "A1-LQ-002-PL-EN", From: "pl", To: "en", AnalysisID: "A1-LQ-002-PL-EN:analyze", RealizationID: "A1-LQ-002-PL-EN:realize"},
	{CoverageTags: []string{"semantic_quantity", "exactly_two", "number_morphology"}, ID: "A1-LQ-003-EN-PL", From: "en", To: "pl", AnalysisID: "A1-LQ-003-EN-PL:analyze", RealizationID: "A1-LQ-003-EN-PL:realize"},
	{CoverageTags: []string{"semantic_quantity", "exactly_two", "number_morphology"}, ID: "A1-LQ-004-PL-EN", From: "pl", To: "en", AnalysisID: "A1-LQ-004-PL-EN:analyze", RealizationID: "A1-LQ-004-PL-EN:realize"},
	{CoverageTags: []string{"semantic_polarity_scope", "clause_negation"}, ID: "A1-NEG-001-EN-PL", From: "en", To: "pl", AnalysisID: "A1-NEG-001-EN-PL:analyze", RealizationID: "A1-NEG-001-EN-PL:realize"},
	{CoverageTags: []string{"semantic_polarity_scope", "clause_negation"}, ID: "A1-NEG-002-PL-EN", From: "pl", To: "en", AnalysisID: "A1-NEG-002-PL-EN:analyze", RealizationID: "A1-NEG-002-PL-EN:realize"},
	{CoverageTags: []string{"semantic_polarity_scope", "clause_negation"}, ID: "A1-NEG-003-EN-PL", From: "en", To: "pl", AnalysisID: "A1-NEG-003-EN-PL:analyze", RealizationID: "A1-NEG-003-EN-PL:realize"},
	{CoverageTags: []string{"semantic_polarity_scope", "clause_negation"}, ID: "A1-NEG-004-PL-EN", From: "pl", To: "en", AnalysisID: "A1-NEG-004-PL-EN:analyze", RealizationID: "A1-NEG-004-PL-EN:realize"},
	{CoverageTags: []string{"semantic_possession", "participant_roles", "valency", "case"}, ID: "A1-PO-001-EN-PL", From: "en", To: "pl", AnalysisID: "A1-PO-001-EN-PL:analyze", RealizationID: "A1-PO-001-EN-PL:realize"},
	{CoverageTags: []string{"semantic_possession", "participant_roles", "valency", "case"}, ID: "A1-PO-002-PL-EN", From: "pl", To: "en", AnalysisID: "A1-PO-002-PL-EN:analyze", RealizationID: "A1-PO-002-PL-EN:realize"},
	{CoverageTags: []string{"semantic_possession", "participant_roles", "valency", "case"}, ID: "A1-PO-003-EN-PL", From: "en", To: "pl", AnalysisID: "A1-PO-003-EN-PL:analyze", RealizationID: "A1-PO-003-EN-PL:realize"},
	{CoverageTags: []string{"semantic_possession", "participant_roles", "valency", "case"}, ID: "A1-PO-004-PL-EN", From: "pl", To: "en", AnalysisID: "A1-PO-004-PL-EN:analyze", RealizationID: "A1-PO-004-PL-EN:realize"},
	{CoverageTags: []string{"semantic_speech_act", "yes_no_question", "reference"}, ID: "A1-Q-001-EN-PL", From: "en", To: "pl", AnalysisID: "A1-Q-001-EN-PL:analyze", RealizationID: "A1-Q-001-EN-PL:realize"},
	{CoverageTags: []string{"semantic_speech_act", "yes_no_question", "reference"}, ID: "A1-Q-002-PL-EN", From: "pl", To: "en", AnalysisID: "A1-Q-002-PL-EN:analyze", RealizationID: "A1-Q-002-PL-EN:realize"},
	{CoverageTags: []string{"semantic_speech_act", "wh_location", "reference"}, ID: "A1-Q-003-EN-PL", From: "en", To: "pl", AnalysisID: "A1-Q-003-EN-PL:analyze", RealizationID: "A1-Q-003-EN-PL:realize"},
	{CoverageTags: []string{"semantic_speech_act", "wh_location", "reference"}, ID: "A1-Q-004-PL-EN", From: "pl", To: "en", AnalysisID: "A1-Q-004-PL-EN:analyze", RealizationID: "A1-Q-004-PL-EN:realize"},
}
