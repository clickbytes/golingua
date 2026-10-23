package tui

import (
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"strings"
	"unicode"
)

type editor struct {
	text     string
	cursor   int
	revision uint64
}

// plain removes terminal escape sequences and control characters from data.
func plain(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, ansi.Strip(s))
}
func (e *editor) set(s string) { e.text = plain(s); e.cursor = len([]rune(e.text)); e.revision++ }
func (e *editor) insert(s string) {
	s = plain(s)
	if len(e.text)+len(s) > 64<<10 {
		return
	}
	r := []rune(e.text)
	e.text = string(r[:e.cursor]) + s + string(r[e.cursor:])
	e.cursor += len([]rune(s))
	e.revision++
}
func (e *editor) update(k tea.KeyPressMsg) {
	r := []rune(e.text)
	switch k.String() {
	case "left":
		if e.cursor > 0 {
			e.cursor--
		}
	case "right":
		if e.cursor < len(r) {
			e.cursor++
		}
	case "home", "ctrl+a":
		e.cursor = 0
	case "end", "ctrl+e":
		e.cursor = len(r)
	case "backspace":
		if e.cursor > 0 {
			e.text = string(r[:e.cursor-1]) + string(r[e.cursor:])
			e.cursor--
			e.revision++
		}
	case "delete":
		if e.cursor < len(r) {
			e.text = string(r[:e.cursor]) + string(r[e.cursor+1:])
			e.revision++
		}
	case "ctrl+u":
		e.set("")
	default:
		if k.Key().Text != "" {
			e.insert(k.Key().Text)
		}
	}
}
func (e editor) visible(width int) string {
	r := []rune(e.text)
	before := string(r[:e.cursor])
	after := string(r[e.cursor:])
	if width < 4 {
		return "|"
	}
	for ansi.StringWidth(before) > width-3 {
		before = string([]rune(before)[1:])
	}
	return before + "|" + after
}
