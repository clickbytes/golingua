package translation_test

import (
	"context"
	"encoding/json"
	"os"
	"sort"
	"testing"

	"github.com/clickbytes/golingua/internal/engine"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/tests/interlingua"
	"github.com/clickbytes/golingua/tests/matrix"
)

func TestMeaningAttributesReadingTranslation(t *testing.T) {
	runtime := matrix.Runtime(t)
	cases := []struct {
		name      string
		from, to  features.LanguageID
		input, id string
		texts     []string
	}{
		{"English present to Polish", "en", "pl", "The teacher reads the message.", "reading.teacher-message.present", []string{"Nauczyciel czyta wiadomość."}},
		{"English past to Polish", "en", "pl", "The teacher read the message.", "reading.teacher-message.past", []string{"Nauczyciel czytał wiadomość."}},
		{"Polish plural present to English", "pl", "en", "Nauczyciele czytają wiadomość.", "reading.teachers-message.present", []string{"Teachers read a message.", "Teachers read the message.", "The teachers read a message.", "The teachers read the message."}},
		{"Polish plural past to English", "pl", "en", "Nauczyciele czytali wiadomość.", "reading.teachers-message.past", []string{"Teachers read a message.", "Teachers read the message.", "The teachers read a message.", "The teachers read the message."}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, e := runtime.Translate(context.Background(), engine.Request{From: tc.from, To: tc.to, Text: tc.input})
			logTranslationAttributes(t, tc.input, result, e)
			if e != nil {
				t.Fatalf("FAIL: translation error: %v", e)
			}
			if len(result.Candidates) == 0 {
				t.Fatalf("GAP: translation has no candidates (%s): %+v", result.Status, result.Diagnostics)
			}
			want, ok, err := interlingua.Reading(tc.id)
			if err != nil || !ok {
				t.Fatalf("fixture: %v", err)
			}
			texts := map[string]bool{}
			matched := false
			for _, candidate := range result.Candidates {
				if candidate.Meaning == nil || candidate.Text == "" {
					t.Errorf("translation candidate %q lacks required Meaning or surface", candidate.ID)
				}
				if candidate.Text != "" {
					texts[candidate.Text] = true
				}
				if candidate.Meaning != nil {
					same, err := matrix.GraphMatches(want, *candidate.Meaning)
					if err == nil && same {
						matched = true
					}
					if err != nil || !same {
						t.Errorf("FAIL: intermediate Meaning differs from %s (match=%t error=%v)", tc.id, same, err)
					}
				}
				for _, a := range candidate.Assumptions {
					if a.Field == "tense" || a.Field == "number" {
						t.Errorf("FAIL: known I attribute assumed: %+v", a)
					}
				}
			}
			if !matched {
				t.Errorf("FAIL: no translation candidate retained expected Meaning %s", tc.id)
			}
			got := make([]string, 0, len(texts))
			for text := range texts {
				got = append(got, text)
			}
			sort.Strings(got)
			expected := append([]string(nil), tc.texts...)
			sort.Strings(expected)
			if !equalStringsAttributes(got, expected) {
				t.Errorf("FAIL: texts=%v want=%v; status=%s", got, expected, result.Status)
			}
		})
	}
}

func logTranslationAttributes(t *testing.T, label string, value any, err error) {
	t.Helper()
	b, _ := json.Marshal(value)
	sha := os.Getenv("GOLINGUA_MATRIX_SHA")
	if sha == "" {
		sha = "UNRECORDED"
	}
	t.Logf("SHA=%s %s result=%s error=%v", sha, label, b, err)
}
func equalStringsAttributes(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
