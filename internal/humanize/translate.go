package humanize

import (
	"fmt"
	"math"
	"strings"
)

func Bytes(n int64) string {
	if n < 0 {
		n = 0
	}
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n/div >= unit && exp < 5 {
		div *= unit
		exp++
	}
	value := float64(n) / float64(div)
	suffix := []string{"KB", "MB", "GB", "TB", "PB", "EB"}[exp]
	if value >= 10 || math.Abs(value-math.Round(value)) < 0.05 {
		return fmt.Sprintf("%.0f %s", value, suffix)
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}

func Percent(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return 100 * float64(part) / float64(total)
}

func HealthHeadline(freePct float64) string {
	switch {
	case freePct > 50:
		return "cruising fine"
	case freePct > 20:
		return "a little cluttered"
	case freePct > 10:
		return "getting cramped"
	case freePct > 3:
		return "in the danger zone"
	default:
		return "about to seize up"
	}
}

func DiskSummary(free, total uint64) string {
	used := uint64(0)
	if total > free {
		used = total - free
	}
	freePct := Percent(free, total)
	return fmt.Sprintf("You've got %s free out of %s (%s used) — you're %s.",
		Bytes(int64(free)),
		Bytes(int64(total)),
		Bytes(int64(used)),
		HealthHeadline(freePct),
	)
}

func Count(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	word := plural
	if word == "" {
		word = singular + "s"
	}
	return fmt.Sprintf("%d %s", n, word)
}

func JoinEnglish(parts []string) string {
	switch len(parts) {
	case 0:
		return ""
	case 1:
		return parts[0]
	case 2:
		return parts[0] + " and " + parts[1]
	default:
		return strings.Join(parts[:len(parts)-1], ", ") + ", and " + parts[len(parts)-1]
	}
}
