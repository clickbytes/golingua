package a1

// Negacja dotyczy identyfikacji lub zdarzenia widzenia.
var NegationCases = []Case{
	{
		ID: "A1-NEG-001-EN-PL", From: "en", To: "pl", Input: "This is not a cat.",
		ReferenceTexts: []string{"To nie jest kot."},
		MustPreserve:   []string{"identyfikacja kota jest zanegowana", "negacja obejmuje predykat identyfikacji"},
		AmbiguityNote:  "Wybór this/that/it i rodzajnika jest zależny od kontekstu; polska propozycja nie koduje rodzaju angielskiego kota.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-NEG-002-PL-EN", From: "pl", To: "en", Input: "To nie jest kot.",
		ReferenceTexts: []string{"This is not a cat.", "That is not a cat.", "It is not a cat."},
		MustPreserve:   []string{"identyfikacja kota jest zanegowana", "negacja obejmuje predykat identyfikacji"},
		AmbiguityNote:  "Polskie to nie rozstrzyga angielskiego wskazania this/that/it ani określoności rzeczownika.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-NEG-003-EN-PL", From: "en", To: "pl", Input: "The teacher does not see the dog.",
		ReferenceTexts: []string{"Nauczyciel nie widzi psa."},
		MustPreserve:   []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdarzenie widzenia jest zanegowane"},
		AmbiguityNote:  "Angielskie rodzajniki nie mają odrębnego odpowiednika w polszczyźnie; ich dobór zależy od kontekstu.",
		InputOrigin:    "data/sentences/observed-2026-10-23-g10-17-present-negation.tsv#NEG-EN-SEE-SG", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-NEG-004-PL-EN", From: "pl", To: "en", Input: "Nauczyciel nie widzi psa.",
		ReferenceTexts: []string{"The teacher does not see the dog.", "The teacher does not see a dog.", "A teacher does not see the dog.", "A teacher does not see a dog."},
		MustPreserve:   []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdarzenie widzenia jest zanegowane"},
		AmbiguityNote:  "Zachowujemy zakres negacji i role; dobór rodzajników angielskich zależy od kontekstu.",
		InputOrigin:    "data/sentences/observed-2026-10-23-g10-17-present-negation.tsv#NEG-PL-SEE-SG", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
}
