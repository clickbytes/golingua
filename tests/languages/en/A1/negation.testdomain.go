package a1en

import "github.com/clickbytes/golingua/tests/matrix"

var NegationAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-NEG-001-EN-PL:analyze", MeaningID: "", Input: "This is not a cat.", MustPreserve: []string{"identyfikacja kota jest zanegowana", "negacja obejmuje predykat identyfikacji"}, AmbiguityNote: "Wybór this/that/it i rodzajnika jest zależny od kontekstu; polska propozycja nie koduje rodzaju angielskiego kota.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-NEG-003-EN-PL:analyze", MeaningID: "", Input: "The teacher does not see the dog.", MustPreserve: []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdarzenie widzenia jest zanegowane"}, AmbiguityNote: "Angielskie rodzajniki nie mają odrębnego odpowiednika w polszczyźnie; ich dobór zależy od kontekstu.", InputOrigin: "data/sentences/observed-2026-10-23-g10-17-present-negation.tsv#NEG-EN-SEE-SG", OwnerAcceptance: "NOT_EVALUATED"},
}

var NegationRealizationCases = []matrix.RealizationCase{
	{ID: "A1-NEG-002-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"This is not a cat.", "That is not a cat.", "It is not a cat."}, AmbiguityNote: "Polskie to nie rozstrzyga angielskiego wskazania this/that/it ani określoności rzeczownika.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-NEG-004-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"The teacher does not see the dog.", "The teacher does not see a dog.", "A teacher does not see the dog.", "A teacher does not see a dog."}, AmbiguityNote: "Zachowujemy zakres negacji i role; dobór rodzajników angielskich zależy od kontekstu.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
