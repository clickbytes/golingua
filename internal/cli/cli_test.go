package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestBootstrapCLI(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		code    int
		output  string
		errtext string
	}{
		{name: "help by default", args: nil, code: ExitOK, output: "NOT IMPLEMENTED"},
		{name: "help flag", args: []string{"--help"}, code: ExitOK, output: "golingua bootstrap"},
		{name: "version", args: []string{"--version"}, code: ExitOK, output: Version},
		{name: "translate denied", args: []string{"translate", "--from", "pl", "--to", "en", "--text", "Kot."}, code: ExitUnsupported, errtext: "LANGUAGE_RUNTIME_UNAVAILABLE"},
		{name: "unknown denied", args: []string{"generate"}, code: ExitUsage, errtext: "INVALID_COMMAND"},
		{name: "malformed capabilities denied", args: []string{"capabilities"}, code: ExitUsage, errtext: "INVALID_COMMAND"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var out, errs bytes.Buffer
			code := Run(tc.args, &out, &errs)
			if code != tc.code {
				t.Fatalf("got exit %d; want %d", code, tc.code)
			}
			if !strings.Contains(out.String(), tc.output) {
				t.Fatalf("stdout %q does not contain %q", out.String(), tc.output)
			}
			if !strings.Contains(errs.String(), tc.errtext) {
				t.Fatalf("stderr %q does not contain %q", errs.String(), tc.errtext)
			}
			if tc.code != ExitOK && out.Len() > 0 {
				t.Fatalf("unexpected success stdout %q", out.String())
			}
		})
	}
}

func TestCapabilitiesAreEmptyAndDeterministic(t *testing.T) {
	var first, second, errs bytes.Buffer
	if Run([]string{"capabilities", "--json"}, &first, &errs) != ExitOK || errs.Len() != 0 {
		t.Fatalf("invalid capabilities: %q", errs.String())
	}
	if Run([]string{"capabilities", "--json"}, &second, &errs) != ExitOK {
		t.Fatal("second invocation failed")
	}
	if first.String() != second.String() {
		t.Fatalf("non-deterministic json %q != %q", first.String(), second.String())
	}
	var result struct {
		Status    string   `json:"status"`
		Version   string   `json:"version"`
		Languages []string `json:"languages"`
		Features  []string `json:"features"`
		Ready     bool     `json:"ready"`
	}
	if err := json.Unmarshal(first.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if result.Ready || result.Status != "unsupported" || result.Version != Version || len(result.Languages) != 0 || len(result.Features) != 0 {
		t.Fatalf("claims unsupported capabilities: %+v", result)
	}
}
