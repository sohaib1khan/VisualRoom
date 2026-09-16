package cmd

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
	"vroom/internal/about"
	"vroom/internal/config"
	"vroom/internal/theme"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Preview health-mascot animations (1–5 switch mood, q quit)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		p := tea.NewProgram(newPreviewModel(cfg), tea.WithAltScreen())
		_, err := p.Run()
		return err
	},
}

type previewTick time.Time

type previewModel struct {
	anim   *theme.Animator
	th     config.Thresholds
	state  theme.State
	width  int
	height int
	pct    float64
}

func newPreviewModel(cfg *config.Config) previewModel {
	a := theme.NewAnimator(cfg.Animation, cfg.Thresholds)
	pct := 80.0
	a.SetFreePct(pct)
	return previewModel{anim: a, th: cfg.Thresholds, state: theme.StateRelaxed, pct: pct}
}

func (m previewModel) Init() tea.Cmd {
	return previewTickCmd(m.anim.TickInterval())
}

func previewTickCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg { return previewTick(t) })
}

func (m previewModel) setState(st theme.State) previewModel {
	m.state = st
	switch st {
	case theme.StateRelaxed:
		m.pct = 80
	case theme.StateSarcastic:
		m.pct = 35
	case theme.StateSweating:
		m.pct = 15
	case theme.StatePanicking:
		m.pct = 6
	default:
		m.pct = 1
	}
	m.anim.SetFreePct(m.pct)
	return m
}

func (m previewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		return m, nil
	case previewTick:
		m.anim.Advance()
		return m, previewTickCmd(m.anim.TickInterval())
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "1":
			m = m.setState(theme.StateRelaxed)
		case "2":
			m = m.setState(theme.StateSarcastic)
		case "3":
			m = m.setState(theme.StateSweating)
		case "4":
			m = m.setState(theme.StatePanicking)
		case "5":
			m = m.setState(theme.StateCritical)
		case "n", "right", "l":
			m = m.setState(theme.AllStates()[(int(m.state)+1)%5])
		case "p", "left", "h":
			m = m.setState(theme.AllStates()[(int(m.state)+4)%5])
		}
	}
	return m, nil
}

func (m previewModel) View() string {
	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B")).
		Render(" VROOM  ·  mood preview ")
	mood := strings.ToUpper(m.state.String())
	meta := fmt.Sprintf("%s   simulated free space %.0f%%", mood, m.pct)
	help := "1 relaxed  2 sarcastic  3 sweat  4 fire  5 critical   n/p cycle   q quit"
	art := m.anim.Frame()
	phrase := `"` + m.anim.Phrase() + `"`
	body := lipgloss.JoinVertical(lipgloss.Center,
		title,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#E2E8F0")).Render(meta),
		"",
		art,
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Italic(true).Render(phrase),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#94A3B8")).Render(help),
		"",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#64748B")).Render(about.Credit()),
	)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, body)
}
