package a1pl

import "github.com/clickbytes/golingua/tests/matrix"

var QuestionAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-Q-002-PL-EN:analyze", MeaningID: "", Input: "Czy to jest kot?", MustPreserve: []string{"pytanie o identyfikację jako kota", "wypowiedź pozostaje pytaniem, nie twierdzeniem"}, AmbiguityNote: "Polskie to i brak rodzajnika nie rozstrzygają this/that/it ani angielskiego a/the.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-Q-004-PL-EN:analyze", MeaningID: "", Input: "Gdzie jest kot?", MustPreserve: []string{"pytanie dotyczy miejsca kota", "wypowiedź nie podaje miejsca jako odpowiedzi"}, AmbiguityNote: "Angielski rodzajnik zależy od kontekstu; zdanie nie wskazuje miejsca ani nie wymaga odpowiedzi w danych.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var QuestionRealizationCases = []matrix.RealizationCase{
	{ID: "A1-Q-001-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Czy to jest kot?"}, AmbiguityNote: "Angielskie this oraz a nie mają jednego obowiązkowego odpowiednika w polskiej formie.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-Q-003-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Gdzie jest kot?"}, AmbiguityNote: "Polski nie wyraża angielskiego the; propozycja nie dodaje domyślnego miejsca.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
