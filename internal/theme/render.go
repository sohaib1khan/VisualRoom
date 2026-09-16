package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	FrameWidth  = 34
	FrameHeight = 16
)

func padFrame(s string, width, height int) string {
	s = strings.TrimPrefix(s, "\n")
	s = strings.TrimRight(s, "\n")
	lines := strings.Split(s, "\n")
	out := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = strings.TrimRight(lines[i], " ")
		}
		out = append(out, padLine(line, width))
	}
	return strings.Join(out, "\n")
}

func padLine(s string, width int) string {
	r := []rune(s)
	for len(r) > 0 && lipgloss.Width(string(r)) > width {
		r = r[:len(r)-1]
	}
	s = string(r)
	for lipgloss.Width(s) < width {
		s += " "
	}
	return s
}

func Normalize(frames []string, width, height int) []string {
	out := make([]string, len(frames))
	for i, f := range frames {
		out[i] = padFrame(f, width, height)
	}
	return out
}

func Paint(frame string, state State, idx int) string {
	lines := strings.Split(frame, "\n")
	out := make([]string, len(lines))
	for i, line := range lines {
		var b strings.Builder
		for _, r := range line {
			if r == ' ' {
				b.WriteByte(' ')
				continue
			}
			b.WriteString(glyphStyle(r, state, idx).Render(string(r)))
		}
		out[i] = b.String()
	}
	return strings.Join(out, "\n")
}

func glyphStyle(r rune, state State, idx int) lipgloss.Style {
	st := lipgloss.NewStyle()
	metal := lipgloss.Color("#C5CDD6")
	dark := lipgloss.Color("#7D8793")
	switch state {
	case StateRelaxed:
		metal = "#D5DCE3"
	case StateSarcastic:
		metal = "#D6C4A8"
		dark = "#9A8164"
	case StateSweating:
		metal = "#E8E392"
		dark = "#A3A15C"
	case StatePanicking:
		metal = "#E8A598"
		dark = "#B4533A"
	case StateCritical:
		metal = "#6B7280"
		dark = "#4B5563"
		if idx%2 == 1 {
			metal = "#4B5563"
		}
	}

	fireA := lipgloss.Color("#F97316")
	fireB := lipgloss.Color("#FACC15")
	if idx%2 == 1 {
		fireA, fireB = fireB, fireA
	}

	switch r {
	case 'o', '●':
		c := lipgloss.Color("#E11D48")
		if state == StateCritical {
			c = "#6B7280"
		}
		return st.Foreground(c).Bold(true)
	case '0', 'O', 'Q', '-':
		c := lipgloss.Color("#FFE566")
		if state == StateSweating {
			c = "#FFF1A8"
		}
		if state == StatePanicking {
			c = "#FFF7C2"
		}
		if state == StateCritical {
			c = "#EF4444"
		}
		return st.Foreground(c).Bold(true)
	case 'x', 'X':
		return st.Foreground(lipgloss.Color("#EF4444")).Bold(true)
	case '▓':
		return st.Foreground(lipgloss.Color("#111827"))
	case '▒':
		if state == StateSarcastic {
			return st.Foreground(lipgloss.Color("#92400E"))
		}
		if state == StatePanicking {
			return st.Foreground(fireA).Bold(true)
		}
		return st.Foreground(lipgloss.Color("#F59E0B")).Bold(true)
	case '░', '.', '`', '\'', '°', '~':
		switch state {
		case StateSarcastic:
			return st.Foreground(lipgloss.Color("#E2E8F0")).Faint(true)
		case StateSweating:
			return st.Foreground(lipgloss.Color("#7DD3FC"))
		case StatePanicking:
			return st.Foreground(fireB).Bold(true)
		case StateCritical:
			return st.Foreground(lipgloss.Color("#22C55E"))
		default:
			return st.Foreground(lipgloss.Color("#FDE68A"))
		}
	case '^', '*', '▲', '△':
		if state == StatePanicking || state == StateCritical {
			return st.Foreground(fireA).Bold(true)
		}
		return st.Foreground(lipgloss.Color(metal))
	case '▀', '▄', '█', '▌', '▐':
		return st.Foreground(lipgloss.Color(metal)).Bold(true)
	case '═', '─', '│', '┌', '┐', '└', '┘', '╭', '╮', '╰', '╯',
		'║', '╔', '╗', '╚', '╝', '┬', '┴', '├', '┤', '┼', '/', '\\', '|', '_', '+':
		return st.Foreground(lipgloss.Color(dark))
	case 'K', 'I', 'L', 'A', 'H', 'U', 'M', 'N', 'S', ' ':
		if state == StateCritical {
			return st.Foreground(lipgloss.Color("#4ADE80")).Bold(true)
		}
	}
	if state == StateCritical && r >= 'A' && r <= 'Z' {
		return st.Foreground(lipgloss.Color("#4ADE80")).Bold(true)
	}
	return st.Foreground(lipgloss.Color(metal))
}
