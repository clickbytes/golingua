// Package a1 przechowuje małą, jawną próbkę zdań dla poziomu A1.
// Próbka nie certyfikuje poziomu CEFR ani pełnego pokrycia A1.
package a1

// Case opisuje wejście i propozycje tekstów docelowych do ręcznej oceny.
// ReferenceTexts są niepełnymi propozycjami agenta, a nie asercjami.
type Case struct {
	ID              string
	From            string
	To              string
	Input           string
	ReferenceTexts  []string
	MustPreserve    []string
	AmbiguityNote   string
	InputOrigin     string
	ReferenceOrigin string
	OwnerAcceptance string
}
