package a1en

import "github.com/clickbytes/golingua/tests/matrix"

var IdentificationAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-ID-001-EN-PL:analyze", MeaningID: "", Input: "This is a cat", MustPreserve: []string{"wypowiedź identyfikuje obiekt jako kota"}, AmbiguityNote: "Propozycja zależy od kontekstu dla this; przekład w drugą stronę może brzmieć This is a cat, That is a cat lub It is a cat. Polski nie koduje tu rodzajnika.", InputOrigin: "owner_provided", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-ID-003-EN-PL:analyze", MeaningID: "", Input: "This cat is black.", MustPreserve: []string{"wskazany kot ma cechę czarnego koloru", "wypowiedź dotyczy wskazanego kota, nie wszystkich kotów"}, AmbiguityNote: "Angielskie cat nie określa płci; polski rodzaj gramatyczny kot nie jest twierdzeniem o płci.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var IdentificationRealizationCases = []matrix.RealizationCase{
	{ID: "A1-ID-002-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"This is a cat.", "That is a cat.", "It is a cat."}, AmbiguityNote: "To może wskazywać obiekt bliski lub odległy albo pełnić rolę zaimka; angielski rodzajnik i wybór this/that/it wymagają kontekstu.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-ID-004-PL-EN:realize", MeaningID: "", ProposalTexts: []string{"This cat is black."}, AmbiguityNote: "Należy zachować wskazanie wyrażone przez ten; polski rodzaj gramatyczny kot nie jest twierdzeniem o płci.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
