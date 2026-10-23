package a1

// Lokalizacja i liczba; propozycje nie rozstrzygają wszystkich kontekstów.
var LocationQuantityCases = []Case{
	{
		ID: "A1-LQ-001-EN-PL", From: "en", To: "pl", Input: "The cat is on the chair.",
		ReferenceTexts: []string{"Kot jest na krześle."},
		MustPreserve:   []string{"kot znajduje się na krześle", "relacja przestrzenna on"},
		AmbiguityNote:  "Polski nie koduje angielskiego the; dobór przyimka może zależeć od zamierzonej relacji przestrzennej.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-LQ-002-PL-EN", From: "pl", To: "en", Input: "Kot jest na krześle.",
		ReferenceTexts: []string{"The cat is on the chair.", "A cat is on a chair."},
		MustPreserve:   []string{"kot znajduje się na krześle", "relacja przestrzenna na"},
		AmbiguityNote:  "Polski nie przesądza rodzajników angielskich; kontekst rozstrzyga, czy obiekty są określone.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-LQ-003-EN-PL", From: "en", To: "pl", Input: "There are two cats.",
		ReferenceTexts: []string{"Są dwa koty."},
		MustPreserve:   []string{"istnieją koty", "liczba wynosi dokładnie dwa"},
		AmbiguityNote:  "Szyk i forma orzeczenia w polszczyźnie zależą od składni zdania; ta propozycja używa konstrukcji egzystencjalnej.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-LQ-004-PL-EN", From: "pl", To: "en", Input: "Są dwa koty.",
		ReferenceTexts: []string{"There are two cats."},
		MustPreserve:   []string{"istnieją koty", "liczba wynosi dokładnie dwa"},
		AmbiguityNote:  "Sama liczebność nie określa, czy angielskie koty są określone; propozycja zachowuje konstrukcję egzystencjalną.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
}
