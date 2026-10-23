package tui

import (
	tea "charm.land/bubbletea/v2"
	"fmt"
	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/engine"
	"strings"
)

type previewDone struct {
	id, config uint64
	response   app.TranslationResponse
}

func (m model) selectedCandidate() *engine.TranslationAlternative {
	if m.response.Result != nil {
		for i := range m.response.Result.Candidates {
			c := &m.response.Result.Candidates[i]
			if c.ID == m.selectedID {
				return c
			}
		}
	}
	return nil
}
func (m *model) moveCandidate(delta int) {
	if m.response.Result == nil || len(m.response.Result.Candidates) == 0 {
		return
	}
	cs := m.response.Result.Candidates
	index := -1
	for i, c := range cs {
		if c.ID == m.selectedID {
			index = i
		}
	}
	if index < 0 {
		if delta > 0 {
			index = 0
		} else {
			index = len(cs) - 1
		}
	} else {
		index = (index + delta + len(cs)) % len(cs)
	}
	m.selectedID = cs[index].ID
	m.candidateDetail = false
	m.scroll[0] = 0
}
func (m *model) submitPreview() tea.Cmd {
	if m.pendingTranslate != 0 {
		m.notice = "Wait for the submitted translation before previewing."
		return nil
	}
	c := m.selectedCandidate()
	if c == nil {
		m.notice = "Select a candidate with Up/Down first."
		return nil
	}
	if c.Text == "" {
		m.notice = "Preview unavailable: this candidate returned no text."
		return nil
	}
	id := m.next()
	m.pendingPreview = id
	cfg := m.configRevision
	req := m.resultRequest
	req.PreviewID = c.ID
	s, ctx := m.service, m.ctx
	m.notice = ""
	return func() tea.Msg { return previewDone{id, cfg, s.Translate(ctx, req)} }
}
func candidateDetail(c *engine.TranslationAlternative) string {
	if c == nil {
		return "No candidate selected."
	}
	b := "Candidate ID: " + plain(c.ID) + "\nOutcome: " + string(c.Outcome) + "\nText: " + candidateText(c.Text)
	b += "\nSource ID: " + plain(c.SourceCandidateID) + "\nSemantic ID: " + plain(c.SemanticID)
	if len(c.Assumptions) == 0 {
		b += "\nAssumptions: none returned"
	}
	for _, a := range c.Assumptions {
		fmtPath := make([]string, len(a.Path))
		for i, p := range a.Path {
			fmtPath[i] = string(p)
		}
		b += "\nAssumption " + plain(strings.Join(fmtPath, "/")) + " " + plain(a.Field) + "=" + plain(a.Value) + ": " + plain(a.Reason)
	}
	if len(c.Diagnostics) == 0 {
		b += "\nDiagnostics: none returned"
	} else {
		b += diagnostics(c.Diagnostics)
	}
	for _, loss := range c.Losses {
		b += "\nLoss: " + plain(loss.Field) + " " + string(loss.Outcome) + ": " + plain(loss.Reason)
	}
	return b
}
func candidateText(text string) string {
	if text == "" {
		return "(not returned)"
	}
	return plain(text)
}
func (m model) candidatesBody() string {
	r := m.response.Result
	if r == nil {
		return ""
	}
	b := fmt.Sprintf("\nCandidates: %d · Up/Down select · Enter details · Preview control\n", len(r.Candidates))
	if m.candidateDetail {
		return b + "Inspection only; Esc returns to list.\n" + candidateDetail(m.selectedCandidate())
	}
	for i, c := range r.Candidates {
		mark := "  "
		if c.ID == m.selectedID {
			mark = ">>"
		}
		b += fmt.Sprintf("%s %d · %s · %s\nID: %s\nAssumptions: %d · Diagnostics: %d (Enter details)\n", mark, i+1, c.Outcome, candidateText(c.Text), plain(c.ID), len(c.Assumptions), len(c.Diagnostics))
	}
	return b
}
func (m model) previewBody() string {
	b := ""
	if m.pendingPreview != 0 {
		b += "\nPreview: waiting for Service.Translate"
	}
	if m.previewResponse.Failure != nil {
		b += "\nPreview setup status: " + string(m.previewResponse.Failure.Status) + diagnostics(m.previewResponse.Failure.Diagnostics)
	}
	if r := m.previewResponse.Result; r != nil {
		b += "\nPreview status: " + string(r.Status) + "\nPreview ID: " + plain(r.PreviewID) + "\nPreview text: " + candidateText(r.Text) + diagnostics(r.Diagnostics)
	}
	return b
}
