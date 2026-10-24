package a1pl

import "github.com/clickbytes/golingua/tests/matrix"

var NegationAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-NEG-002-PL-EN:analyze", MeaningID: "", Input: "To nie jest kot.", MustPreserve: []string{"identyfikacja kota jest zanegowana", "negacja obejmuje predykat identyfikacji"}, AmbiguityNote: "Polskie to nie rozstrzyga angielskiego wskazania this/that/it ani określoności rzeczownika.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-NEG-004-PL-EN:analyze", MeaningID: "", Input: "Nauczyciel nie widzi psa.", MustPreserve: []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdarzenie widzenia jest zanegowane"}, AmbiguityNote: "Zachowujemy zakres negacji i role; dobór rodzajników angielskich zależy od kontekstu.", InputOrigin: "data/sentences/observed-2026-10-23-g10-17-present-negation.tsv#NEG-PL-SEE-SG", OwnerAcceptance: "NOT_EVALUATED"},
}

var NegationRealizationCases = []matrix.RealizationCase{
	{ID: "A1-NEG-001-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"To nie jest kot."}, AmbiguityNote: "Wybór this/that/it i rodzajnika jest zależny od kontekstu; polska propozycja nie koduje rodzaju angielskiego kota.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-NEG-003-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Nauczyciel nie widzi psa."}, AmbiguityNote: "Angielskie rodzajniki nie mają odrębnego odpowiednika w polszczyźnie; ich dobór zależy od kontekstu.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
