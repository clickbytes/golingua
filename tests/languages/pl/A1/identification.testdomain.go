package a1pl

import "github.com/clickbytes/golingua/tests/matrix"

var IdentificationAnalysisCases = []matrix.AnalysisCase{
	{ID: "A1-ID-002-PL-EN:analyze", MeaningID: "", Input: "To jest kot.", MustPreserve: []string{"wypowiedź identyfikuje obiekt jako kota"}, AmbiguityNote: "To może wskazywać obiekt bliski lub odległy albo pełnić rolę zaimka; angielski rodzajnik i wybór this/that/it wymagają kontekstu.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-ID-004-PL-EN:analyze", MeaningID: "", Input: "Ten kot jest czarny.", MustPreserve: []string{"wskazany kot ma cechę czarnego koloru", "wypowiedź dotyczy wskazanego kota, nie wszystkich kotów"}, AmbiguityNote: "Należy zachować wskazanie wyrażone przez ten; polski rodzaj gramatyczny kot nie jest twierdzeniem o płci.", InputOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}

var IdentificationRealizationCases = []matrix.RealizationCase{
	{ID: "A1-ID-001-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"To jest kot."}, AmbiguityNote: "Propozycja zależy od kontekstu dla this; przekład w drugą stronę może brzmieć This is a cat, That is a cat lub It is a cat. Polski nie koduje tu rodzajnika.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
	{ID: "A1-ID-003-EN-PL:realize", MeaningID: "", ProposalTexts: []string{"Ten kot jest czarny."}, AmbiguityNote: "Angielskie cat nie określa płci; polski rodzaj gramatyczny kot nie jest twierdzeniem o płci.", ReferenceOrigin: "agent_proposal", OwnerAcceptance: "NOT_EVALUATED"},
}
