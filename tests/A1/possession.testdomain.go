package a1

// Posiadanie; rodzajniki angielskie nie wynikają z polskiej formy.
var PossessionCases = []Case{
	{
		ID: "A1-PO-001-EN-PL", From: "en", To: "pl", Input: "I have a cat.",
		ReferenceTexts: []string{"Mam kota."},
		MustPreserve:   []string{"mówiący jest posiadaczem", "posiadanym obiektem jest kot", "liczba pojedyncza"},
		AmbiguityNote:  "Czasownik mam wskazuje pierwszą osobę mimo pominiętego zaimka; polski nie koduje angielskiego rodzajnika a, a określoność wymaga kontekstu.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-PO-002-PL-EN", From: "pl", To: "en", Input: "Mam kota.",
		ReferenceTexts: []string{"I have a cat.", "I have the cat."},
		MustPreserve:   []string{"mówiący jest posiadaczem", "posiadanym obiektem jest kot", "liczba pojedyncza"},
		AmbiguityNote:  "Czasownik mam wskazuje pierwszą osobę; angielski a/the wymaga kontekstu, bo polski nie koduje tu określoności.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-PO-003-EN-PL", From: "en", To: "pl", Input: "She has a book.",
		ReferenceTexts: []string{"Ona ma książkę."},
		MustPreserve:   []string{"posiadaczką jest wskazana osoba rodzaju żeńskiego", "posiadanym obiektem jest książka", "liczba pojedyncza"},
		AmbiguityNote:  "Angielski rodzajnik a nie ma obowiązkowego odpowiednika w polszczyźnie.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-PO-004-PL-EN", From: "pl", To: "en", Input: "Ona ma książkę.",
		ReferenceTexts: []string{"She has a book.", "She has the book."},
		MustPreserve:   []string{"posiadaczką jest wskazana osoba rodzaju żeńskiego", "posiadanym obiektem jest książka", "liczba pojedyncza"},
		AmbiguityNote:  "Angielski rodzajnik i określoność zależą od kontekstu; polski przypadek nie określa ich samodzielnie.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
}
