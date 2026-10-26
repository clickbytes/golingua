package facts

import (
	"fmt"
	"strings"
	"time"

	"github.com/clickbytes/golingua/internal/algebra"
)

// BuildNary validates an explicit request against the selected declaration
// snapshot and returns an immutable, request-local view. It never changes a
// registry or any global knowledge state.
func BuildNary(schema NarySchema, input NaryInput) (*NaryView, error) {
	if input.SchemaVersion != SchemaVersion {
		return nil, factError("INVALID_FACT_SCHEMA_VERSION", "schema_version", "unsupported n-ary fact input schema version")
	}
	if !validNaryID(schema.SnapshotID) {
		return nil, factError("INVALID_FACT_SNAPSHOT", "schema.snapshot_id", "selected schema requires a bounded snapshot ID")
	}
	if input.SnapshotID != "" && (!validNaryID(input.SnapshotID) || input.SnapshotID != schema.SnapshotID) {
		return nil, factError("FACT_SNAPSHOT_MISMATCH", "snapshot_id", "request snapshot pin does not match the selected schema")
	}
	if !validNaryID(input.ContextID) {
		return nil, factError("INVALID_FACT_CONTEXT", "context_id", "context ID must be bounded valid UTF-8 text")
	}
	if len(input.At) > 128 || !validUTCInstant(input.At) {
		return nil, factError("INVALID_FACT_QUERY_TIME", "at", "query instant must be a bounded RFC3339 UTC timestamp")
	}
	limits, err := NormalizeNaryLimits(input.Limits)
	if err != nil {
		return nil, err
	}
	if len(input.Facts) > limits.MaxFacts || len(input.Referents) > limits.MaxReferents {
		return nil, factLimit("facts_or_referents", len(input.Facts)+len(input.Referents))
	}
	if len(schema.Document.Sorts)+len(schema.Document.Signatures)+len(input.Referents) > algebra.DefaultMaxSymbols {
		return nil, factLimit("schema_or_referents", len(schema.Document.Sorts)+len(schema.Document.Signatures)+len(input.Referents))
	}
	if schema.Document.SchemaVersion != algebra.SchemaVersion || schema.Document.Expression != nil || len(schema.Document.Parameters) != 0 || len(schema.Document.References) != 0 {
		return nil, factError("INVALID_FACT_SCHEMA_DOCUMENT", "schema.document", "schema must contain declarations only, without expression, parameters, or references")
	}
	for index, referent := range input.Referents {
		if !validNaryID(referent.ID) {
			return nil, factError("INVALID_REFERENT_ID", fmt.Sprintf("referents[%d].id", index), "referent ID must be bounded valid UTF-8 text")
		}
	}
	// Validate original type graphs before copying or allocating lookup maps.
	declarations := schema.Document
	declarations.References = input.Referents
	if err := algebra.CheckDeclarations(declarations); err != nil {
		return nil, err
	}
	for index, referent := range input.Referents {
		if !isGroundReferentType(referent.Type) {
			return nil, factError("INVALID_FACT_REFERENT_TYPE", fmt.Sprintf("referents[%d].type", index), "ground referents require Entity, Event, Interval, or a declared Nominal sort")
		}
	}
	for index, signature := range schema.Document.Signatures {
		if len(signature.Ports) > algebra.DefaultMaxPorts {
			return nil, factLimit(fmt.Sprintf("schema.document.signatures[%d].ports", index), len(signature.Ports))
		}
	}

	referents := make(map[string]algebra.Declaration, len(input.Referents))
	for _, referent := range input.Referents {
		referents[referent.ID] = referent
	}
	canonical := make([]NaryFact, len(input.Facts))
	factIDs := make(map[string]struct{}, len(input.Facts))
	for index, fact := range input.Facts {
		path := fmt.Sprintf("facts[%d]", index)
		if !validLocalID(fact.ID) {
			return nil, factError("INVALID_FACT_ID", path+".id", "fact ID must be a bounded explicit identifier")
		}
		if _, exists := factIDs[fact.ID]; exists {
			return nil, factError("DUPLICATE_FACT_ID", path+".id", "fact ID is duplicated")
		}
		factIDs[fact.ID] = struct{}{}
		if fact.Polarity != Positive && fact.Polarity != Negative {
			return nil, factError("INVALID_FACT_POLARITY", path+".polarity", "fact requires explicit positive or negative polarity")
		}
		if fact.AssertionScope != ScopeAsserted {
			return nil, factError("NONASSERTED_FACT", path+".assertion_scope", "only explicitly asserted content can enter a fact view")
		}
		if err := validateSource(fact.Source, path+".source"); err != nil {
			return nil, err
		}
		if err := validateNaryValidTime(fact.ValidTime, path+".valid_time"); err != nil {
			return nil, err
		}
		atom, err := validateNaryAtom(schema.Document, referents, fact.Atom, path+".atom")
		if err != nil {
			return nil, err
		}
		fact.Atom = atom
		canonical[index] = fact
	}
	return newValidatedNaryView(schema, input, canonical)
}

// validateNaryAtom validates a role map and returns arguments in the exact
// declared port order. Equal types do not make distinct roles interchangeable.
func validateNaryAtom(document algebra.Document, referents map[string]algebra.Declaration, atom NaryAtom, path string) (NaryAtom, error) {
	if !validNaryID(atom.Symbol) {
		return NaryAtom{}, factError("INVALID_FACT_SYMBOL", path+".symbol", "relation symbol must be a bounded explicit identifier")
	}
	var signature *algebra.Signature
	for index := range document.Signatures {
		if document.Signatures[index].ID == atom.Symbol {
			signature = &document.Signatures[index]
			break
		}
	}
	if signature == nil {
		return NaryAtom{}, factError("UNKNOWN_FACT_SYMBOL", path+".symbol", "relation symbol is not declared in the selected schema")
	}
	if len(atom.Args) > algebra.DefaultMaxPorts || len(atom.Args) != len(signature.Ports) {
		return NaryAtom{}, factError("FACT_ARITY_MISMATCH", path+".args", "fact arguments must match the declared relation arity")
	}
	provided := make(map[string]NaryArgument, len(atom.Args))
	declaredRoles := make(map[string]struct{}, len(signature.Ports))
	for _, port := range signature.Ports {
		declaredRoles[port.Role] = struct{}{}
	}
	for index, argument := range atom.Args {
		argumentPath := fmt.Sprintf("%s.args[%d]", path, index)
		if !validNaryID(argument.Role) {
			return NaryAtom{}, factError("INVALID_FACT_ROLE", argumentPath+".role", "role must be a bounded explicit identifier")
		}
		if _, exists := declaredRoles[argument.Role]; !exists {
			return NaryAtom{}, factError("UNKNOWN_FACT_ROLE", argumentPath+".role", "role is not declared by the selected relation")
		}
		if _, exists := provided[argument.Role]; exists {
			return NaryAtom{}, factError("DUPLICATE_FACT_ROLE", argumentPath+".role", "relation role is supplied more than once")
		}
		provided[argument.Role] = argument
	}
	canonical := NaryAtom{Symbol: atom.Symbol, Args: make([]NaryArgument, 0, len(signature.Ports))}
	for _, port := range signature.Ports {
		argument, exists := provided[port.Role]
		if !exists {
			return NaryAtom{}, factError("MISSING_FACT_ROLE", fmt.Sprintf("%s.args", path), "fact omits a declared relation role: "+port.Role)
		}
		delete(provided, port.Role)
		argumentPath := fmt.Sprintf("%s.args[role=%s]", path, port.Role)
		if argument.Value.Kind != algebra.TermRef {
			return NaryAtom{}, factError("UNGROUNDED_FACT_VALUE", argumentPath+".value.kind", "ground facts require Ref values; Var, Param, and Hole are not ground")
		}
		if !validNaryID(argument.Value.ID) {
			return NaryAtom{}, factError("INVALID_FACT_REFERENT", argumentPath+".value.id", "referent ID must be a bounded explicit identifier")
		}
		referent, exists := referents[argument.Value.ID]
		if !exists {
			return NaryAtom{}, factError("UNKNOWN_FACT_REFERENT", argumentPath+".value.id", "fact argument is not declared in this request")
		}
		if !algebra.TypeEqual(referent.Type, port.Type) {
			return NaryAtom{}, factError("FACT_ROLE_TYPE_MISMATCH", argumentPath+".value.id", "referent type does not match declared role: "+port.Role)
		}
		canonical.Args = append(canonical.Args, argument)
	}
	return canonical, nil
}

func isGroundReferentType(value algebra.Type) bool {
	switch value.Kind {
	case algebra.TypeEntity, algebra.TypeEvent, algebra.TypeInterval, algebra.TypeNominal:
		return true
	default:
		return false
	}
}

func validateNaryValidTime(interval ValidTime, path string) error {
	if len(interval.From) > 128 || len(interval.To) > 128 {
		return factError("INVALID_FACT_VALID_TIME", path, "valid-time endpoints exceed their bound")
	}
	if err := validateTime(interval, path); err != nil {
		return err
	}
	if interval.From != "" && interval.To != "" {
		from, _ := time.Parse(time.RFC3339Nano, interval.From)
		to, _ := time.Parse(time.RFC3339Nano, interval.To)
		if from.After(to) {
			return factError("INVALID_FACT_VALID_TIME", path, "valid time starts after it ends")
		}
	}
	return nil
}

func validUTCInstant(value string) bool {
	if value == "" {
		return true
	}
	if !strings.HasSuffix(value, "Z") {
		return false
	}
	_, err := time.Parse(time.RFC3339Nano, value)
	return err == nil
}
