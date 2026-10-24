package a1en

import "github.com/clickbytes/golingua/tests/matrix"

var PossessionAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-PO-001-EN-PL:analyze", MeaningID: "", Input: "I have a cat.", MustPreserve: []string{"mówiący jest posiadaczem", "posiadanym obiektem jest kot", "liczba pojedyncza"}, AmbiguityNote: "Czasownik mam wskazuje pierwszą osobę mimo pominiętego zaimka; polski nie koduje angielskiego rodzajnika a, a określoność wymaga kontekstu.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-PO-003-EN-PL:analyze", MeaningID: "", Input: "She has a book.", MustPreserve: []string{"posiadaczką jest wskazana osoba rodzaju żeńskiego", "posiadanym obiektem jest książka", "liczba pojedyncza"}, AmbiguityNote: "Angielski rodzajnik a nie ma obowiązkowego odpowiednika w polszczyźnie.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var PossessionRealizationCases = []matrix.RealizationCase{
	{ID: "A1-PO-002-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"I have a cat.", "I have the cat."}, AmbiguityNote: "Czasownik mam wskazuje pierwszą osobę; angielski a/the wymaga kontekstu, bo polski nie koduje tu określoności.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-PO-004-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"She has a book.", "She has the book."}, AmbiguityNote: "Angielski rodzajnik i określoność zależą od kontekstu; polski przypadek nie określa ich samodzielnie.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
