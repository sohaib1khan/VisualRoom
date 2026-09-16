package tui

import (
	"fmt"
	"path/filepath"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"vroom/internal/cleanup"
	"vroom/internal/scanner"
	"vroom/internal/suggest"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		m.anim.Advance()
		return m, tick(m.anim.TickInterval())

	case scanMsg:
		m.scanning = false
		if msg.err != nil {
			m.err = msg.err
			m.status = msg.err.Error()
			return m, nil
		}
		m.err = nil
		m.current = msg.node
		m.disk = msg.disk
		m.anim.SetFreePct(msg.disk.FreePct())
		m.cursor = 0
		m.offset = 0
		m.status = fmt.Sprintf("scanned %s", msg.node.Path)
		m.suggests = suggest.Analyze(m.cfg, m.current, m.disk)
		return m, nil

	case cleanMsg:
		if msg.err != nil {
			m.status = "cleanup failed: " + msg.err.Error()
			m.view = viewTree
			return m, nil
		}
		n := 0
		if msg.res != nil {
			n = msg.res.Moved
		}
		m.status = fmt.Sprintf("quarantined %d item(s)", n)
		m.marked = map[string]bool{}
		m.view = viewTree
		m.scanning = true
		return m, tea.Batch(m.scan(m.current.Path), m.spinner.Tick)

	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.quitting {
		return m, tea.Quit
	}

	switch m.view {
	case viewHelp:
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "?" || msg.String() == "enter" {
			m.view = viewTree
		}
		return m, nil
	case viewSuggest:
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "s" {
			m.view = viewTree
		}
		return m, nil
	case viewDetail:
		if msg.String() == "esc" || msg.String() == "q" || msg.String() == "i" || msg.String() == "enter" {
			m.view = viewTree
			m.detail = nil
		}
		return m, nil
	case viewConfirm:
		switch msg.String() {
		case "esc", "n", "q":
			m.view = viewTree
			m.status = "cleanup cancelled"
		case "y", "enter":
			paths := m.markedList()
			if len(paths) == 0 {
				m.view = viewTree
				return m, nil
			}
			if err := cleanup.RequireElevatedOK(false); err != nil {
				m.status = err.Error()
				m.view = viewTree
				return m, nil
			}
			if scanner.InContainer() {
				if scanner.IsReadOnlyMount(paths[0]) {
					m.status = "docker read-only mount: cleanup blocked"
					m.view = viewTree
					return m, nil
				}
			}
			cfg := m.cfg
			return m, func() tea.Msg {
				res, err := cleanup.Quarantine(cfg, paths, false, false)
				return cleanMsg{res: res, err: err}
			}
		}
		return m, nil
	}

	switch msg.String() {
	case "ctrl+c", "q":
		m.quitting = true
		return m, tea.Quit
	case "?":
		m.view = viewHelp
		return m, nil
	case "s":
		m.suggests = suggest.Analyze(m.cfg, m.current, m.disk)
		m.view = viewSuggest
		return m, nil
	case "i":
		if sel := m.selected(); sel != nil {
			m.detail = sel
			m.view = viewDetail
		}
		return m, nil
	case "r", "ctrl+r":
		if m.current != nil && !m.scanning {
			m.scanning = true
			m.status = "rescanning…"
			return m, tea.Batch(m.scan(m.current.Path), m.spinner.Tick)
		}
		return m, nil
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case "down", "j":
		if m.cursor < len(m.children())-1 {
			m.cursor++
		}
		return m, nil
	case "pgup":
		m.cursor -= 10
		m.clampCursor()
		return m, nil
	case "pgdown":
		m.cursor += 10
		m.clampCursor()
		return m, nil
	case "home", "g":
		m.cursor = 0
		return m, nil
	case "end", "G":
		m.cursor = len(m.children()) - 1
		m.clampCursor()
		return m, nil
	case "enter", "l", "right":
		sel := m.selected()
		if sel == nil {
			return m, nil
		}
		if sel.IsDir {
			m.stack = append(m.stack, m.current.Path)
			m.scanning = true
			m.status = "opening " + sel.Name
			return m, tea.Batch(m.scan(sel.Path), m.spinner.Tick)
		}
		m.detail = sel
		m.view = viewDetail
		return m, nil
	case "backspace", "h", "left":
		if len(m.stack) == 0 {
			parent := m.parentPath()
			if m.current != nil && parent != m.current.Path {
				m.scanning = true
				return m, tea.Batch(m.scan(parent), m.spinner.Tick)
			}
			return m, nil
		}
		prev := m.stack[len(m.stack)-1]
		m.stack = m.stack[:len(m.stack)-1]
		m.scanning = true
		return m, tea.Batch(m.scan(prev), m.spinner.Tick)
	case " ":
		if sel := m.selected(); sel != nil {
			m.marked[sel.Path] = !m.marked[sel.Path]
			if !m.marked[sel.Path] {
				delete(m.marked, sel.Path)
			}
			if m.cursor < len(m.children())-1 {
				m.cursor++
			}
		}
		return m, nil
	case "d":
		if sel := m.selected(); sel != nil {
			m.marked[sel.Path] = true
		}
		if len(m.marked) == 0 {
			m.status = "nothing marked"
			return m, nil
		}
		m.view = viewConfirm
		return m, nil
	case "c":
		if len(m.marked) == 0 {
			m.status = "nothing marked — space or d to mark"
			return m, nil
		}
		m.view = viewConfirm
		return m, nil
	case "esc":
		if len(m.marked) > 0 {
			m.marked = map[string]bool{}
			m.status = "cleared marks"
		}
		return m, nil
	}

	_ = filepath.Separator
	return m, nil
}
