package a1

// Proste działanie widzenia; rekord S001 ma obserwację w korpusie.
var ActionCases = []Case{
	{
		ID: "A1-AC-001-EN-PL", From: "en", To: "pl", Input: "The teacher sees the dog.",
		ReferenceTexts: []string{"Nauczyciel widzi psa."},
		MustPreserve:   []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdanie twierdzi, że widzenie zachodzi"},
		AmbiguityNote:  "Polskie końcówki odróżniają role; angielskie the nie ma osobnego odpowiednika.",
		InputOrigin:    "data/sentences/observed-2026-10-23-g10-14-full-corpus.tsv#S001", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-AC-002-PL-EN", From: "pl", To: "en", Input: "Nauczyciel widzi psa.",
		ReferenceTexts: []string{"The teacher sees the dog.", "The teacher sees a dog.", "A teacher sees the dog.", "A teacher sees a dog."},
		MustPreserve:   []string{"nauczyciel jest doświadczającym widzenia (experiencer)", "pies jest bodźcem widzenia (stimulus)", "zdanie twierdzi, że widzenie zachodzi"},
		AmbiguityNote:  "Polszczyzna nie koduje angielskiej określoności; dobór rodzajników zależy od kontekstu.",
		InputOrigin:    "data/sentences/observed-2026-10-23-g10-14-full-corpus.tsv#S001", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-AC-003-EN-PL", From: "en", To: "pl", Input: "The dog sees the teacher.",
		ReferenceTexts: []string{"Pies widzi nauczyciela."},
		MustPreserve:   []string{"pies jest doświadczającym widzenia (experiencer)", "nauczyciel jest bodźcem widzenia (stimulus)", "role uczestników są odwrócone względem przypadku bazowego"},
		AmbiguityNote:  "Propozycja zachowuje odwrócenie ról semantycznych; polskie formy przypadków powinny je odzwierciedlać.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-AC-004-PL-EN", From: "pl", To: "en", Input: "Pies widzi nauczyciela.",
		ReferenceTexts: []string{"The dog sees the teacher.", "A dog sees the teacher.", "The dog sees a teacher.", "A dog sees a teacher."},
		MustPreserve:   []string{"pies jest doświadczającym widzenia (experiencer)", "nauczyciel jest bodźcem widzenia (stimulus)", "role uczestników są odwrócone względem przypadku bazowego"},
		AmbiguityNote:  "Rodzajniki pozostają nieokreślone przez polski; kolejność i role nie mogą zamienić wykonawcy z obiektem.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
}
