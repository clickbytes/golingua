// Package tui provides the interactive terminal client of the shared app service.
package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/clickbytes/golingua/internal/app"
	"github.com/clickbytes/golingua/internal/status"
	"golang.org/x/term"
)

type translationDone struct {
	id, revision, config uint64
	response             app.TranslationResponse
	request              app.TranslateRequest
}
type runtimeDone struct {
	id, config uint64
	result     app.RuntimeValidationResult
}

type model struct {
	palette                                           paletteState
	selectedID                                        string
	candidateDetail                                   bool
	pendingPreview                                    uint64
	previewResponse                                   app.TranslationResponse
	resultRequest                                     app.TranslateRequest
	service                                           *app.Service
	ctx                                               context.Context
	width, height, screen                             int
	focuses                                           [5]int
	drafts                                            [5]editor
	from, to                                          string
	sequence, configRevision                          uint64
	pendingTranslate, pendingRuntime, pendingSave     uint64
	resultRevision                                    uint64
	response                                          app.TranslationResponse
	runtime                                           app.RuntimeValidationResult
	runtimePath, settingsFile, notice, settingsNotice string
	scroll                                            [5]int
	help                                              bool
	loaded                                            bool
}

// Interactive is deliberately limited to no-argument stdin+stdout TTY launches.
func Interactive(in, out *os.File) bool {
	return term.IsTerminal(int(in.Fd())) && term.IsTerminal(int(out.Fd())) && os.Getenv("TERM") != "dumb"
}
func Run(service *app.Service, in, out *os.File) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m := model{service: service, ctx: ctx, from: "en", to: "pl", width: 80, height: 24}
	_, err := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out), tea.WithContext(ctx)).Run()
	return err
}
func (m model) Init() tea.Cmd { return func() tea.Msg { return loadSettings() } }
func (m *model) next() uint64 { m.sequence++; return m.sequence }
func (m model) focusNames() []string {
	if m.screen == 0 {
		return []string{"prompt", "direction", "candidates", "preview", "result", "tabs"}
	}
	if m.screen == 4 {
		return []string{"prompt", "validate", "save", "result", "tabs"}
	}
	return []string{"prompt", "tabs"}
}
func (m model) focus() string { return m.focusNames()[m.focuses[m.screen]%len(m.focusNames())] }
func (m *model) selectRuntime(path string) {
	if path == m.runtimePath {
		return
	}
	m.runtimePath = path
	m.configRevision++
	m.pendingTranslate = 0
	m.response = app.TranslationResponse{}
	m.pendingPreview = 0
	m.previewResponse = app.TranslationResponse{}
	m.selectedID = ""
	m.candidateDetail = false
	m.runtime = app.RuntimeValidationResult{}
}
func (m *model) checkRuntime() tea.Cmd {
	id := m.next()
	m.pendingRuntime = id
	config := m.configRevision
	path := m.runtimePath
	s, ctx := m.service, m.ctx
	return func() tea.Msg {
		return runtimeDone{id, config, s.ValidateRuntime(ctx, app.RuntimeValidationRequest{RuntimeConfig: path})}
	}
}
func (m *model) pathFromDraft() bool {
	text := strings.TrimSpace(m.drafts[4].text)
	if text == "" {
		m.selectRuntime("")
		return true
	}
	path, err := filepath.Abs(text)
	if err != nil {
		m.settingsNotice = "Path error: " + err.Error()
		return false
	}
	m.drafts[4].set(path)
	m.selectRuntime(path)
	return true
}
func (m *model) submitTranslate() tea.Cmd {
	id := m.next()
	m.pendingPreview = 0
	m.previewResponse = app.TranslationResponse{}
	m.pendingTranslate = id
	rev, cfg := m.drafts[0].revision, m.configRevision
	req := app.TranslateRequest{From: m.from, To: m.to, Text: m.drafts[0].text, RuntimeConfig: m.runtimePath}
	s, ctx := m.service, m.ctx
	return func() tea.Msg {
		return translationDone{id: id, revision: rev, config: cfg, response: s.Translate(ctx, req), request: req}
	}
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
	case settingsLoaded:
		m.loaded = true
		m.settingsFile = v.path
		if v.err != nil {
			m.settingsNotice = "Warning: " + v.err.Error()
			return m, nil
		}
		m.selectRuntime(v.settings.RuntimeConfig)
		m.drafts[4].set(v.settings.RuntimeConfig)
		return m, m.checkRuntime()
	case settingsSaved:
		if v.id == m.pendingSave {
			m.pendingSave = 0
			if v.err != nil {
				m.settingsNotice = "Save failed: " + v.err.Error()
			} else {
				m.settingsNotice = "Settings saved explicitly."
			}
		}
	case runtimeDone:
		if v.id == m.pendingRuntime && v.config == m.configRevision {
			m.pendingRuntime = 0
			m.runtime = v.result
		}
	case previewDone:
		if v.id == m.pendingPreview && v.config == m.configRevision {
			m.pendingPreview = 0
			m.previewResponse = v.response
			m.scroll[0] = 0
		}
	case translationDone:
		if v.id == m.pendingTranslate && v.config == m.configRevision {
			m.pendingTranslate = 0
			m.response = v.response
			m.resultRequest = v.request
			m.previewResponse = app.TranslationResponse{}
			m.pendingPreview = 0
			if m.selectedCandidate() == nil {
				m.selectedID = ""
				m.candidateDetail = false
			}
			m.resultRevision = v.revision
			m.scroll[0] = 0
		}
	case tea.PasteMsg:
		if m.palette.open {
			m.palette.input.insert(v.Content)
			m.palette.selected = -1
			return m, nil
		}
		if m.focus() == "prompt" && (m.screen == 0 || m.screen == 4) {
			m.drafts[m.screen].insert(v.Content)
		}
	case tea.KeyPressMsg:
		key := v.String()
		if key == "ctrl+c" || key == "ctrl+d" {
			return m, tea.Quit
		}
		switch key {
		case "f1", "f2", "f3", "f4", "f5":
			m.screen = int(key[1] - '1')
			m.help = false
			m.palette.open = false
			return m, nil
		case "tab", "shift+tab":
			m.palette.open = false
			n := len(m.focusNames())
			delta := 1
			if key == "shift+tab" {
				delta = n - 1
			}
			m.focuses[m.screen] = (m.focuses[m.screen] + delta) % n
			return m, nil
		case "esc":
			if m.palette.open {
				m.palette.open = false
				m.notice = ""
				return m, nil
			}
			if m.screen == 0 && m.candidateDetail {
				m.candidateDetail = false
				m.scroll[0] = 0
				return m, nil
			}
			m.help = false
			m.focuses[m.screen] = 0
			return m, nil
		case "?":
			if m.palette.open {
				m.paletteKey(v)
				return m, nil
			}
			m.help = !m.help
			return m, nil
		}
		if key == "ctrl+p" {
			m.openPalette()
			return m, nil
		}
		if m.palette.open {
			m.paletteKey(v)
			return m, nil
		}
		focus := m.focus()
		if focus == "prompt" && m.screen != 4 && v.Key().Text == "/" && (m.drafts[m.screen].cursor == 0 || m.screen != 0 && m.screen != 4) {
			m.openPalette()
			return m, nil
		}
		if m.screen == 0 && focus == "preview" && key == "enter" {
			return m, m.submitPreview()
		}
		if m.screen == 0 && focus == "candidates" {
			switch key {
			case "down", "up":
				if m.candidateDetail {
					if key == "down" {
						m.scroll[0]++
					} else {
						m.scroll[0] = max(0, m.scroll[0]-1)
					}
				} else {
					delta := 1
					if key == "up" {
						delta = -1
					}
					m.moveCandidate(delta)
				}
			case "enter":
				if m.selectedCandidate() != nil {
					m.candidateDetail = true
					m.scroll[0] = 0
				}
			case "pgdown":
				m.scroll[0] += max(1, m.height-8)
			case "pgup":
				m.scroll[0] = max(0, m.scroll[0]-max(1, m.height-8))
			}
			return m, nil
		}
		if focus == "tabs" && (key == "left" || key == "right") {
			delta := 1
			if key == "left" {
				delta = 4
			}
			m.screen = (m.screen + delta) % 5
			m.focuses[m.screen] = len(m.focusNames()) - 1
			return m, nil
		}
		if (focus == "direction" && key == "enter") || (m.screen == 0 && key == "alt+d") {
			m.from, m.to = m.to, m.from
			m.drafts[0].revision++
			return m, nil
		}
		if focus == "result" {
			switch key {
			case "down":
				m.scroll[m.screen]++
			case "up":
				if m.scroll[m.screen] > 0 {
					m.scroll[m.screen]--
				}
			case "pgdown":
				m.scroll[m.screen] += max(1, m.height-8)
			case "pgup":
				m.scroll[m.screen] = max(0, m.scroll[m.screen]-max(1, m.height-8))
			}
			return m, nil
		}
		if m.screen == 4 && (key == "ctrl+s" || key == "enter" && focus == "save") {
			if !m.loaded || m.pendingSave != 0 {
				return m, nil
			}
			if !m.pathFromDraft() {
				return m, nil
			}
			id := m.next()
			m.pendingSave = id
			path := m.settingsFile
			value := settings{SchemaVersion: 1, RuntimeConfig: m.runtimePath}
			save := func() tea.Msg { return settingsSaved{id, saveSettings(path, value)} }
			return m, tea.Batch(save, m.checkRuntime())
		}
		if key == "enter" {
			if m.screen == 0 && focus == "prompt" {
				return m, m.submitTranslate()
			}
			if m.screen == 4 && (focus == "prompt" || focus == "validate") && m.loaded {
				if m.pathFromDraft() {
					return m, m.checkRuntime()
				}
			}
			return m, nil
		}
		if focus == "prompt" && (m.screen == 0 || m.screen == 4) {
			m.drafts[m.screen].update(v)
		}
	}
	return m, nil
}
func diagnostics(ds []status.Diagnostic) string {
	var b strings.Builder
	for _, d := range ds {
		fmt.Fprintf(&b, "\n%s: %s (%s)", plain(d.Code), plain(d.Message), plain(d.Path))
	}
	return b.String()
}
func (m model) translationBody() string {
	b := "F1 Translate · " + m.from + " -> " + m.to + " · Alt+D changes direction\n"
	if m.pendingTranslate != 0 {
		b += "Busy: waiting for Service.Translate\n"
	}
	if m.response.Failure != nil {
		return b + "Setup status: " + string(m.response.Failure.Status) + diagnostics(m.response.Failure.Diagnostics)
	}
	r := m.response.Result
	if r == nil {
		return b + "Enter submits the prompt. Select runtime config in F5."
	}
	if m.resultRevision != m.drafts[0].revision {
		b += "Result belongs to an older draft/direction.\n"
	}
	b += "Result status: " + string(r.Status) + "\nText: "
	if r.Text == "" {
		b += "(not returned)"
	} else {
		b += plain(r.Text)
	}
	return b + diagnostics(r.Diagnostics) + m.previewBody() + m.candidatesBody()
}
func (m model) View() tea.View {
	body := "Not implemented yet. This tab has no service actions."
	footer := "F1-F5 Screens · Tab Focus · Enter Action · Esc Back · Ctrl+C Exit"
	switch m.screen {
	case 0:
		body = m.translationBody()
	case 4:
		state := "not verified"
		if m.pendingRuntime != 0 {
			state = "checking"
		} else if m.runtime.Result.Status != "" {
			state = string(m.runtime.Result.Status)
		}
		body = "F5 Runtime config · Enter Validate/select · Ctrl+S Save\nValidation: " + state + diagnostics(m.runtime.Result.Diagnostics) + "\nSettings file: " + plain(m.settingsFile) + "\n" + plain(m.settingsNotice) + "\nSave is explicit; no snapshot is applied."
	}
	if m.help {
		body = "Tab/Shift+Tab cycle controls; tabs support Left/Right.\nF1: prompt Enter translates; direction Enter or Alt+D swaps PL/EN.\nF5: prompt Enter validates/selects path; Ctrl+S explicitly saves.\nF1 candidates: Up/Down select, Enter details; Preview control calls service.\nCtrl+P opens navigation palette; Esc cancels. Result focus scrolls. Ctrl+U clears input.\nEsc closes help; Ctrl+C or Ctrl+D restores terminal and exits."
	}
	if m.palette.open {
		body = m.paletteBody()
	} else if m.notice != "" {
		body = plain(m.notice) + "\n" + body
	}
	body = ansi.Hardwrap(body, max(1, m.width), true)
	lines := strings.Split(body, "\n")
	start := min(m.scroll[m.screen], max(0, len(lines)-max(1, m.height-6)))
	if m.screen == 0 && m.focus() == "candidates" && !m.candidateDetail && !m.palette.open {
		for i, line := range lines {
			if strings.HasPrefix(line, ">> ") {
				start = max(0, i-2)
				break
			}
		}
	}
	body = strings.Join(lines[start:], "\n")
	prompt := m.drafts[m.screen].visible(m.width - 4)
	if m.screen != 0 && m.screen != 4 {
		prompt = "No text action available"
	}
	if m.palette.open {
		prompt = m.palette.input.visible(m.width - 4)
	}
	if m.settingsNotice != "" && m.screen != 4 {
		footer = plain(m.settingsNotice) + " · F5 details · Ctrl+C Exit"
	}
	focus := m.focus()
	if m.palette.open {
		focus = "palette"
	}
	v := tea.NewView(renderShell(m.width, m.height, m.screen, focus, plain(m.runtimePath), body, prompt, footer))
	v.AltScreen = true
	return v
}
