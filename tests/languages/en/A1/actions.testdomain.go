package a1en

import "github.com/clickbytes/golingua/tests/matrix"

var ActionAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-AC-001-EN-PL:analyze", MeaningID: "seeing.teacher-dog", Input: "The teacher sees the dog.", MustPreserve: []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdanie twierdzi, że widzenie zachodzi"}, AmbiguityNote: "Polskie końcówki odróżniają role; angielskie the nie ma osobnego odpowiednika.", InputOrigin: "data/sentences/observed-2026-10-23-g10-14-full-corpus.tsv#S001", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-AC-003-EN-PL:analyze", MeaningID: "seeing.dog-teacher", Input: "The dog sees the teacher.", MustPreserve: []string{"pies jest doświadczającym widzenia (experiencer)", "nauczyciel jest bodźcem widzenia (stimulus)", "role uczestników są odwrócone względem przypadku bazowego"}, AmbiguityNote: "Propozycja zachowuje odwrócenie ról semantycznych; polskie formy przypadków powinny je odzwierciedlać.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var ActionRealizationCases = []matrix.RealizationCase{
	{ID: "A1-AC-002-PL-EN:realize", MeaningID: "seeing.teacher-dog", ProposalTexts: []string{"The teacher sees the dog.", "The teacher sees a dog.", "A teacher sees the dog.", "A teacher sees a dog."}, AllowedTexts: []string{"The teacher sees the dog.", "The teacher sees a dog.", "A teacher sees the dog.", "A teacher sees a dog."}, RequiredTexts: []string{"The teacher sees the dog.", "The teacher sees a dog.", "A teacher sees the dog.", "A teacher sees a dog."}, AllowedAssumptionFields: []string{"tense", "person", "number", "determiner"}, AmbiguityNote: "Polszczyzna nie koduje angielskiej określoności; dobór rodzajników zależy od kontekstu.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-AC-004-PL-EN:realize", MeaningID: "seeing.dog-teacher", ProposalTexts: []string{"The dog sees the teacher.", "A dog sees the teacher.", "The dog sees a teacher.", "A dog sees a teacher."}, AllowedTexts: []string{"The dog sees the teacher.", "A dog sees the teacher.", "The dog sees a teacher.", "A dog sees a teacher."}, RequiredTexts: []string{"The dog sees the teacher.", "A dog sees the teacher.", "The dog sees a teacher.", "A dog sees a teacher."}, AllowedAssumptionFields: []string{"tense", "person", "number", "determiner"}, AmbiguityNote: "Rodzajniki pozostają nieokreślone przez polski; kolejność i role nie mogą zamienić wykonawcy z obiektem.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
