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
	"github.com/clickbytes/golingua/internal/core"
	"github.com/clickbytes/golingua/internal/ids"
	"github.com/clickbytes/golingua/internal/status"
	"golang.org/x/term"
)

type translationDone struct {
	id, revision, config uint64
	response             app.TranslationResponse
	request              app.TranslateRequest
}
type runtimeDone struct {
	id, config, draftRevision uint64
	path                      string
	result                    app.RuntimeValidationResult
}
type formsDone struct {
	id, revision uint64
	response     app.FormsResult
}
type conceptsStart struct{}
type conceptsCatalogDone struct {
	id       uint64
	response app.CatalogResult
}
type conceptsLookupDone struct {
	id, revision uint64
	conceptID    string
	concept      core.Concept
	found        bool
	err          error
}
type languagesCatalogDone struct {
	id       uint64
	response app.CatalogResult
}
type languagesValidationDone struct {
	id, config uint64
	path       string
	result     app.RuntimeValidationResult
}

type model struct {
	palette                                           paletteState
	forms                                             formsState
	formsDetails                                      bool
	concepts                                          conceptsState
	languages                                         languagesState
	preferences                                       preferencesState
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
	m := model{service: service, ctx: ctx, from: "en", to: "pl", width: 80, height: 24, preferences: newPreferencesState()}
	_, err := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out), tea.WithContext(ctx)).Run()
	return err
}
func (m model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return loadSettings() },
		func() tea.Msg { return conceptsStart{} },
	)
}
func (m *model) next() uint64 { m.sequence++; return m.sequence }
func (m *model) loadConcepts() tea.Cmd {
	id := m.next()
	m.concepts.pendingCatalog = id
	s, ctx := m.service, m.ctx
	return func() tea.Msg {
		return conceptsCatalogDone{id: id, response: s.Catalog(ctx, app.CatalogRequest{Details: true})}
	}
}
func (m *model) lookupConcept() tea.Cmd {
	if m.concepts.selectedID == "" {
		return nil
	}
	id := m.next()
	revision, conceptID := m.concepts.revision, m.concepts.selectedID
	m.concepts.pendingLookup = id
	m.concepts.detail = true
	m.concepts.lookup = nil
	m.concepts.lookupErr = ""
	m.concepts.lookupFound = false
	m.concepts.detailScroll = 0
	s, ctx := m.service, m.ctx
	return func() tea.Msg {
		parsed, err := ids.ParseConceptID(conceptID)
		if err != nil {
			return conceptsLookupDone{id: id, revision: revision, conceptID: conceptID, err: err}
		}
		concept, found, err := s.LookupConcept(ctx, parsed)
		return conceptsLookupDone{id: id, revision: revision, conceptID: conceptID, concept: concept, found: found, err: err}
	}
}
func (m *model) refreshLanguages() tea.Cmd {
	catalogID, validationID := m.next(), m.next()
	path, config := m.runtimePath, m.configRevision
	m.languages.path = path
	m.languages.pendingCatalog = catalogID
	m.languages.pendingValidation = validationID
	s, ctx := m.service, m.ctx
	catalogCmd := func() tea.Msg {
		return languagesCatalogDone{id: catalogID, response: s.Catalog(ctx, app.CatalogRequest{})}
	}
	validationCmd := func() tea.Msg {
		return languagesValidationDone{id: validationID, config: config, path: path, result: s.ValidateRuntime(ctx, app.RuntimeValidationRequest{RuntimeConfig: path})}
	}
	return tea.Batch(catalogCmd, validationCmd)
}
func (m *model) enterScreen(target int) tea.Cmd {
	m.screen = target
	m.help = false
	m.palette.open = false
	if target == 3 {
		return m.refreshLanguages()
	}
	return nil
}
func (m model) focusNames() []string {
	if m.screen == 0 {
		return []string{"prompt", "direction", "candidates", "preview", "result", "tabs"}
	}
	if m.screen == 4 {
		return []string{"prompt", "validate", "level", "default_direction", "store", "save", "cancel", "result", "tabs"}
	}
	if m.screen == 1 {
		if m.forms.generate {
			return []string{"prompt", "language", "mode", "features", "run", "result", "tabs"}
		}
		return []string{"prompt", "language", "mode", "run", "result", "tabs"}
	}
	if m.screen == 2 {
		return []string{"prompt", "list", "result", "tabs"}
	}
	if m.screen == 3 {
		return []string{"refresh", "result", "tabs"}
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
	m.languages.pendingValidation = 0
	m.languages.validation = app.RuntimeValidationResult{}
	m.languages.path = path
}
func (m *model) checkRuntime() tea.Cmd {
	id := m.next()
	m.pendingRuntime = id
	config := m.configRevision
	draftRevision := m.drafts[4].revision
	path := m.drafts[4].text
	if path != "" {
		absolute, err := filepath.Abs(path)
		if err != nil {
			m.pendingRuntime = 0
			m.runtime = app.RuntimeValidationResult{}
			m.settingsNotice = "Runtime path error: " + err.Error()
			return nil
		}
		path = absolute
	}
	s, ctx := m.service, m.ctx
	return func() tea.Msg {
		return runtimeDone{id: id, config: config, draftRevision: draftRevision, path: path, result: s.ValidateRuntime(ctx, app.RuntimeValidationRequest{RuntimeConfig: path})}
	}
}
func (m *model) runtimeDraftChanged() {
	m.pendingRuntime = 0
	m.runtime = app.RuntimeValidationResult{}
}
func (m *model) cancelPreferencesDraft() {
	m.preferences.reset(m.preferences.active)
	m.drafts[4].set(m.preferences.active.RuntimeConfig)
	m.runtimeDraftChanged()
	m.settingsNotice = "Preference drafts reset to saved values."
}
func (m *model) savePreferences() tea.Cmd {
	if !m.loaded || m.pendingSave != 0 {
		return nil
	}
	value, err := m.preferences.value(m.drafts[4].text)
	if err != nil {
		m.settingsNotice = "Invalid preference draft: " + err.Error()
		return nil
	}
	id, path := m.next(), m.settingsFile
	m.pendingSave = id
	m.settingsNotice = "Saving preferences…"
	return func() tea.Msg {
		return settingsSaved{id: id, value: value, err: saveSettings(path, value)}
	}
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
func (m *model) submitForms() tea.Cmd {
	req, err := m.forms.request()
	if err != nil {
		m.forms.pending = 0
		m.forms.response = &app.FormsResult{
			Status:      status.OutcomeInvalid,
			Language:    m.forms.languageID(),
			Diagnostics: []status.Diagnostic{{Code: "INVALID_FORMS_INPUT", Message: err.Error(), Path: "forms", Severity: status.SeverityError}},
		}
		m.forms.resultRevision = m.forms.revision
		return nil
	}
	id := m.next()
	m.forms.pending = id
	revision := m.forms.revision
	s, ctx := m.service, m.ctx
	return func() tea.Msg {
		return formsDone{id: id, revision: revision, response: s.Forms(ctx, req)}
	}
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch v := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
	case conceptsStart:
		if m.concepts.catalog == nil && m.concepts.pendingCatalog == 0 {
			return m, m.loadConcepts()
		}
	case conceptsCatalogDone:
		if v.id == m.concepts.pendingCatalog {
			m.concepts.pendingCatalog = 0
			previous := m.concepts.selectedID
			m.concepts.catalog = &v.response
			m.concepts.reconcile()
			if previous != m.concepts.selectedID {
				m.concepts.pendingLookup = 0
				m.concepts.detail = false
				m.concepts.lookup = nil
				m.concepts.lookupErr = ""
				m.concepts.lookupFound = false
				m.concepts.detailScroll = 0
			}
			m.concepts.listScroll = 0
		}
	case conceptsLookupDone:
		if v.id == m.concepts.pendingLookup && v.revision == m.concepts.revision && v.conceptID == m.concepts.selectedID {
			m.concepts.pendingLookup = 0
			m.concepts.lookupErr = ""
			if v.err != nil {
				m.concepts.lookupErr = v.err.Error()
			}
			m.concepts.lookupFound = v.found && v.err == nil
			if m.concepts.lookupFound {
				concept := v.concept
				m.concepts.lookup = &concept
			}
		}
	case languagesCatalogDone:
		if v.id == m.languages.pendingCatalog {
			m.languages.pendingCatalog = 0
			m.languages.catalog = &v.response
		}
	case languagesValidationDone:
		if v.id == m.languages.pendingValidation && v.config == m.configRevision && v.path == m.runtimePath && v.path == m.languages.path {
			m.languages.pendingValidation = 0
			m.languages.validation = v.result
		}
	case settingsLoaded:
		m.loaded = true
		m.settingsFile = v.path
		if v.err != nil {
			m.settingsNotice = "Warning: " + v.err.Error()
			if m.screen == 3 {
				return m, tea.Batch(m.checkRuntime(), m.refreshLanguages())
			}
			return m, m.checkRuntime()
		}
		beforeFrom, beforeTo := m.from, m.to
		m.preferences.reset(v.settings)
		m.from = v.settings.DefaultFrom
		if m.from == "pl" {
			m.to = "en"
		} else {
			m.from, m.to = "en", "pl"
		}
		if beforeFrom != m.from || beforeTo != m.to {
			m.drafts[0].revision++
		}
		m.selectRuntime(v.settings.RuntimeConfig)
		m.drafts[4].set(v.settings.RuntimeConfig)
		if m.screen == 3 {
			return m, tea.Batch(m.checkRuntime(), m.refreshLanguages())
		}
		return m, m.checkRuntime()
	case settingsSaved:
		if v.id == m.pendingSave {
			m.pendingSave = 0
			if v.err != nil {
				m.settingsNotice = "Save failed: " + v.err.Error()
				return m, nil
			}
			activeSettings := m.preferences.active
			m.preferences.active = v.value
			m.preferences.reset(v.value)
			m.drafts[4].set(v.value.RuntimeConfig)
			m.selectRuntime(v.value.RuntimeConfig)
			if activeSettings.DefaultFrom != v.value.DefaultFrom {
				m.from = v.value.DefaultFrom
				if m.from == "pl" {
					m.to = "en"
				} else {
					m.from, m.to = "en", "pl"
				}
				m.drafts[0].revision++
			}
			if activeSettings.DisplayLevel != v.value.DisplayLevel {
				m.formsDetails = false
			}
			m.settingsNotice = "Settings saved and applied."
			check := m.checkRuntime()
			if m.screen == 3 {
				return m, tea.Batch(check, m.refreshLanguages())
			}
			return m, check
		}
	case runtimeDone:
		currentPath := m.drafts[4].text
		if currentPath != "" {
			currentPath, _ = filepath.Abs(currentPath)
		}
		if v.id == m.pendingRuntime && v.config == m.configRevision && v.draftRevision == m.drafts[4].revision && v.path == currentPath {
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
	case formsDone:
		if v.id == m.forms.pending {
			m.forms.pending = 0
			m.forms.response = &v.response
			m.forms.resultRevision = v.revision
			m.scroll[1] = 0
		}
	case tea.PasteMsg:
		if m.palette.open {
			m.palette.input.insert(v.Content)
			m.palette.selected = -1
			return m, nil
		}
		if m.screen == 4 && m.pendingSave != 0 {
			return m, nil
		}
		if m.screen == 1 && (m.focus() == "prompt" || m.focus() == "features") {
			e := m.forms.activeEditor(m.focus())
			before := e.revision
			e.insert(v.Content)
			if e.revision != before {
				m.forms.revision++
			}
		} else if m.screen == 4 && m.focus() == "store" {
			m.preferences.store.insert(v.Content)
		} else if m.focus() == "prompt" && (m.screen == 0 || m.screen == 4) {
			before := m.drafts[4].revision
			m.drafts[m.screen].insert(v.Content)
			if m.screen == 4 && before != m.drafts[4].revision {
				m.runtimeDraftChanged()
			}
		} else if m.screen == 2 && m.focus() == "prompt" {
			before := m.concepts.filter.revision
			m.concepts.filter.insert(v.Content)
			if before != m.concepts.filter.revision {
				m.concepts.edited()
			}
		}
	case tea.KeyPressMsg:
		key := v.String()
		if key == "ctrl+c" || key == "ctrl+d" {
			return m, tea.Quit
		}
		switch key {
		case "f1", "f2", "f3", "f4", "f5":
			return m, m.enterScreen(int(key[1] - '1'))
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
			if m.screen == 4 && m.help {
				m.help = false
				return m, nil
			}
			if m.screen == 0 && m.candidateDetail {
				m.candidateDetail = false
				m.scroll[0] = 0
				return m, nil
			}
			if m.screen == 2 && m.concepts.detail {
				m.concepts.detail = false
				m.concepts.pendingLookup = 0
				m.concepts.lookup = nil
				m.concepts.lookupErr = ""
				m.concepts.lookupFound = false
				return m, nil
			}
			if m.screen == 4 {
				if m.pendingSave != 0 {
					return m, nil
				}
				m.cancelPreferencesDraft()
				return m, nil
			}
			m.help = false
			m.focuses[m.screen] = 0
			return m, nil
		case "?":
			if m.palette.open {
				return m, m.paletteKey(v)
			}
			m.help = !m.help
			return m, nil
		}
		if key == "ctrl+p" {
			m.openPalette()
			return m, nil
		}
		if m.palette.open {
			return m, m.paletteKey(v)
		}
		focus := m.focus()
		if m.screen == 4 && m.pendingSave != 0 {
			return m, nil
		}
		cursor := m.drafts[m.screen].cursor
		if m.screen == 1 && (focus == "prompt" || focus == "features") {
			cursor = m.forms.activeEditor(focus).cursor
		} else if m.screen == 2 && focus == "prompt" {
			cursor = m.concepts.filter.cursor
		}
		if focus == "prompt" && m.screen != 4 && v.Key().Text == "/" && cursor == 0 {
			m.openPalette()
			return m, nil
		}
		if m.screen == 2 && m.concepts.detail && (focus == "list" || focus == "result") {
			switch key {
			case "up":
				m.concepts.detailScroll = max(0, m.concepts.detailScroll-1)
			case "down":
				m.concepts.detailScroll++
			case "pgup":
				m.concepts.detailScroll = max(0, m.concepts.detailScroll-max(1, m.height-8))
			case "pgdown":
				m.concepts.detailScroll += max(1, m.height-8)
			}
			return m, nil
		}
		if m.screen == 2 && focus == "list" {
			switch key {
			case "up", "down":
				delta := 1
				if key == "up" {
					delta = -1
				}
				m.concepts.move(delta)
				return m, nil
			case "enter":
				return m, m.lookupConcept()
			}
		}
		if m.screen == 2 && focus == "prompt" && key == "enter" {
			return m, m.lookupConcept()
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
			cmd := m.enterScreen((m.screen + delta) % 5)
			m.focuses[m.screen] = len(m.focusNames()) - 1
			return m, cmd
		}
		if m.screen == 3 && focus == "refresh" && key == "enter" {
			return m, m.refreshLanguages()
		}
		if m.screen == 1 && key == "enter" && (focus == "language" || focus == "mode") {
			if focus == "language" {
				if m.forms.languageID() == "pl" {
					m.forms.language = "en"
				} else {
					m.forms.language = "pl"
				}
			} else {
				m.forms.generate = !m.forms.generate
			}
			m.forms.revision++
			m.scroll[1] = 0
			return m, nil
		}
		if (focus == "direction" && key == "enter") || (m.screen == 0 && key == "alt+d") {
			m.from, m.to = m.to, m.from
			m.drafts[0].revision++
			return m, nil
		}
		if m.screen == 2 && focus == "result" {
			if key == "enter" && !m.concepts.detail {
				return m, m.lookupConcept()
			}
			if !m.concepts.detail {
				if key == "down" {
					m.concepts.listScroll++
				} else if key == "up" {
					m.concepts.listScroll = max(0, m.concepts.listScroll-1)
				} else if key == "pgdown" {
					m.concepts.listScroll += max(1, m.height-8)
				} else if key == "pgup" {
					m.concepts.listScroll = max(0, m.concepts.listScroll-max(1, m.height-8))
				}
			}
			return m, nil
		}
		if m.screen == 1 && focus == "result" && key == "enter" {
			m.formsDetails = !m.formsDetails
			return m, nil
		}
		if m.screen == 4 && key == "ctrl+s" {
			return m, m.savePreferences()
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
		if key == "enter" {
			if m.screen == 4 {
				switch focus {
				case "prompt", "validate":
					return m, m.checkRuntime()
				case "level":
					switch m.preferences.level {
					case "simple":
						m.preferences.level = "normal"
					case "normal":
						m.preferences.level = "technical"
					default:
						m.preferences.level = "simple"
					}
				case "default_direction":
					if m.preferences.defaultFrom == "en" {
						m.preferences.defaultFrom = "pl"
					} else {
						m.preferences.defaultFrom = "en"
					}
				case "save":
					return m, m.savePreferences()
				case "cancel":
					m.cancelPreferencesDraft()
				case "store":
					return m, nil
				}
				return m, nil
			}
			if m.screen == 1 && (focus == "prompt" || focus == "features" || focus == "run") {
				return m, m.submitForms()
			}
			if m.screen == 0 && focus == "prompt" {
				return m, m.submitTranslate()
			}
			return m, nil
		}
		if m.screen == 1 && (focus == "prompt" || focus == "features") {
			e := m.forms.activeEditor(focus)
			before := e.revision
			e.update(v)
			if e.revision != before {
				m.forms.revision++
			}
		} else if focus == "prompt" && (m.screen == 0 || m.screen == 4) {
			before := m.drafts[4].revision
			m.drafts[m.screen].update(v)
			if m.screen == 4 && before != m.drafts[4].revision {
				m.runtimeDraftChanged()
			}
		} else if m.screen == 4 && focus == "store" {
			m.preferences.store.update(v)
		} else if m.screen == 2 && focus == "prompt" {
			before := m.concepts.filter.revision
			m.concepts.filter.update(v)
			if before != m.concepts.filter.revision {
				m.concepts.edited()
			}
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
		b += "Setup status: " + string(m.response.Failure.Status) + diagnostics(m.response.Failure.Diagnostics)
		if m.displayLevel() == "technical" {
			b += "\nTechnical response:\n" + technicalJSON(m.response)
		}
		return b
	}
	r := m.response.Result
	if r == nil {
		b += "Enter submits the prompt. Select runtime config in F5."
		if m.displayLevel() == "technical" && m.pendingTranslate != 0 {
			b += "\nTechnical response:\n" + technicalJSON(m.response)
		}
		return b
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
	b += diagnostics(r.Diagnostics) + m.previewBody() + m.candidatesBody()
	if m.displayLevel() == "technical" {
		b += "\nTechnical response:\n" + technicalJSON(m.response)
	}
	return b
}
func (m model) displayLevel() string {
	if m.preferences.active.DisplayLevel == "simple" || m.preferences.active.DisplayLevel == "normal" || m.preferences.active.DisplayLevel == "technical" {
		return m.preferences.active.DisplayLevel
	}
	return "normal"
}
func (m model) formsBody() string {
	if m.displayLevel() != "simple" || m.formsDetails || m.forms.response == nil {
		b := m.forms.body()
		if m.displayLevel() == "technical" && m.forms.response != nil {
			b += "\nTechnical FormsResult:\n" + technicalJSON(m.forms.response)
		}
		if m.displayLevel() == "simple" {
			b += "\nEnter on result toggles detailed presentation."
		}
		return b
	}
	var b strings.Builder
	fmt.Fprintf(&b, "F2 Forms · %s · language %s\n", map[bool]string{true: "Generate", false: "Analyze"}[m.forms.generate], m.forms.languageID())
	if m.forms.resultRevision != m.forms.revision {
		b.WriteString("Result is from an older draft; current draft has changed.\n")
	}
	fmt.Fprintf(&b, "Status: %s\n", plain(string(m.forms.response.Status)))
	b.WriteString(diagnostics(m.forms.response.Diagnostics))
	b.WriteByte('\n')
	fmt.Fprintf(&b, "Readings: %d\n", len(m.forms.response.Readings))
	for _, reading := range m.forms.response.Readings {
		fmt.Fprintf(&b, "  %s\n", plain(reading.Surface))
	}
	fmt.Fprintf(&b, "Forms: %d\n", len(m.forms.response.Forms))
	for _, form := range m.forms.response.Forms {
		fmt.Fprintf(&b, "  %s\n", plain(form.Surface))
	}
	b.WriteString("Enter on result toggles detailed presentation.")
	return b.String()
}
func (m model) View() tea.View {
	body := "Not implemented yet. This tab has no service actions."
	footer := "F1-F5 Screens · Tab Focus · Enter Action · Esc Back · Ctrl+C Exit"
	switch m.screen {
	case 0:
		body = m.translationBody()
	case 1:
		body = m.formsBody()
	case 2:
		body = m.concepts.body()
	case 3:
		body = m.languages.body()
	case 4:
		body = m.preferences.body(m.settingsFile, m.settingsNotice, m.runtime, m.pendingRuntime != 0)
		if !strings.HasSuffix(body, "\n") {
			body += "\n"
		}
		body += fmt.Sprintf("Runtime config path · active: %s · draft: %s\n", plain(m.runtimePath), plain(m.drafts[4].text))
		body += "Enter on Validate checks the draft path; Enter on Cancel resets unsaved preference edits.\n"
		body += "Save applies the captured preferences only after the settings file is written successfully."
	}
	if m.help {
		body = "Tab/Shift+Tab cycle controls; tabs support Left/Right.\nF1: prompt Enter translates; direction Enter or Alt+D swaps PL/EN.\nF5: prompt Enter validates/selects path; Ctrl+S explicitly saves.\nF1 candidates: Up/Down select, Enter details; Preview control calls service.\nCtrl+P opens navigation palette; Esc cancels. Result focus scrolls. Ctrl+U clears input.\nEsc closes help; Ctrl+C or Ctrl+D restores terminal and exits."
		body += "\nF2: language and Analyze/Generate mode use Enter; Generate features use explicit name=value pairs."
		body += "\nF3: edit the ConceptID filter; list Up/Down select, Enter opens details, Esc returns."
	}
	if m.palette.open {
		body = m.paletteBody()
	} else if m.notice != "" {
		body = plain(m.notice) + "\n" + body
	}
	body = ansi.Hardwrap(body, max(1, m.width), true)
	lines := strings.Split(body, "\n")
	start := min(m.scroll[m.screen], max(0, len(lines)-max(1, m.height-6)))
	if m.screen == 2 && (m.palette.open || m.help) {
		start = 0
	} else if m.screen == 2 {
		conceptScroll := m.concepts.listScroll
		if m.concepts.detail {
			conceptScroll = m.concepts.detailScroll
		} else if m.focus() == "list" {
			for i, line := range lines {
				if strings.HasPrefix(line, ">> ") {
					conceptScroll = max(0, i-3)
					break
				}
			}
		}
		start = min(conceptScroll, max(0, len(lines)-max(1, m.height-6)))
	}
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
	if m.screen == 1 {
		prompt = m.forms.activeEditor(m.focus()).visible(m.width - 4)
	} else if m.screen == 4 && m.focus() == "store" {
		prompt = m.preferences.store.visible(m.width - 4)
	} else if m.screen == 3 {
		prompt = "F4 inventory and runtime readiness · Enter on Refresh to recheck"
	} else if m.screen == 2 {
		prompt = m.concepts.filter.visible(m.width - 4)
	} else if m.screen != 0 && m.screen != 4 {
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
