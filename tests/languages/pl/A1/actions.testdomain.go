package a1pl

import "github.com/clickbytes/golingua/tests/matrix"

var ActionAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-AC-002-PL-EN:analyze", MeaningID: "seeing.teacher-dog", Input: "Nauczyciel widzi psa.", MustPreserve: []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdanie twierdzi, że widzenie zachodzi"}, AmbiguityNote: "Polszczyzna nie koduje angielskiej określoności; dobór rodzajników zależy od kontekstu.", InputOrigin: "data/sentences/observed-2026-10-23-g10-14-full-corpus.tsv#S001", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-AC-004-PL-EN:analyze", MeaningID: "seeing.dog-teacher", Input: "Pies widzi nauczyciela.", MustPreserve: []string{"pies jest doświadczającym widzenia (experiencer)", "nauczyciel jest bodźcem widzenia (stimulus)", "role uczestników są odwrócone względem przypadku bazowego"}, AmbiguityNote: "Rodzajniki pozostają nieokreślone przez polski; kolejność i role nie mogą zamienić wykonawcy z obiektem.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var ActionRealizationCases = []matrix.RealizationCase{
	{ID: "A1-AC-001-EN-PL:realize", MeaningID: "seeing.teacher-dog", ProposalTexts: []string{"Nauczyciel widzi psa."}, AllowedTexts: []string{"Nauczyciel widzi psa."}, RequiredTexts: []string{"Nauczyciel widzi psa."}, AllowedAssumptionFields: []string{"tense", "person", "number"}, AmbiguityNote: "Polskie końcówki odróżniają role; angielskie the nie ma osobnego odpowiednika.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-AC-003-EN-PL:realize", MeaningID: "seeing.dog-teacher", ProposalTexts: []string{"Pies widzi nauczyciela."}, AllowedTexts: []string{"Pies widzi nauczyciela."}, RequiredTexts: []string{"Pies widzi nauczyciela."}, AllowedAssumptionFields: []string{"tense", "person", "number"}, AmbiguityNote: "Propozycja zachowuje odwrócenie ról semantycznych; polskie formy przypadków powinny je odzwierciedlać.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
