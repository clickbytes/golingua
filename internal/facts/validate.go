package facts

import (
	"fmt"
	"strings"
	"time"

	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
)

func validateContext(ontology *core.ConceptGraph, context meaning.RequestContext) error {
	if ontology == nil {
		return factError("CORE_UNAVAILABLE", "core", "fact validation requires a pinned CORE graph")
	}
	if err := context.Validate(); err != nil {
		return err
	}
	for index, referent := range context.Referents {
		conceptID, err := ids.ParseConceptID(string(referent.Concept))
		if err != nil || conceptID.String() != string(referent.Concept) {
			return factError("INVALID_REFERENT_CONCEPT", fmt.Sprintf("context.referents[%d].concept", index), "referent concept ID must be canonical")
		}
		concept, exists := ontology.Lookup(conceptID)
		if !exists || (concept.Kind != core.KindEntity && concept.Kind != core.KindAbstract) {
			return factError("INVALID_REFERENT_CONCEPT", fmt.Sprintf("context.referents[%d].concept", index), "referent must have an entity or abstract concept in CORE")
		}
	}
	return nil
}

func validateFact(ontology *core.ConceptGraph, context meaning.RequestContext, fact Fact, index int) error {
	path := fmt.Sprintf("facts[%d]", index)
	if !validLocalID(fact.ID) {
		return factError("INVALID_FACT_ID", path+".id", "fact ID must be bounded ASCII")
	}
	predicate, err := core.ParseRelationID(fact.Predicate)
	if err != nil || predicate.String() != fact.Predicate {
		return factError("INVALID_FACT_PREDICATE", path+".predicate", "predicate must be a canonical CORE relation ID")
	}
	relation, exists := ontology.Relation(predicate)
	if !exists {
		return factError("UNKNOWN_FACT_PREDICATE", path+".predicate", "predicate relation is not declared in CORE")
	}
	if fact.Polarity != Positive && fact.Polarity != Negative {
		return factError("INVALID_FACT_POLARITY", path+".polarity", "fact requires explicit positive or negative polarity")
	}
	if fact.AssertionScope != ScopeAsserted {
		return factError("NONASSERTED_FACT", path+".assertion_scope", "only explicitly asserted content can enter a FactSet")
	}
	if err := validateSource(fact.Source, path+".source"); err != nil {
		return err
	}
	if err := validateTime(fact.ValidTime, path+".valid_time"); err != nil {
		return err
	}
	for argumentIndex, id := range fact.ArgReferentIDs {
		var concept core.Concept
		found := false
		for _, referent := range context.Referents {
			if referent.ID == id {
				conceptID, _ := ids.ParseConceptID(string(referent.Concept))
				concept, found = ontology.Lookup(conceptID)
				break
			}
		}
		if !found {
			return factError("UNKNOWN_FACT_REFERENT", fmt.Sprintf("%s.arg_referent_ids[%d]", path, argumentIndex), "fact argument is not a referent in this request")
		}
		allowed := relation.SubjectKind
		if argumentIndex == 1 {
			allowed = relation.ObjectKind
		}
		if !allowsKind(allowed, concept.Kind) {
			return factError("FACT_RELATION_KIND_MISMATCH", fmt.Sprintf("%s.arg_referent_ids[%d]", path, argumentIndex), "referent concept kind violates the relation signature")
		}
	}
	return nil
}

func validateSource(source SourceEvidence, path string) error {
	if !validLocalID(source.ID) {
		return factError("INVALID_FACT_SOURCE", path+".id", "source requires a bounded explicit evidence ID")
	}
	switch source.Kind {
	case SourceUser:
		if source.Reference != nil {
			return factError("INVALID_FACT_SOURCE", path+".reference", "user evidence cannot claim an external source reference")
		}
	case SourceDocument:
		if source.Reference == nil {
			return factError("MISSING_DOCUMENT_SOURCE", path+".reference", "document evidence requires a source reference")
		}
		if err := source.Reference.Validate(); err != nil {
			return factError("INVALID_DOCUMENT_SOURCE", path+".reference", err.Error())
		}
	default:
		return factError("INVALID_FACT_SOURCE", path+".kind", "source kind is not supported")
	}
	return nil
}

func validateTime(interval ValidTime, path string) error {
	var from, to time.Time
	for index, value := range []string{interval.From, interval.To} {
		if value == "" {
			continue
		}
		parsed, err := time.Parse(time.RFC3339Nano, value)
		if err != nil || !strings.HasSuffix(value, "Z") {
			return factError("INVALID_FACT_VALID_TIME", path, "valid time endpoints must use RFC3339 UTC timestamps")
		}
		if index == 0 {
			from = parsed
		} else {
			to = parsed
		}
	}
	if !from.IsZero() && !to.IsZero() && from.After(to) {
		return factError("INVALID_FACT_VALID_TIME", path, "valid time starts after it ends")
	}
	return nil
}

// AssertedMeaning is a guarded input boundary for a later explicit mapping
// from utterance assertions to relation facts. It never creates facts itself.
// Any unresolved, scoped, or nonassertive content rejects the whole import.
func AssertedMeaning(input meaning.Meaning, ontology *core.ConceptGraph, budget meaning.Budget) ([]meaning.Assertion, error) {
	if err := meaning.ValidateMeaning(input, ontology, budget); err != nil {
		return nil, err
	}
	projection, err := meaning.ProjectAsserted(input, budget)
	if err != nil {
		return nil, err
	}
	if len(projection.Unresolved) != 0 || len(projection.Asserted) == 0 {
		return nil, factError("NONASSERTED_MEANING", "meaning", "meaning contains question, negative, hypothetical, modal, or otherwise nonasserted content")
	}
	result := make([]meaning.Assertion, len(projection.Asserted))
	for index, assertion := range projection.Asserted {
		result[index] = assertion
		result[index].Participants = append([]meaning.Participant(nil), assertion.Participants...)
		result[index].SourcePath = append([]meaning.NodeID(nil), assertion.SourcePath...)
	}
	return result, nil
}

func allowsKind(allowed []core.ConceptKind, actual core.ConceptKind) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, kind := range allowed {
		if kind == actual {
			return true
		}
	}
	return false
}

func validLocalID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if !((character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') || (character >= '0' && character <= '9') || character == '_' || character == '-' || character == '.') {
			return false
		}
	}
	return true
}

type StatusError struct{ result status.Result }

func (failure *StatusError) Error() string {
	if failure == nil || len(failure.result.Diagnostics) == 0 {
		return "fact set operation failed"
	}
	return failure.result.Diagnostics[0].Message
}
func (failure *StatusError) Result() status.Result {
	if failure == nil {
		return status.Result{}
	}
	result := failure.result
	result.Diagnostics = append([]status.Diagnostic(nil), result.Diagnostics...)
	return result
}

func factError(code, path, message string) error {
	return &StatusError{result: status.Result{Status: status.OutcomeInvalid, Diagnostics: []status.Diagnostic{{Code: code, Path: path, Message: message, Severity: status.SeverityError}}}}
}

func factLimit(path string, count int) error {
	return &StatusError{result: status.Result{Status: status.OutcomeResourceLimit, Diagnostics: []status.Diagnostic{{Code: "FACT_RESOURCE_LIMIT", Path: path, Message: fmt.Sprintf("fact request exceeds its bound (%d items)", count), Severity: status.SeverityError}}}}
}
