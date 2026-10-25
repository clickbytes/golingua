package translation_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/meaning"
	"github.com/clickbytes/golingua/internal/status"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestAmbiguityAssumptionsFollowRoles(t *testing.T) {
	runtime := matrix.Runtime(t)
	result, err := runtime.Translate(context.Background(), engine.Request{From: features.LanguageID("pl"), To: features.LanguageID("en"), Text: "Nauczyciel widzi psa."})
	encoded, _ := json.Marshal(result)
	sha := os.Getenv("GOLINGUA_MATRIX_SHA")
	if sha == "" {
		sha = "UNRECORDED"
	}
	t.Logf("SHA=%s check_class=ambiguity_assumptions input=%q result=%s error=%v", sha, "Nauczyciel widzi psa.", encoded, err)
	if err != nil {
		t.Fatalf("GAP_ANALYZE: integrated PL source analysis failed: %v", err)
	}
	if result.Status == status.OutcomeUnsupported || len(result.Candidates) == 0 {
		t.Fatalf("%s: no translation candidates (status=%s)", ambiguityGap(result), result.Status)
	}
	want, ok, err := interlingua.Seeing("seeing.teacher-dog")
	if err != nil || !ok {
		t.Fatalf("INFRA: load authored Seeing fixture: %v", err)
	}
	texts := map[string][2]string{
		"The teacher sees the dog.": {"definite", "definite"},
		"The teacher sees a dog.":   {"definite", "indefinite"},
		"A teacher sees the dog.":   {"indefinite", "definite"},
		"A teacher sees a dog.":     {"indefinite", "indefinite"},
	}
	seen := map[string]bool{}
	for _, candidate := range result.Candidates {
		if candidate.Meaning == nil {
			t.Errorf("FAIL: candidate %q has no Meaning", candidate.ID)
			continue
		}
		matches, compareErr := matrix.GraphMatches(want, *candidate.Meaning)
		if compareErr != nil || !matches {
			t.Errorf("FAIL: candidate %q differs from independent Seeing fixture (matches=%t error=%v)", candidate.ID, matches, compareErr)
			continue
		}
		roleNodes, roleErr := seeingRoleNodes(*candidate.Meaning)
		if roleErr != nil {
			t.Errorf("FAIL: candidate %q role mapping: %v", candidate.ID, roleErr)
			continue
		}
		expected, licensed := texts[candidate.Text]
		if !licensed {
			t.Errorf("FAIL: unlicensed or absent English surface %q", candidate.Text)
			continue
		}
		seen[candidate.Text] = true
		for i, role := range []string{"experiencer", "stimulus"} {
			found := false
			for _, assumption := range candidate.Assumptions {
				if assumption.Field == "determiner" && assumption.Value == expected[i] && hasRolePath(assumption.Path, roleNodes["event"], roleNodes[role]) {
					found = true
				}
			}
			if !found {
				t.Errorf("FAIL: surface %q lacks determiner=%s assumption on %s node %s", candidate.Text, expected[i], role, roleNodes[role])
			}
		}
	}
	for text := range texts {
		if !seen[text] {
			t.Errorf("FAIL: missing article alternative %q", text)
		}
	}
}

func ambiguityGap(result engine.TranslationResult) string {
	if len(result.Candidates) == 0 {
		return "GAP_ANALYZE"
	}
	for _, diagnostic := range result.Diagnostics {
		code := strings.ToUpper(diagnostic.Code)
		if strings.Contains(code, "SOURCE") || strings.Contains(code, "ANALYSIS") || strings.Contains(code, "NO_SEMANTIC") {
			return "GAP_ANALYZE"
		}
	}
	return "GAP_REALIZE"
}

func seeingRoleNodes(graph meaning.Meaning) (map[string]meaning.NodeID, error) {
	for _, node := range graph.Nodes {
		if node.Kind != meaning.KindEvent || node.Concept != "concept:seeing" {
			continue
		}
		roles := map[string]meaning.NodeID{"event": node.ID}
		for _, participant := range node.Participants {
			if participant.Role == "experiencer" || participant.Role == "stimulus" {
				roles[string(participant.Role)] = participant.Target
			}
		}
		if roles["experiencer"] != "" && roles["stimulus"] != "" {
			return roles, nil
		}
	}
	return nil, errors.New("Meaning has no seeing event with experiencer and stimulus roles")
}

func hasRolePath(path []meaning.NodeID, event, target meaning.NodeID) bool {
	return len(path) == 2 && path[0] == event && path[1] == target
}
