package a1

// Identyfikacja i opis; rodzajnik lub wskazanie zależą od kontekstu.
var IdentificationCases = []Case{
	{
		ID: "A1-ID-001-EN-PL", From: "en", To: "pl", Input: "This is a cat",
		ReferenceTexts: []string{"To jest kot."},
		MustPreserve:   []string{"wypowiedź identyfikuje obiekt jako kota"},
		AmbiguityNote:  "Propozycja zależy od kontekstu dla this; przekład w drugą stronę może brzmieć This is a cat, That is a cat lub It is a cat. Polski nie koduje tu rodzajnika.",
		InputOrigin:    "owner_provided", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-ID-002-PL-EN", From: "pl", To: "en", Input: "To jest kot.",
		ReferenceTexts: []string{"This is a cat.", "That is a cat.", "It is a cat."},
		MustPreserve:   []string{"wypowiedź identyfikuje obiekt jako kota"},
		AmbiguityNote:  "To może wskazywać obiekt bliski lub odległy albo pełnić rolę zaimka; angielski rodzajnik i wybór this/that/it wymagają kontekstu.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-ID-003-EN-PL", From: "en", To: "pl", Input: "This cat is black.",
		ReferenceTexts: []string{"Ten kot jest czarny."},
		MustPreserve:   []string{"wskazany kot ma cechę czarnego koloru", "wypowiedź dotyczy wskazanego kota, nie wszystkich kotów"},
		AmbiguityNote:  "Angielskie cat nie określa płci; polski rodzaj gramatyczny kot nie jest twierdzeniem o płci.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
	{
		ID: "A1-ID-004-PL-EN", From: "pl", To: "en", Input: "Ten kot jest czarny.",
		ReferenceTexts: []string{"This cat is black."},
		MustPreserve:   []string{"wskazany kot ma cechę czarnego koloru", "wypowiedź dotyczy wskazanego kota, nie wszystkich kotów"},
		AmbiguityNote:  "Należy zachować wskazanie wyrażone przez ten; polski rodzaj gramatyczny kot nie jest twierdzeniem o płci.",
		InputOrigin:    "agent_proposal", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED",
	},
}
