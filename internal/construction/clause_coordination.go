package construction

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/provenance"
)

const ClauseCoordinationSchemaVersion = 1

// ClauseCoordinationPlan is a small language-local plan for joining two
// complete statement clauses after each has independently passed realization.
type ClauseCoordinationPlan struct {
	SchemaVersion int                  `json:"schema_version"`
	ID            string               `json:"id"`
	Language      features.LanguageID  `json:"language"`
	Connector     string               `json:"connector"`
	LowerSecond   bool                 `json:"lower_second_initial"`
	FinalPeriod   string               `json:"final_period"`
	Provenance    provenance.SourceRef `json:"provenance"`
}

// DecodeClauseCoordinationPlan strictly decodes a single bounded language plan.
func DecodeClauseCoordinationPlan(data []byte, language features.LanguageID) (ClauseCoordinationPlan, error) {
	if len(data) == 0 || len(data) > 16*1024 {
		return ClauseCoordinationPlan{}, fmt.Errorf("invalid coordination plan size")
	}
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return ClauseCoordinationPlan{}, fmt.Errorf("invalid coordination JSON keys: %w", err)
	}
	var plan ClauseCoordinationPlan
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&plan); err != nil {
		return ClauseCoordinationPlan{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return ClauseCoordinationPlan{}, fmt.Errorf("coordination plan must contain one JSON value")
	}
	if plan.SchemaVersion != ClauseCoordinationSchemaVersion {
		return ClauseCoordinationPlan{}, fmt.Errorf("unsupported coordination schema version")
	}
	constructionID, idErr := ids.ParseConstructionID(plan.ID)
	if idErr != nil || constructionID.String() != plan.ID || plan.Language != language || plan.Connector == "" || strings.TrimSpace(plan.Connector) != plan.Connector || len(plan.Connector) > 32 || !utf8.ValidString(plan.Connector) || plan.FinalPeriod != "." {
		return ClauseCoordinationPlan{}, fmt.Errorf("invalid coordination plan fields")
	}
	if err := plan.Provenance.Validate(); err != nil {
		return ClauseCoordinationPlan{}, fmt.Errorf("invalid coordination provenance: %w", err)
	}
	return plan, nil
}
