package theme

import "strings"

func Frames(state State) []string {
	n := 8
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = padFrame(draw(state, i), FrameWidth, FrameHeight)
	}
	return out
}

type canvas [][]rune

func newCanvas() canvas {
	c := make(canvas, FrameHeight)
	for i := range c {
		c[i] = []rune(spaces(FrameWidth))
	}
	return c
}

func spaces(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

func (c canvas) blit(x, y int, lines ...string) {
	for dy, line := range lines {
		for dx, r := range []rune(line) {
			if r == ' ' {
				continue
			}
			px, py := x+dx, y+dy
			if py >= 0 && py < len(c) && px >= 0 && px < len(c[py]) {
				c[py][px] = r
			}
		}
	}
}

func (c canvas) blitSoft(x, y int, lines ...string) {
	for dy, line := range lines {
		for dx, r := range []rune(line) {
			if r == ' ' {
				continue
			}
			px, py := x+dx, y+dy
			if py >= 0 && py < len(c) && px >= 0 && px < len(c[py]) && c[py][px] == ' ' {
				c[py][px] = r
			}
		}
	}
}

func (c canvas) string() string {
	lines := make([]string, len(c))
	for i, row := range c {
		lines[i] = string(row)
	}
	return strings.Join(lines, "\n")
}

func draw(state State, i int) string {
	c := newCanvas()
	x, y := 0, 1
	pose := "beer"
	eyes := "open"
	fx := "none"

	switch state {
	case StateRelaxed:
		switch i % 8 {
		case 0:
			pose, eyes = "beer", "open"
		case 1:
			x, pose, eyes = -1, "leftup", "open"
		case 2:
			y, pose, eyes = 0, "out", "happy"
		case 3:
			x, pose, eyes = 1, "toast", "open"
		case 4:
			pose, eyes = "beer", "blink"
		case 5:
			x, pose, eyes = -1, "outbeer", "open"
		case 6:
			pose, eyes = "sip", "happy"
		case 7:
			y, pose, eyes = 0, "toast", "happy"
		}
	case StateSarcastic:
		x, pose, fx = 1, "cigar", "smoke"
		eyes = "side"
		switch i % 8 {
		case 1, 5:
			eyes = "side2"
		case 4:
			eyes = "blink"
		}
	case StateSweating:
		walk := []int{0, -1, -2, -1, 0, 1, 2, 1}
		x = walk[i%8]
		fx, eyes = "sweat", "wide"
		if i%2 == 0 {
			pose = "paceL"
		} else {
			pose = "paceR"
		}
	case StatePanicking:
		shake := []int{0, -2, 1, -1, 2, -2, 1, 0}
		x = shake[i%8]
		if i%2 == 1 {
			y = 2
		}
		pose, eyes, fx = "panic", "panic", "fire"
	case StateCritical:
		y = 2
		if i%2 == 1 {
			y = 3
		}
		pose, eyes, fx = "dead", "dead", "marquee"
	}

	c.blit(x, y, poseArt(pose)...)
	c.blit(x, y, visor(eyes)...)
	drawFX(c, x, y, fx, i)
	return c.string()
}

func visor(kind string) []string {
	// Overlay onto row 4 of the sprite (head visor). Leading spaces match poseArt.
	line := "         █▓▓0▓▓0▓▓█"
	switch kind {
	case "blink":
		line = "         █▓▓----▓▓█"
	case "side":
		line = "         █▓▓0▓▓-▓▓█"
	case "side2":
		line = "         █▓▓-▓▓0▓▓█"
	case "wide":
		line = "         █▓▓O▓▓O▓▓█"
	case "panic":
		line = "         █▓▓Q▓▓Q▓▓█"
	case "dead":
		line = "         █▓▓x▓▓x▓▓█"
	case "happy":
		line = "         █▓▓0▓▓0▓▓█"
	}
	return []string{"", "", "", "", line}
}

func poseArt(pose string) []string {
	switch pose {
	case "leftup":
		return splitArt(`
              o
      █▌     ▄█▄
      █  ▄████████▄
      ▀  ██████████
         █▓▓0▓▓0▓▓█
         ▀█╰────╯█▀
       ▄▄██▀▀▀▀▀▀██▄▄
      █  █  ▄██▄  █  █▒▓
      █  █  █  █  █  ▒▒
      ▀  █  ▀██▀  █  ▀▀
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "out":
		return splitArt(`
              o
             ▄█▄
         ▄████████▄
         ██████████
         █▓▓0▓▓0▓▓█
         ▀█╰────╯█▀
       ▄▄██▀▀▀▀▀▀██▄▄
  ▀▀▀██  █  ▄██▄  █  ██▀▀▀
     ▀   █  █  █  █   ▀
         █  ▀██▀  █
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "outbeer":
		return splitArt(`
              o
             ▄█▄
         ▄████████▄
         ██████████
         █▓▓0▓▓0▓▓█
         ▀█╰────╯█▀
       ▄▄██▀▀▀▀▀▀██▄▄
  ▀▀▀██  █  ▄██▄  █  █▒▓
     ▀   █  █  █  █  ▒▒
         █  ▀██▀  █  ▀▀
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "toast":
		return splitArt(`
              o
             ▄█▄             ▒▓
         ▄████████▄          ▒▒
         ██████████         ▐█
         █▓▓0▓▓0▓▓█          ▀
         ▀█╰────╯█▀
       ▄▄██▀▀▀▀▀▀██▄▄
      █  █  ▄██▄  █  █
      █  █  █  █  █  █
      ▀  █  ▀██▀  █  ▀
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "sip":
		return splitArt(`
              o
             ▄█▄
         ▄████████▄     ▒▓
         ██████████     ▒▒
         █▓▓0▓▓0▓▓█     ▐█
         ▀█▀▀▀▀▀▀█▀
       ▄▄██▀▀▀▀▀▀██▄▄
      █  █  ▄██▄  █
      █  █  █  █  █
      ▀  █  ▀██▀  █
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "cigar":
		return splitArt(`
              o
             ▄█▄
         ▄████████▄
         ██████████
         █▓▓0▓▓0▓▓█
         ▀█──────█▀▒▒─
       ▄▄██▀▀▀▀▀▀██▄▄
      █  █  ▄██▄  █  █
      █  █  █  █  █  █
      ▀  █  ▀██▀  █  ▀
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "paceL":
		return splitArt(`
              o
             ▄█▄
         ▄████████▄
         ██████████
         █▓▓0▓▓0▓▓█
         ▀█╰────╯█▀
       ▄▄██▀▀▀▀▀▀██▄▄
  ▀▀▀██  █  ▄██▄  █  █
     ▀   █  █  █  █  █
         █  ▀██▀  █  ▀
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "paceR":
		return splitArt(`
              o
             ▄█▄
         ▄████████▄
         ██████████
         █▓▓0▓▓0▓▓█
         ▀█╰────╯█▀
       ▄▄██▀▀▀▀▀▀██▄▄
      █  █  ▄██▄  █  ██▀▀▀
      █  █  █  █  █   ▀
      ▀  █  ▀██▀  █
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "panic":
		return splitArt(`
      █       o       █
      █      ▄█▄      █
      ▀  ▄████████▄   ▀
         ██████████
         █▓▓0▓▓0▓▓█
         ▀█░▀▀▀▀░█▀
       ▄▄██▀▀▀▀▀▀██▄▄
      █  █  ▄██▄  █  █
      █  █  █  █  █  █
      ▀  █  ▀██▀  █  ▀
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	case "dead":
		return splitArt(`
              o
             ▄█▄
         ▄████████▄
         ██████████
         █▓▓x▓▓x▓▓█
         ▀█──────█▀
        ▄██▀▀▀▀▀▀██▄
        █  ▄████▄  █
        █  █    █  █
        █  ▀████▀  █
        ▀██▄▄▄▄▄▄██▀
         █        █
        ▀▀        ▀▀
                      `)
	default: // beer
		return splitArt(`
              o
             ▄█▄
         ▄████████▄
         ██████████
         █▓▓0▓▓0▓▓█
         ▀█╰────╯█▀
       ▄▄██▀▀▀▀▀▀██▄▄
      █  █  ▄██▄  █  █▒▓
      █  █  █  █  █  ▒▒
      ▀  █  ▀██▀  █  ▀▀
       ▀▀██▄▄▄▄▄▄██▀▀
         █        █
        ██        ██
        ▀▀        ▀▀`)
	}
}

func splitArt(s string) []string {
	s = strings.TrimPrefix(s, "\n")
	s = strings.TrimRight(s, "\n")
	return strings.Split(s, "\n")
}

func drawFX(c canvas, x, y int, fx string, i int) {
	switch fx {
	case "smoke":
		puffs := []struct {
			dx, dy int
			ch     string
		}{
			{24, 2, "~"},
			{25, 1, "°"},
			{26, 0, "."},
			{27, 2, "░"},
			{23, 3, "`"},
		}
		for n, p := range puffs {
			dx := p.dx + (i+n)%3
			dy := p.dy - (i+n)%4
			c.blitSoft(x+dx, y+dy, p.ch)
		}
	case "sweat":
		drops := []struct{ dx, dy int }{
			{8, 3}, {22, 3}, {7, 5}, {23, 6}, {10, 7}, {21, 8}, {9, 9},
		}
		ch := []string{"'", "`", "░", "."}
		for n, d := range drops {
			c.blitSoft(x+d.dx, y+((d.dy+i)%10), ch[(n+i)%len(ch)])
		}
	case "fire":
		spots := []struct {
			dx, dy int
			ch     string
		}{
			{6, 0, "^"}, {24, 0, "^"}, {4, 2, "▲"}, {26, 2, "▲"},
			{5, 5, "*"}, {25, 5, "*"}, {8, 1, "░"}, {22, 1, "▒"},
			{10, 0, "*"}, {20, 0, "░"}, {7, 12, "^"}, {23, 12, "*"},
			{3, 8, "▒"}, {27, 8, "░"}, {12, 0, "^"}, {16, 0, "▲"},
		}
		for n, s := range spots {
			if (n+i)%4 == 0 {
				continue
			}
			c.blitSoft(x+s.dx+(i%2), y+s.dy+(n+i)%2, s.ch)
		}
	case "marquee":
		msg := "KILL ALL HUMANS  KILL ALL HUMANS  "
		start := (i * 2) % 18
		c.blit(x+12, y+8, msg[start:start+6])
		if i%2 == 0 {
			c.blitSoft(x+8, y+1, "*")
		}
	}
}
