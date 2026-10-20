package meaning

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/ids"
)

const (
	MaxReferenceIDBytes = 64
	MaxContextIDBytes   = 64
	MaxMentions         = 4096
	MaxReferents        = 4096
	MaxAnchorBytes      = 256
	MaxSurfaceBytes     = 1 << 20
)

// ReferentID identifies one entity instance inside a single request context.
// Equal labels in different contexts never identify the same entity.
type ReferentID string

// MentionID identifies one referring expression within a request context.
type MentionID string

// DeicticAnchor records bounded explicit context for a deictic mention.
type DeicticAnchor struct {
	Kind  string `json:"kind"`
	Value string `json:"value"`
}

// Referent is a request-local entity instance. Concept describes its class;
// it does not establish identity with another referent.
type Referent struct {
	ID      ReferentID `json:"id"`
	Concept ConceptRef `json:"concept"`
}

// Mention records a referring expression and an optional evidence-backed
// resolution. An empty ReferentID means unresolved, never an absent entity.
type Mention struct {
	ID         MentionID      `json:"id"`
	NodeID     NodeID         `json:"node_id"`
	ReferentID ReferentID     `json:"referent_id,omitempty"`
	EvidenceID string         `json:"evidence_id,omitempty"`
	Anchor     *DeicticAnchor `json:"anchor,omitempty"`
}

// RequestContext contains identity and mention evidence for one request only.
// It must not be persisted as conversational memory.
type RequestContext struct {
	ID        string     `json:"id"`
	Referents []Referent `json:"referents"`
	Mentions  []Mention  `json:"mentions"`
}

func validOpaqueID(value string, maximum int) bool {
	if len(value) == 0 || len(value) > maximum || !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' || r == '.') {
			return false
		}
	}
	return true
}

// Validate checks context-local identity, mention links and explicit anchors.
func (context RequestContext) Validate() error {
	if !validOpaqueID(context.ID, MaxContextIDBytes) {
		return invalid("INVALID_CONTEXT_ID", "context.id", "context ID is malformed or exceeds its bound")
	}
	if len(context.Referents) > MaxReferents || len(context.Mentions) > MaxMentions {
		return resourceLimit("references", len(context.Referents)+len(context.Mentions), MaxReferents+MaxMentions)
	}
	referents := make(map[ReferentID]struct{}, len(context.Referents))
	for index, referent := range context.Referents {
		path := fmt.Sprintf("context.referents[%d]", index)
		if !validOpaqueID(string(referent.ID), MaxReferenceIDBytes) {
			return invalid("INVALID_REFERENT_ID", path+".id", "referent ID is malformed or exceeds its bound")
		}
		if _, exists := referents[referent.ID]; exists {
			return invalid("DUPLICATE_REFERENT_ID", path+".id", "referent ID is duplicated")
		}
		if _, err := parseConceptRef(referent.Concept); err != nil {
			return invalid("INVALID_CONCEPT_ID", path+".concept", "referent concept ID is malformed")
		}
		referents[referent.ID] = struct{}{}
	}
	mentions := make(map[MentionID]struct{}, len(context.Mentions))
	for index, mention := range context.Mentions {
		path := fmt.Sprintf("context.mentions[%d]", index)
		if !validOpaqueID(string(mention.ID), MaxReferenceIDBytes) || !validNodeID(mention.NodeID) {
			return invalid("INVALID_MENTION", path, "mention ID or node ID is malformed")
		}
		if _, exists := mentions[mention.ID]; exists {
			return invalid("DUPLICATE_MENTION_ID", path+".id", "mention ID is duplicated")
		}
		mentions[mention.ID] = struct{}{}
		if mention.ReferentID != "" {
			if _, exists := referents[mention.ReferentID]; !exists {
				return unknown("UNKNOWN_REFERENT", path+".referent_id", "mention refers to no referent in this request context")
			}
			if !validOpaqueID(mention.EvidenceID, MaxReferenceIDBytes) {
				return invalid("MISSING_IDENTITY_EVIDENCE", path+".evidence_id", "resolved mention requires an explicit bounded evidence ID")
			}
		} else if mention.EvidenceID != "" {
			return invalid("UNBOUND_IDENTITY_EVIDENCE", path+".evidence_id", "unresolved mention cannot carry identity evidence")
		}
		if mention.Anchor != nil {
			anchor := mention.Anchor
			if !validOpaqueID(anchor.Kind, 32) || len(anchor.Value) == 0 || len(anchor.Value) > MaxAnchorBytes || !utf8.ValidString(anchor.Value) || strings.TrimSpace(anchor.Value) != anchor.Value {
				return invalid("INVALID_DEICTIC_ANCHOR", path+".anchor", "deictic anchor must have a bounded kind and non-empty trimmed value")
			}
		}
	}
	return nil
}

func parseConceptRef(reference ConceptRef) (ids.ConceptID, error) {
	return ids.ParseConceptID(string(reference))
}

// SurfaceFingerprint returns a stable digest of one explicitly supplied
// surface form. It stores no text and does not normalize spelling or order.
func SurfaceFingerprint(surface string) (string, error) {
	if !utf8.ValidString(surface) {
		return "", invalid("INVALID_SURFACE", "surface", "surface text is not valid UTF-8")
	}
	if len(surface) == 0 {
		return "", invalid("INVALID_SURFACE", "surface", "surface text must not be empty")
	}
	if len(surface) > MaxSurfaceBytes {
		return "", resourceLimit("surface_bytes", len(surface), MaxSurfaceBytes)
	}
	hash := sha256.Sum256([]byte(surface))
	return hex.EncodeToString(hash[:]), nil
}
