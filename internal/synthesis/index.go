package synthesis

import (
	"sort"

	"github.com/clickbytes/golingua/internal/algebra"
)

// SignatureChoice identifies one exact role assignment for a primitive
// signature. It contains IDs only: labels and symbol spellings are not used
// to infer a meaning or preferred orientation.
type SignatureChoice struct {
	Symbol    string `json:"symbol"`
	FromRole  string `json:"from_role,omitempty"`
	ToRole    string `json:"to_role,omitempty"`
	UnaryRole string `json:"unary_role,omitempty"`
}

// SignatureIndex is a detached, immutable view of declaration signatures.
// Callers decide which selected metadata is approved for this input before
// constructing the index.
type SignatureIndex struct {
	signatures []algebra.Signature
}

// NewSignatureIndex validates a declaration-only context before copying any
// caller-owned type graph. The caller is responsible for selecting the
// approved primitive inventory from its trusted metadata snapshot.
func NewSignatureIndex(context algebra.Document) (*SignatureIndex, error) {
	if context.Expression != nil || len(context.Parameters) != 0 || len(context.References) != 0 {
		return nil, goalInvalid("INVALID_CONTEXT", "context", "signature index requires declarations only")
	}
	if err := algebra.CheckDeclarations(context); err != nil {
		return nil, err
	}

	index := &SignatureIndex{signatures: make([]algebra.Signature, len(context.Signatures))}
	for i, signature := range context.Signatures {
		copySignature := algebra.Signature{
			ID: signature.ID, Label: signature.Label,
			Ports: make([]algebra.Port, len(signature.Ports)),
		}
		for j, port := range signature.Ports {
			copySignature.Ports[j] = algebra.Port{Role: port.Role, Type: cloneGoalType(port.Type)}
		}
		index.signatures[i] = copySignature
	}
	sort.Slice(index.signatures, func(i, j int) bool {
		return index.signatures[i].ID < index.signatures[j].ID
	})
	return index, nil
}

// Lookup returns a detached signature copy. Mutating the result cannot alter
// future lookup or choice results from the index.
func (index *SignatureIndex) Lookup(symbol string) (algebra.Signature, bool) {
	if index == nil {
		return algebra.Signature{}, false
	}
	position := sort.Search(len(index.signatures), func(i int) bool {
		return index.signatures[i].ID >= symbol
	})
	if position == len(index.signatures) || index.signatures[position].ID != symbol {
		return algebra.Signature{}, false
	}
	return cloneIndexSignature(index.signatures[position]), true
}

// BinaryChoices returns every exact role assignment from a binary signature
// whose port types match fromType and toType. A non-empty role filter is exact;
// empty role filters enumerate all legal orientations. Results are stable by
// symbol and role IDs, independent of input declaration order.
func (index *SignatureIndex) BinaryChoices(fromType, toType algebra.Type, fromRole, toRole string) []SignatureChoice {
	if index == nil {
		return []SignatureChoice{}
	}
	choices := make([]SignatureChoice, 0)
	for _, signature := range index.signatures {
		if len(signature.Ports) != 2 {
			continue
		}
		forward := signature.Ports[0]
		reverse := signature.Ports[1]
		if binaryRoleAssignmentMatches(forward, reverse, fromType, toType, fromRole, toRole) {
			choices = append(choices, SignatureChoice{Symbol: signature.ID, FromRole: forward.Role, ToRole: reverse.Role})
		}
		if binaryRoleAssignmentMatches(reverse, forward, fromType, toType, fromRole, toRole) {
			choices = append(choices, SignatureChoice{Symbol: signature.ID, FromRole: reverse.Role, ToRole: forward.Role})
		}
	}
	sort.Slice(choices, func(i, j int) bool {
		if choices[i].Symbol != choices[j].Symbol {
			return choices[i].Symbol < choices[j].Symbol
		}
		if choices[i].FromRole != choices[j].FromRole {
			return choices[i].FromRole < choices[j].FromRole
		}
		return choices[i].ToRole < choices[j].ToRole
	})
	return choices
}

// UnaryChoices returns all exact one-port signatures with the requested type,
// in stable symbol/role order.
func (index *SignatureIndex) UnaryChoices(expectedType algebra.Type) []SignatureChoice {
	if index == nil {
		return []SignatureChoice{}
	}
	choices := make([]SignatureChoice, 0)
	for _, signature := range index.signatures {
		if len(signature.Ports) == 1 && algebra.TypeEqual(signature.Ports[0].Type, expectedType) {
			choices = append(choices, SignatureChoice{Symbol: signature.ID, UnaryRole: signature.Ports[0].Role})
		}
	}
	return choices
}

func binaryRoleAssignmentMatches(from, to algebra.Port, fromType, toType algebra.Type, fromRole, toRole string) bool {
	return (fromRole == "" || from.Role == fromRole) &&
		(toRole == "" || to.Role == toRole) &&
		algebra.TypeEqual(from.Type, fromType) && algebra.TypeEqual(to.Type, toType)
}

func cloneIndexSignature(signature algebra.Signature) algebra.Signature {
	copySignature := algebra.Signature{
		ID: signature.ID, Label: signature.Label,
		Ports: make([]algebra.Port, len(signature.Ports)),
	}
	for i, port := range signature.Ports {
		copySignature.Ports[i] = algebra.Port{Role: port.Role, Type: cloneGoalType(port.Type)}
	}
	return copySignature
}
