package tui

import (
	"context"
	"path/filepath"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"vroom/internal/cleanup"
	"vroom/internal/config"
	"vroom/internal/scanner"
	"vroom/internal/suggest"
	"vroom/internal/theme"
)

type viewState int

const (
	viewTree viewState = iota
	viewDetail
	viewConfirm
	viewSuggest
	viewHelp
)

type scanMsg struct {
	node *scanner.Node
	disk scanner.DiskUsage
	err  error
}

type cleanMsg struct {
	res *cleanup.Result
	err error
}

type tickMsg time.Time

type Model struct {
	cfg     *config.Config
	walker  *scanner.Walker
	anim    *theme.Animator
	spinner spinner.Model

	width  int
	height int

	rootPath string
	current  *scanner.Node
	stack    []string
	cursor   int
	offset   int

	disk     scanner.DiskUsage
	marked   map[string]bool
	view     viewState
	detail   *scanner.Node
	suggests []suggest.Suggestion
	status   string
	err      error
	scanning bool
	quitting bool
}

func New(root string, cfg *config.Config) Model {
	sp := spinner.New()
	sp.Spinner = spinner.Dot
	return Model{
		cfg:      cfg,
		walker:   scanner.New(scanner.DefaultOptions()),
		anim:     theme.NewAnimator(cfg.Animation, cfg.Thresholds),
		spinner:  sp,
		rootPath: root,
		marked:   map[string]bool{},
		scanning: true,
		status:   "scanning…",
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.scan(m.rootPath), m.spinner.Tick, tick(m.anim.TickInterval()))
}

func (m Model) scan(path string) tea.Cmd {
	w := m.walker
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		disk, _ := scanner.Usage(path)
		node, err := w.ScanDir(ctx, path)
		_ = w.SaveCache()
		if err == nil && node != nil {
			_ = cleanup.AppendAudit("SCAN root=" + path)
		}
		return scanMsg{node: node, disk: disk, err: err}
	}
}

func tick(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func (m Model) children() []*scanner.Node {
	if m.current == nil {
		return nil
	}
	return m.current.Children
}

func (m Model) selected() *scanner.Node {
	kids := m.children()
	if m.cursor < 0 || m.cursor >= len(kids) {
		return nil
	}
	return kids[m.cursor]
}

func (m *Model) clampCursor() {
	n := len(m.children())
	if n == 0 {
		m.cursor = 0
		return
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= n {
		m.cursor = n - 1
	}
}

func (m Model) markedList() []string {
	var out []string
	for p, ok := range m.marked {
		if ok {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}

func (m Model) parentPath() string {
	if m.current == nil {
		return m.rootPath
	}
	parent := filepath.Dir(m.current.Path)
	if parent == m.current.Path {
		return m.current.Path
	}
	return parent
}

func Start(root string) error {
	cfg := config.MustLoad()
	p := tea.NewProgram(New(root, cfg), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
