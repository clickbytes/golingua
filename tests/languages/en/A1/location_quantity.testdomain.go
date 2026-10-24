package a1en

import "github.com/clickbytes/golingua/tests/matrix"

var LocationQuantityAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-LQ-001-EN-PL:analyze", MeaningID: "", Input: "The cat is on the chair.", MustPreserve: []string{"kot znajduje się na krześle", "relacja przestrzenna on"}, AmbiguityNote: "Polski nie koduje angielskiego the; dobór przyimka może zależeć od zamierzonej relacji przestrzennej.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-LQ-003-EN-PL:analyze", MeaningID: "", Input: "There are two cats.", MustPreserve: []string{"istnieją koty", "liczba wynosi dokładnie dwa"}, AmbiguityNote: "Szyk i forma orzeczenia w polszczyźnie zależą od składni zdania; ta propozycja używa konstrukcji egzystencjalnej.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var LocationQuantityRealizationCases = []matrix.RealizationCase{
	{ID: "A1-LQ-002-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"The cat is on the chair.", "A cat is on a chair."}, AmbiguityNote: "Polski nie przesądza rodzajników angielskich; kontekst rozstrzyga, czy obiekty są określone.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-LQ-004-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"There are two cats."}, AmbiguityNote: "Sama liczebność nie określa, czy angielskie koty są określone; propozycja zachowuje konstrukcję egzystencjalną.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
