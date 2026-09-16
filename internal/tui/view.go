package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"vroom/internal/about"
	"vroom/internal/humanize"
	"vroom/internal/theme"
)

var (
	orange = lipgloss.Color("#F59E0B")
	red    = lipgloss.Color("#EF4444")
	green  = lipgloss.Color("#22C55E")
	yellow = lipgloss.Color("#EAB308")
	muted  = lipgloss.Color("#94A3B8")
	white  = lipgloss.Color("#E2E8F0")
)

func (m Model) View() string {
	if m.width == 0 {
		return "loading…"
	}
	if m.quitting {
		return ""
	}

	switch m.view {
	case viewHelp:
		return m.frame(m.helpView())
	case viewSuggest:
		return m.frame(m.suggestView())
	case viewDetail:
		return m.frame(m.detailView())
	case viewConfirm:
		return m.frame(m.confirmView())
	default:
		return m.frame(m.treeView())
	}
}

func (m Model) frame(body string) string {
	header := m.headerBar()
	footer := m.footerBar()
	avail := m.height - lipgloss.Height(header) - lipgloss.Height(footer)
	if avail < 5 {
		avail = 5
	}
	body = lipgloss.NewStyle().Height(avail).MaxHeight(avail).Width(m.width).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, header, body, footer)
}

func (m Model) headerBar() string {
	path := m.rootPath
	if m.current != nil {
		path = m.current.Path
	}
	left := lipgloss.NewStyle().Bold(true).Foreground(orange).Render(" VROOM ")
	mid := lipgloss.NewStyle().Foreground(white).Render(path)
	right := lipgloss.NewStyle().Foreground(muted).Render(about.Author)
	if m.scanning {
		right = m.spinner.View() + " scanning"
	}
	gap := m.width - lipgloss.Width(left) - lipgloss.Width(mid) - lipgloss.Width(right) - 2
	if gap < 1 {
		gap = 1
	}
	return lipgloss.NewStyle().Background(lipgloss.Color("#1E293B")).Width(m.width).
		Render(left + mid + strings.Repeat(" ", gap) + right)
}

func (m Model) footerBar() string {
	help := "↑↓ navigate  ↵ open  ⌫ up  space mark  d clean  i info  s suggest  r refresh  ? help  q quit"
	st := lipgloss.NewStyle().Foreground(muted).Width(m.width)
	if m.status != "" {
		help = m.status + "  ·  " + help
	}
	if len(help) > m.width {
		help = help[:m.width]
	}
	return st.Render(help)
}

func (m Model) treeView() string {
	rightW := theme.FrameWidth + 6
	if rightW > m.width*3/5 {
		rightW = m.width * 3 / 5
	}
	if rightW < 38 {
		rightW = 38
	}
	leftW := m.width - rightW - 1
	if leftW < 20 {
		leftW = 20
		rightW = m.width - leftW - 1
	}
	left := m.treePane(leftW)
	right := m.mascotPane(rightW)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, right)
}

func (m Model) treePane(width int) string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		Width(width - 2).
		Height(m.height - 5)
	kids := m.children()
	innerH := m.height - 7
	if innerH < 3 {
		innerH = 3
	}
	start := m.offset
	if m.cursor < start {
		start = m.cursor
	}
	if m.cursor >= start+innerH {
		start = m.cursor - innerH + 1
	}
	if start < 0 {
		start = 0
	}

	var b strings.Builder
	title := " folders "
	if m.current != nil {
		title = fmt.Sprintf(" %s  %s ", humanize.Bytes(m.current.Size), m.current.Name)
	}
	b.WriteString(lipgloss.NewStyle().Foreground(orange).Bold(true).Render(title) + "\n")

	if m.scanning && m.current == nil {
		b.WriteString("  " + m.spinner.View() + " walking the disk…\n")
		return style.Render(b.String())
	}
	if len(kids) == 0 {
		b.WriteString(lipgloss.NewStyle().Foreground(muted).Render("  (empty)"))
		return style.Render(b.String())
	}
	for i := start; i < len(kids) && i < start+innerH-1; i++ {
		n := kids[i]
		mark := " "
		if m.marked[n.Path] {
			mark = "*"
		}
		kind := " "
		if n.IsDir {
			kind = "/"
		}
		line := fmt.Sprintf("%s %8s  %s%s", mark, humanize.Bytes(n.Size), n.Name, kind)
		if n.Err != nil {
			line += "  !"
		}
		if len(line) > width-4 {
			line = line[:width-4]
		}
		if i == m.cursor {
			line = lipgloss.NewStyle().Background(lipgloss.Color("#334155")).Foreground(orange).Bold(true).Width(width - 4).Render(line)
		} else if m.marked[n.Path] {
			line = lipgloss.NewStyle().Foreground(yellow).Render(line)
		}
		b.WriteString(" " + line + "\n")
	}
	return style.Render(b.String())
}

func (m Model) mascotPane(width int) string {
	stColor := stateColor(m.anim.State())
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(stColor).
		Width(width - 2).
		Height(m.height - 5)

	art := m.anim.Frame()
	summary := humanize.DiskSummary(m.disk.Avail, m.disk.Total)
	phrase := `"` + m.anim.Phrase() + `"`
	bar := usageBar(m.anim.UsagePct(), width-6)
	state := strings.ToUpper(m.anim.State().String())

	body := lipgloss.JoinVertical(lipgloss.Center,
		lipgloss.NewStyle().Foreground(stColor).Bold(true).Render(" "+state+" "),
		art,
		"",
		lipgloss.NewStyle().Foreground(white).Width(width-6).Align(lipgloss.Center).Render(summary),
		bar,
		lipgloss.NewStyle().Foreground(orange).Italic(true).Width(width-6).Align(lipgloss.Center).Render(phrase),
	)
	return style.Render(body)
}

func stateColor(s theme.State) lipgloss.Color {
	switch s {
	case theme.StateRelaxed:
		return green
	case theme.StateSarcastic:
		return orange
	case theme.StateSweating:
		return yellow
	case theme.StatePanicking, theme.StateCritical:
		return red
	default:
		return muted
	}
}

func usageBar(pct float64, width int) string {
	if width < 8 {
		width = 8
	}
	inner := width - 2
	filled := int(pct / 100 * float64(inner))
	if filled > inner {
		filled = inner
	}
	if filled < 0 {
		filled = 0
	}
	col := green
	if pct > 80 {
		col = red
	} else if pct > 50 {
		col = orange
	} else if pct > 30 {
		col = yellow
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", inner-filled)
	return lipgloss.NewStyle().Foreground(col).Render(fmt.Sprintf("%s %5.1f%% used", bar, pct))
}

func (m Model) detailView() string {
	n := m.detail
	if n == nil {
		return "no selection"
	}
	kind := "file"
	if n.IsDir {
		kind = "directory"
	}
	lines := []string{
		lipgloss.NewStyle().Bold(true).Foreground(orange).Render(n.Name),
		"",
		fmt.Sprintf("path:      %s", n.Path),
		fmt.Sprintf("type:      %s", kind),
		fmt.Sprintf("size:      %s (%d bytes)", humanize.Bytes(n.Size), n.Size),
		fmt.Sprintf("items:     %s", humanize.Count(n.Count, "item", "items")),
		fmt.Sprintf("mode:      %s", n.Mode.String()),
		fmt.Sprintf("owner:     %s", n.Owner),
		fmt.Sprintf("modified:  %s", n.ModTime.Format("2006-01-02 15:04:05")),
		fmt.Sprintf("accessed:  %s", n.Atime.Format("2006-01-02 15:04:05")),
	}
	if n.Err != nil {
		lines = append(lines, fmt.Sprintf("warning:   %s", n.Err.Error()))
	}
	lines = append(lines, "", lipgloss.NewStyle().Foreground(muted).Render("esc to go back"))
	return lipgloss.NewStyle().Padding(1, 2).Render(strings.Join(lines, "\n"))
}

func (m Model) confirmView() string {
	paths := m.markedList()
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(red).Render("Review quarantine") + "\n\n")
	b.WriteString("These will move to ~/.vroom/quarantine (not hard-deleted):\n\n")
	for _, p := range paths {
		b.WriteString("  • " + p + "\n")
	}
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Foreground(yellow).Render("y / enter  confirm    n / esc  cancel"))
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m Model) suggestView() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(orange).Render("Suggestions") + "\n\n")
	for _, s := range m.suggests {
		b.WriteString(lipgloss.NewStyle().Bold(true).Render("• "+s.Title) + "\n")
		b.WriteString("  " + s.Voice + "\n")
		if s.Detail != "" {
			b.WriteString(lipgloss.NewStyle().Foreground(muted).Render("  "+s.Detail) + "\n")
		}
		b.WriteString("\n")
	}
	b.WriteString(lipgloss.NewStyle().Foreground(muted).Render("esc to go back"))
	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}

func (m Model) helpView() string {
	text := `
Vroom — VisualRoom
` + about.Credit() + `

  ↑ ↓ / j k     move selection
  enter / l     drill into folder (or file details)
  backspace / h go up
  space         toggle mark
  d / c         review marked items for quarantine
  i             details for selection
  s             suggestions
  r             rescan this folder
  ?             this help
  q             quit

Cleanup is never a hard delete. Files move to ~/.vroom/quarantine
with a TTL, and every action is appended to ~/.vroom/audit.log.

vroom clean is dry-run unless you pass --force.
`
	return lipgloss.NewStyle().Padding(1, 2).Foreground(white).Render(text)
}
