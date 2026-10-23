package tui

import (
	tea "charm.land/bubbletea/v2"
	"strings"
)

var navigationCommands = [...]string{"translate", "forms", "concepts", "languages", "settings"}

type paletteState struct {
	open     bool
	input    editor
	selected int
}

func (p paletteState) matches() []int {
	text := p.input.text
	if !strings.HasPrefix(text, "/") || strings.ContainsAny(text, " \t\n") {
		return nil
	}
	prefix := strings.TrimPrefix(text, "/")
	var matches []int
	for i, c := range navigationCommands {
		if strings.HasPrefix(c, prefix) {
			matches = append(matches, i)
		}
	}
	return matches
}
func (m *model) openPalette() {
	m.palette = paletteState{open: true, selected: -1}
	m.palette.input.set("/")
	m.notice = ""
}
func (m *model) paletteKey(k tea.KeyPressMsg) {
	switch k.String() {
	case "esc":
		m.palette.open = false
		m.notice = ""
		return
	case "up", "down":
		matches := m.palette.matches()
		if len(matches) == 0 {
			return
		}
		n := len(matches)
		if k.String() == "down" {
			m.palette.selected = (m.palette.selected + 1) % n
		} else {
			if m.palette.selected < 0 {
				m.palette.selected = 0
			}
			m.palette.selected = (m.palette.selected + n - 1) % n
		}
		return
	case "enter":
		matches := m.palette.matches()
		target := -1
		for _, i := range matches {
			if m.palette.input.text == "/"+navigationCommands[i] {
				target = i
			}
		}
		if target < 0 && m.palette.selected >= 0 && m.palette.selected < len(matches) {
			target = matches[m.palette.selected]
		}
		if target < 0 {
			m.notice = "Choose a registered command; arguments are not supported."
			return
		}
		m.screen = target
		m.palette.open = false
		m.notice = ""
		return
	}
	m.palette.input.update(k)
	m.palette.selected = -1
	m.notice = ""
}
func (m model) paletteBody() string {
	b := "Navigation only · Up/Down choose · Enter open · Esc cancel\n"
	matches := m.palette.matches()
	if len(matches) == 0 {
		b += "No registered command matches.\n"
	}
	for row, i := range matches {
		mark := "  "
		if row == m.palette.selected {
			mark = ">>"
		}
		b += mark + " /" + navigationCommands[i]
		if i == m.screen {
			b += " (current)"
		}
		b += "\n"
	}
	return b + plain(m.notice)
}
