package a1pl

import "github.com/clickbytes/golingua/tests/matrix"

var LocationQuantityAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-LQ-002-PL-EN:analyze", MeaningID: "", Input: "Kot jest na krześle.", MustPreserve: []string{"kot znajduje się na krześle", "relacja przestrzenna na"}, AmbiguityNote: "Polski nie przesądza rodzajników angielskich; kontekst rozstrzyga, czy obiekty są określone.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-LQ-004-PL-EN:analyze", MeaningID: "", Input: "Są dwa koty.", MustPreserve: []string{"istnieją koty", "liczba wynosi dokładnie dwa"}, AmbiguityNote: "Sama liczebność nie określa, czy angielskie koty są określone; propozycja zachowuje konstrukcję egzystencjalną.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var LocationQuantityRealizationCases = []matrix.RealizationCase{
	{ID: "A1-LQ-001-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Kot jest na krześle."}, AmbiguityNote: "Polski nie koduje angielskiego the; dobór przyimka może zależeć od zamierzonej relacji przestrzennej.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-LQ-003-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Są dwa koty."}, AmbiguityNote: "Szyk i forma orzeczenia w polszczyźnie zależą od składni zdania; ta propozycja używa konstrukcji egzystencjalnej.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
