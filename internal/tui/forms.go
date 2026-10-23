package tui

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/features"
	"github.com/clickbytes/golingua/internal/ids"
)

// formsState holds the draft state of the F2 forms screen.
type formsState struct {
	language       string
	generate       bool
	surface        editor
	lexeme         editor
	featureText    editor
	revision       uint64
	pending        uint64
	resultRevision uint64
	response       *app.FormsResult
}

// languageID returns the selected language, or the UI default "pl".
// No language-specific feature defaults are applied here.
func (f formsState) languageID() string {
	if f.language != "" {
		return f.language
	}
	return "pl"
}

// activeEditor returns the editor the cursor is currently driving for the
// given focus item. Features focus wins; otherwise the mode selects lexeme
// (Generate) or surface (Analyze).
func (f *formsState) activeEditor(focus string) *editor {
	if focus == "features" {
		return &f.featureText
	}
	if f.generate {
		return &f.lexeme
	}
	return &f.surface
}

// request builds the app.FormsRequest for the current mode.
//
// Analyze sets ONLY Language and Surface; any retained feature draft is
// ignored and never validated here. Generate sets ONLY Language, LexemeID
// and explicitly parsed Features. No inventories, guessing or language
// mappings happen here; Service.Forms later validates against the actual
// language profile.
func (f formsState) request() (app.FormsRequest, error) {
	lang := f.languageID()
	if lang != "pl" && lang != "en" {
		return app.FormsRequest{}, fmt.Errorf("unsupported language %q (use pl or en)", lang)
	}
	if !f.generate {
		surface := strings.TrimSpace(f.surface.text)
		if surface == "" {
			return app.FormsRequest{}, fmt.Errorf("surface text is required")
		}
		if len(surface) > 1024 {
			return app.FormsRequest{}, fmt.Errorf("surface text exceeds 1024 bytes")
		}
		return app.FormsRequest{Language: lang, Surface: surface}, nil
	}

	req := app.FormsRequest{Language: lang}

	lexeme := f.lexeme.text
	if lexeme == "" {
		return app.FormsRequest{}, fmt.Errorf("lexeme id is required")
	}
	if len(lexeme) > 1024 {
		return app.FormsRequest{}, fmt.Errorf("lexeme id exceeds 1024 bytes")
	}
	parsedID, err := ids.ParseLexemeID(lexeme)
	if err != nil || parsedID.String() != lexeme {
		return app.FormsRequest{}, fmt.Errorf("invalid canonical lexeme ID")
	}

	req.LexemeID = lexeme

	features, err := parseFeatureTokens(f.featureText.text)
	if err != nil {
		return app.FormsRequest{}, err
	}
	if len(features) == 0 {
		return app.FormsRequest{}, fmt.Errorf("generate requires at least one feature (name=value)")
	}
	req.Features = features
	return req, nil
}

// parseFeatureTokens parses whitespace-separated key=value tokens into
// explicit feature values. Syntax and size limits only; values are passed
// through unchanged for later profile validation.
func parseFeatureTokens(text string) ([]features.FeatureValue, error) {
	tokens := strings.Fields(text)
	if len(tokens) > 32 {
		return nil, fmt.Errorf("too many features: %d (max 32)", len(tokens))
	}
	seen := make(map[string]bool, len(tokens))
	out := make([]features.FeatureValue, 0, len(tokens))
	for _, token := range tokens {
		if strings.Count(token, "=") != 1 {
			return nil, fmt.Errorf("feature %q must contain exactly one '='", token)
		}
		parts := strings.SplitN(token, "=", 2)
		name, value := parts[0], parts[1]
		if n := len(name); n == 0 || n > 64 {
			return nil, fmt.Errorf("feature name must be 1..64 bytes in %q", token)
		}
		if n := len(value); n == 0 || n > 64 {
			return nil, fmt.Errorf("feature value must be 1..64 bytes in %q", token)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate feature name %q", name)
		}
		seen[name] = true
		out = append(out, features.FeatureValue{
			Feature: features.FeatureID(name),
			State:   features.StateKnown,
			Value:   features.FeatureValueID(value),
		})
	}
	return out, nil
}

// body renders the F2 forms screen as plain text. It is scrollable by the
// shell: no truncation and no auto-selection of returned results.
func (f formsState) body() string {
	var b strings.Builder

	mode := "Analyze"
	if f.generate {
		mode = "Generate"
	}
	fmt.Fprintf(&b, "F2 Forms — %s (language: %s)\n", mode, f.languageID())
	b.WriteString("Controls use Enter: select language or mode, then press Enter to apply.\n")

	if f.generate {
		fmt.Fprintf(&b, "Lexeme ID: %s\n", plain(f.lexeme.text))
		fmt.Fprintf(&b, "Features (name=value, space-separated): %s\n", plain(f.featureText.text))
	} else {
		fmt.Fprintf(&b, "Surface: %s\n", plain(f.surface.text))
	}

	if f.pending != 0 {
		b.WriteString("Busy: request in progress...\n")
	}

	res := f.response
	if res == nil {
		b.WriteString("Press Enter to submit.\n")
		return b.String()
	}
	if f.resultRevision != f.revision {
		b.WriteString("Result is from an older draft; current draft has changed.\n")
	}
	fmt.Fprintf(&b, "Status: %s | State: %s | Language: %s\n", res.Status, res.State, res.Language)

	if len(res.Diagnostics) > 0 {
		fmt.Fprintf(&b, "Diagnostics (%d): %s\n", len(res.Diagnostics), diagnostics(res.Diagnostics))
	}

	fmt.Fprintf(&b, "Readings: %d\n", len(res.Readings))
	for i, reading := range res.Readings {
		fmt.Fprintf(&b, "Reading %d: %s\n", i+1, plain(reading.Surface))
		writeFormsJSON(&b, i+1, reading)
	}

	fmt.Fprintf(&b, "Forms: %d\n", len(res.Forms))
	for i, form := range res.Forms {
		fmt.Fprintf(&b, "Form %d: %s\n", i+1, plain(form.Surface))
		writeFormsJSON(&b, i+1, form)
	}
	return b.String()
}

// writeFormsJSON preserves returned fields in the scrollable result.
func writeFormsJSON(b *strings.Builder, n int, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(b, "%d: %v\n", n, v)
		return
	}
	fmt.Fprintf(b, "%d: ", n)
	b.Write(data)
	b.WriteString("\n")
}
