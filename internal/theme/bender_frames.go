package theme

type State int

const (
	StateRelaxed State = iota
	StateSarcastic
	StateSweating
	StatePanicking
	StateCritical
)

func (s State) String() string {
	switch s {
	case StateRelaxed:
		return "relaxed"
	case StateSarcastic:
		return "sarcastic"
	case StateSweating:
		return "sweating"
	case StatePanicking:
		return "panicking"
	case StateCritical:
		return "critical"
	default:
		return "unknown"
	}
}

func AllStates() []State {
	return []State{StateRelaxed, StateSarcastic, StateSweating, StatePanicking, StateCritical}
}

func StateFromFreePct(freePct float64, relaxed, sarcastic, sweating, panicking float64) State {
	switch {
	case freePct > relaxed:
		return StateRelaxed
	case freePct > sarcastic:
		return StateSarcastic
	case freePct > sweating:
		return StateSweating
	case freePct > panicking:
		return StatePanicking
	default:
		return StateCritical
	}
}

func Catchphrases(state State) []string {
	switch state {
	case StateRelaxed:
		return []string{
			"Bite my shiny metal disk!",
			"I'm 40% disk, 40% beer, 20% sass.",
			"Cruising. Don't touch my stuff.",
			"This is the best disk of my life.",
		}
	case StateSarcastic:
		return []string{
			"Eh. I've seen worse. I've also seen better.",
			"Neat. You're collecting junk like it's a hobby.",
			"Sure, keep those caches. They're not using anything. Oh wait.",
			"I'm not saying you're messy. I'm implying it.",
		}
	case StateSweating:
		return []string{
			"Hey meatbag, that's not enough room for my beer!",
			"Sweating oil over here. Clean something.",
			"My circuits are pacing. Yours should be too.",
			"We're running out of shiny. That's a problem.",
		}
	case StatePanicking:
		return []string{
			"I need my hands! Do something before I rust shut!",
			"THIS IS NOT A DRILL. Well, maybe a disk drill.",
			"I'm on fire and it's YOUR clutter!",
			"Quarantine the junk. Now. Please. I'm begging. I'm not.",
		}
	default:
		return []string{
			"KILL ALL HUMANS — wait, no, kill all leftover ISOs.",
			"Powered down. Wake me when you delete something.",
			"Blocks new writes warning. That's not a joke.",
			"I'm already dead. Your disk is next.",
		}
	}
}
