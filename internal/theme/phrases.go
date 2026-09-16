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
			"Plenty of room. You're cruising.",
			"Disk looks healthy. Keep it that way.",
			"Free space is comfortable. No rush.",
			"All good. This is the easy part.",
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
			"Free space is getting tight. Time to clean.",
			"This disk is filling up. Pick something to quarantine.",
			"Running low. Caches and downloads first.",
			"We're running out of room. That's a problem.",
		}
	case StatePanicking:
		return []string{
			"Disk is nearly full. Clean something now.",
			"This is not a drill. Quarantine the junk.",
			"Free space is critical. Start with the biggest folders.",
			"Quarantine the clutter before writes start failing.",
		}
	default:
		return []string{
			"Disk is effectively full. Delete leftover ISOs.",
			"Powered down mood. Wake me when you free some space.",
			"New writes may fail. Clear space immediately.",
			"Critical: almost no free space left.",
		}
	}
}
