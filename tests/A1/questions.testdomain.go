package a1

// Pytania zachowują typ pytania i szukany element.
var QuestionCases = []Case{
	{
		ID: "A1-Q-001-EN-PL", From: "en", To: "pl", Input: "Is this a cat?",
		ReferenceTexts: []string{"Czy to jest kot?"},
		MustPreserve:   []string{"pytanie o identyfikację jako kota", "wypowiedź pozostaje pytaniem, nie twierdzeniem"},
		AmbiguityNote:  "Angielskie this oraz a nie mają jednego obowiązkowego odpowiednika w polskiej formie.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-Q-002-PL-EN", From: "pl", To: "en", Input: "Czy to jest kot?",
		ReferenceTexts: []string{"Is this a cat?", "Is that a cat?", "Is it a cat?"},
		MustPreserve:   []string{"pytanie o identyfikację jako kota", "wypowiedź pozostaje pytaniem, nie twierdzeniem"},
		AmbiguityNote:  "Polskie to i brak rodzajnika nie rozstrzygają this/that/it ani angielskiego a/the.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-Q-003-EN-PL", From: "en", To: "pl", Input: "Where is the cat?",
		ReferenceTexts: []string{"Gdzie jest kot?"},
		MustPreserve:   []string{"pytanie dotyczy miejsca kota", "wypowiedź nie podaje miejsca jako odpowiedzi"},
		AmbiguityNote:  "Polski nie wyraża angielskiego the; propozycja nie dodaje domyślnego miejsca.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-Q-004-PL-EN", From: "pl", To: "en", Input: "Gdzie jest kot?",
		ReferenceTexts: []string{"Where is the cat?", "Where is a cat?"},
		MustPreserve:   []string{"pytanie dotyczy miejsca kota", "wypowiedź nie podaje miejsca jako odpowiedzi"},
		AmbiguityNote:  "Angielski rodzajnik zależy od kontekstu; zdanie nie wskazuje miejsca ani nie wymaga odpowiedzi w danych.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
}
