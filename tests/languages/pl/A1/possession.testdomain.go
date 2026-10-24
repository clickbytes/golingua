package a1pl

import "github.com/clickbytes/golingua/tests/matrix"

var PossessionAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-PO-002-PL-EN:analyze", MeaningID: "", Input: "Mam kota.", MustPreserve: []string{"mówiący jest posiadaczem", "posiadanym obiektem jest kot", "liczba pojedyncza"}, AmbiguityNote: "Czasownik mam wskazuje pierwszą osobę; angielski a/the wymaga kontekstu, bo polski nie koduje tu określoności.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-PO-004-PL-EN:analyze", MeaningID: "", Input: "Ona ma książkę.", MustPreserve: []string{"posiadaczką jest wskazana osoba rodzaju żeńskiego", "posiadanym obiektem jest książka", "liczba pojedyncza"}, AmbiguityNote: "Angielski rodzajnik i określoność zależą od kontekstu; polski przypadek nie określa ich samodzielnie.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var PossessionRealizationCases = []matrix.RealizationCase{
	{ID: "A1-PO-001-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Mam kota."}, AmbiguityNote: "Czasownik mam wskazuje pierwszą osobę mimo pominiętego zaimka; polski nie koduje angielskiego rodzajnika a, a określoność wymaga kontekstu.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-PO-003-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Ona ma książkę."}, AmbiguityNote: "Angielski rodzajnik a nie ma obowiązkowego odpowiednika w polszczyźnie.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
