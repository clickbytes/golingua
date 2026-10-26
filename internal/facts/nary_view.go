package facts

import (
	"fmt"
	"strings"
	"time"

	"github.com/clickbytes/golingua/internal/algebra"
)

// NaryAtomEvidence associates an exact canonical atom with its currently
// applicable explicit evidence.
type NaryAtomEvidence struct {
	Atom     NaryAtom     `json:"atom"`
	Evidence NaryEvidence `json:"evidence"`
}

// NarySummary is a detached request-scoped view of stored and active records.
// ActiveFacts counts records, not distinct atoms.
type NarySummary struct {
	SnapshotID    string             `json:"snapshot_id"`
	ContextID     string             `json:"context_id"`
	At            string             `json:"at,omitempty"`
	StoredFacts   int                `json:"stored_facts"`
	ActiveFacts   int                `json:"active_facts"`
	InactiveFacts int                `json:"inactive_facts"`
	Atoms         []NaryAtomEvidence `json:"atoms"`
	Conflicts     []NaryConflict     `json:"conflicts"`
}

// EvidenceFor returns explicit positive and negative evidence for an exact
// atom at the supplied UTC instant. An empty instant selects timeless facts
// only; it never consults the view's At value or the current clock.
func (view *NaryView) EvidenceFor(atom NaryAtom, at string) (NaryEvidence, error) {
	if err := view.validateActiveView(at); err != nil {
		return NaryEvidence{}, err
	}
	referents := make(map[string]algebra.Declaration, len(view.referents))
	for _, referent := range view.referents {
		referents[referent.ID] = referent
	}
	canonical, err := validateNaryAtom(view.schema, referents, atom, "atom")
	if err != nil {
		return NaryEvidence{}, err
	}
	key := naryAtomKey(canonical)
	result := emptyNaryEvidence()
	var instant time.Time
	if at != "" {
		instant, _ = time.Parse(time.RFC3339Nano, at)
	}
	for _, fact := range view.facts {
		if naryAtomKey(fact.Atom) != key || !naryFactActiveAt(fact.ValidTime, instant, at != "") {
			continue
		}
		switch fact.Polarity {
		case Positive:
			result.Positive = append(result.Positive, cloneNaryFact(fact))
		case Negative:
			result.Negative = append(result.Negative, cloneNaryFact(fact))
		}
	}
	result.State = naryEvidenceState(result.Positive, result.Negative)
	return result, nil
}

// SummaryAt returns all stored atoms in first-occurrence order, including
// atoms whose evidence is inactive at at. No current-time fallback is used.
func (view *NaryView) SummaryAt(at string) (NarySummary, error) {
	if err := view.validateActiveView(at); err != nil {
		return NarySummary{}, err
	}
	if len(view.facts) > DefaultNaryMaxFacts {
		return NarySummary{}, factLimit("facts", len(view.facts))
	}
	var instant time.Time
	if at != "" {
		instant, _ = time.Parse(time.RFC3339Nano, at)
	}
	result := NarySummary{
		SnapshotID: view.snapshotID, ContextID: view.contextID, At: at,
		StoredFacts: len(view.facts), Atoms: make([]NaryAtomEvidence, 0, len(view.facts)),
		Conflicts: make([]NaryConflict, 0),
	}
	positions := make(map[string]int, len(view.facts))
	for _, fact := range view.facts {
		key := naryAtomKey(fact.Atom)
		index, exists := positions[key]
		if !exists {
			index = len(result.Atoms)
			positions[key] = index
			result.Atoms = append(result.Atoms, NaryAtomEvidence{Atom: cloneNaryAtom(fact.Atom), Evidence: emptyNaryEvidence()})
		}
		active := naryFactActiveAt(fact.ValidTime, instant, at != "")
		if !active {
			result.InactiveFacts++
			continue
		}
		result.ActiveFacts++
		evidence := &result.Atoms[index].Evidence
		switch fact.Polarity {
		case Positive:
			evidence.Positive = append(evidence.Positive, cloneNaryFact(fact))
		case Negative:
			evidence.Negative = append(evidence.Negative, cloneNaryFact(fact))
		}
	}
	for index := range result.Atoms {
		evidence := &result.Atoms[index].Evidence
		evidence.State = naryEvidenceState(evidence.Positive, evidence.Negative)
		if evidence.State != NaryBoth {
			continue
		}
		conflict := NaryConflict{Atom: cloneNaryAtom(result.Atoms[index].Atom), PositiveIDs: make([]string, 0, len(evidence.Positive)), NegativeIDs: make([]string, 0, len(evidence.Negative))}
		for _, fact := range evidence.Positive {
			conflict.PositiveIDs = append(conflict.PositiveIDs, fact.ID)
		}
		for _, fact := range evidence.Negative {
			conflict.NegativeIDs = append(conflict.NegativeIDs, fact.ID)
		}
		result.Conflicts = append(result.Conflicts, conflict)
	}
	return result, nil
}

func (view *NaryView) validateActiveView(at string) error {
	if view == nil || !validNaryID(view.snapshotID) || !validNaryID(view.contextID) {
		return factError("INVALID_NARY_VIEW", "view", "n-ary evidence view is nil or uninitialized")
	}
	if len(at) > 128 || !validUTCInstant(at) {
		return factError("INVALID_FACT_QUERY_TIME", "at", "query instant must be a bounded RFC3339 UTC timestamp")
	}
	if len(view.facts) > DefaultNaryMaxFacts || len(view.referents) > DefaultNaryMaxReferents {
		return factLimit("view", len(view.facts)+len(view.referents))
	}
	return nil
}

func emptyNaryEvidence() NaryEvidence {
	return NaryEvidence{State: NaryUnknown, Positive: make([]NaryFact, 0), Negative: make([]NaryFact, 0)}
}

func naryEvidenceState(positive, negative []NaryFact) NaryKnowledge {
	switch {
	case len(positive) > 0 && len(negative) > 0:
		return NaryBoth
	case len(positive) > 0:
		return NaryPositive
	case len(negative) > 0:
		return NaryNegative
	default:
		return NaryUnknown
	}
}

func naryFactActiveAt(interval ValidTime, at time.Time, timed bool) bool {
	if !timed {
		return interval.From == "" && interval.To == ""
	}
	if interval.From != "" {
		start, _ := time.Parse(time.RFC3339Nano, interval.From)
		if at.Before(start) {
			return false
		}
	}
	if interval.To != "" {
		end, _ := time.Parse(time.RFC3339Nano, interval.To)
		if at.After(end) {
			return false
		}
	}
	return true
}

func naryAtomKey(atom NaryAtom) string {
	var key strings.Builder
	writeNaryKeyPart(&key, atom.Symbol)
	for _, argument := range atom.Args {
		writeNaryKeyPart(&key, argument.Role)
		writeNaryKeyPart(&key, string(argument.Value.Kind))
		writeNaryKeyPart(&key, argument.Value.ID)
	}
	return key.String()
}

func writeNaryKeyPart(builder *strings.Builder, value string) {
	fmt.Fprintf(builder, "%d:", len(value))
	builder.WriteString(value)
}

func cloneNaryAtom(atom NaryAtom) NaryAtom {
	atom.Args = append([]NaryArgument{}, atom.Args...)
	return atom
}
