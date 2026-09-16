package theme

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestStateFromFreePct(t *testing.T) {
	th := struct{ a, b, c, d float64 }{50, 20, 10, 3}
	if StateFromFreePct(80, th.a, th.b, th.c, th.d) != StateRelaxed {
		t.Fatal("80% free should be relaxed")
	}
	if StateFromFreePct(30, th.a, th.b, th.c, th.d) != StateSarcastic {
		t.Fatal("30% free should be sarcastic")
	}
	if StateFromFreePct(15, th.a, th.b, th.c, th.d) != StateSweating {
		t.Fatal("15% free should be sweating")
	}
	if StateFromFreePct(5, th.a, th.b, th.c, th.d) != StatePanicking {
		t.Fatal("5% free should be panicking")
	}
	if StateFromFreePct(1, th.a, th.b, th.c, th.d) != StateCritical {
		t.Fatal("1% free should be critical")
	}
}

func TestFramesNonEmpty(t *testing.T) {
	for _, st := range AllStates() {
		frames := Frames(st)
		if len(frames) < 6 {
			t.Fatalf("%s: expected a looping animation, got %d frames", st, len(frames))
		}
		for i, f := range frames {
			lines := strings.Split(f, "\n")
			if len(lines) != FrameHeight {
				t.Fatalf("%s frame %d: height %d want %d", st, i, len(lines), FrameHeight)
			}
			for _, line := range lines {
				if w := lipgloss.Width(line); w != FrameWidth {
					t.Fatalf("%s frame %d: width %d want %d (%q)", st, i, w, FrameWidth, line)
				}
			}
			_ = Paint(f, st, i)
		}
	}
}
