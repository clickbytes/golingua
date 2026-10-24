package a1en

import "github.com/clickbytes/golingua/tests/matrix"

var QuestionAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-Q-001-EN-PL:analyze", MeaningID: "", Input: "Is this a cat?", MustPreserve: []string{"pytanie o identyfikację jako kota", "wypowiedź pozostaje pytaniem, nie twierdzeniem"}, AmbiguityNote: "Angielskie this oraz a nie mają jednego obowiązkowego odpowiednika w polskiej formie.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-Q-003-EN-PL:analyze", MeaningID: "", Input: "Where is the cat?", MustPreserve: []string{"pytanie dotyczy miejsca kota", "wypowiedź nie podaje miejsca jako odpowiedzi"}, AmbiguityNote: "Polski nie wyraża angielskiego the; propozycja nie dodaje domyślnego miejsca.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var QuestionRealizationCases = []matrix.RealizationCase{
	{ID: "A1-Q-002-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"Is this a cat?", "Is that a cat?", "Is it a cat?"}, AmbiguityNote: "Polskie to i brak rodzajnika nie rozstrzygają this/that/it ani angielskiego a/the.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-Q-004-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"Where is the cat?", "Where is a cat?"}, AmbiguityNote: "Angielski rodzajnik zależy od kontekstu; zdanie nie wskazuje miejsca ani nie wymaga odpowiedzi w danych.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
