package theme

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/harmonica"
	"vroom/internal/config"
)

type Animator struct {
	state       State
	frameIdx    int
	frames      []string
	phrases     []string
	phrase      string
	phraseN     int
	tick        time.Duration
	spring      harmonica.Spring
	shownUsage  float64
	usageVel    float64
	targetUsage float64
	cfg         config.BenderConfig
	th          config.Thresholds
}

func NewAnimator(cfg config.BenderConfig, th config.Thresholds) *Animator {
	ms := cfg.TickMS
	if ms < 50 {
		ms = 200
	}
	a := &Animator{
		cfg:    cfg,
		th:     th,
		tick:   time.Duration(ms) * time.Millisecond,
		spring: harmonica.NewSpring(harmonica.FPS(30), 6.0, 0.4),
	}
	a.SetFreePct(100)
	return a
}

func (a *Animator) TickInterval() time.Duration {
	d := a.tick
	switch a.state {
	case StateSarcastic:
		d = a.tick + a.tick/5
	case StateSweating:
		d = a.tick * 3 / 5
	case StatePanicking:
		d = a.tick / 2
	case StateCritical:
		d = a.tick * 2 / 3
	}
	if d < 60*time.Millisecond {
		return 60 * time.Millisecond
	}
	return d
}

func (a *Animator) SetFreePct(freePct float64) {
	st := StateFromFreePct(freePct, a.th.Relaxed, a.th.Sarcastic, a.th.Sweating, a.th.Panicking)
	a.targetUsage = 100 - freePct
	if st != a.state || len(a.frames) == 0 {
		a.state = st
		a.frames = Frames(st)
		a.phrases = Catchphrases(st)
		a.frameIdx = 0
		a.pickPhrase()
	}
}

func (a *Animator) pickPhrase() {
	if len(a.phrases) == 0 {
		a.phrase = ""
		return
	}
	if !a.cfg.Catchphrases {
		a.phrase = a.phrases[0]
		return
	}
	a.phrase = a.phrases[rand.Intn(len(a.phrases))]
}

func (a *Animator) Advance() {
	if len(a.frames) == 0 {
		return
	}
	a.frameIdx = (a.frameIdx + 1) % len(a.frames)
	a.phraseN++
	if a.phraseN%12 == 0 {
		a.pickPhrase()
	}
	a.shownUsage, a.usageVel = a.spring.Update(a.shownUsage, a.usageVel, a.targetUsage)
}

func (a *Animator) Frame() string {
	if len(a.frames) == 0 {
		return ""
	}
	return Paint(a.frames[a.frameIdx], a.state, a.frameIdx)
}

func (a *Animator) Phrase() string {
	return a.phrase
}

func (a *Animator) State() State {
	return a.state
}

func (a *Animator) UsagePct() float64 {
	if a.shownUsage < 0 {
		return 0
	}
	if a.shownUsage > 100 {
		return 100
	}
	return a.shownUsage
}
