package suggest

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"vroom/internal/cleanup"
	"vroom/internal/config"
	"vroom/internal/humanize"
	"vroom/internal/scanner"
)

type Suggestion struct {
	Title  string
	Detail string
	Voice  string
	Path   string
	Bytes  int64
	Kind   string
}

func Analyze(cfg *config.Config, root *scanner.Node, disk scanner.DiskUsage) []Suggestion {
	var out []Suggestion
	if root == nil {
		return out
	}
	now := time.Now()

	for _, child := range root.Children {
		if child == nil {
			continue
		}
		name := strings.ToLower(child.Name)
		if name == "downloads" || name == "download" {
			age := now.Sub(child.ModTime)
			if child.Size > 2*1024*1024*1024 && age > 90*24*time.Hour {
				out = append(out, Suggestion{
					Kind:   "stale",
					Path:   child.Path,
					Bytes:  child.Size,
					Title:  fmt.Sprintf("Downloads is %s and idle", humanize.Bytes(child.Size)),
					Detail: fmt.Sprintf("%s hasn't changed in %d days.", child.Path, int(age.Hours()/24)),
					Voice:  fmt.Sprintf("Your Downloads folder is %s and untouched for 90+ days. Romantic.", humanize.Bytes(child.Size)),
				})
			}
		}
	}

	var nmCount int
	var nmBytes int64
	collectNamed(root, cfg.Cleanup.AutoClean.FlagOnly, &nmCount, &nmBytes)
	if nmCount > 0 && nmBytes > 100*1024*1024 {
		out = append(out, Suggestion{
			Kind:   "flag",
			Bytes:  nmBytes,
			Title:  fmt.Sprintf("Flag-only caches totaling %s", humanize.Bytes(nmBytes)),
			Detail: fmt.Sprintf("Found %s of node_modules/.cache/__pycache__ style dirs. Flagged, never auto-deleted.", humanize.Count(nmCount, "folder", "folders")),
			Voice:  fmt.Sprintf("You have cache-ish folders totaling %s across %d spots. I'm not touching them without a meatbag confirmation.", humanize.Bytes(nmBytes), nmCount),
		})
	}

	if batches, err := cleanup.ListBatches(cfg); err == nil {
		var pending int64
		soon := 0
		deadline := time.Now().Add(48 * time.Hour)
		for _, b := range batches {
			for _, it := range b.Items {
				pending += it.Size
			}
			if !b.ExpiresAt.IsZero() && b.ExpiresAt.Before(deadline) {
				soon++
			}
		}
		if pending > 0 {
			out = append(out, Suggestion{
				Kind:   "quarantine",
				Bytes:  pending,
				Title:  fmt.Sprintf("Quarantine holds %s", humanize.Bytes(pending)),
				Detail: fmt.Sprintf("%d batch(es); %d expire within 2 days.", len(batches), soon),
				Voice:  fmt.Sprintf("Your quarantine folder has %s pending permanent deletion. Restore or let it burn.", humanize.Bytes(pending)),
			})
		}
	}

	if disk.FreePct() < cfg.Thresholds.Sarcastic {
		out = append(out, Suggestion{
			Kind:   "health",
			Title:  "Free space is looking grim",
			Detail: humanize.DiskSummary(disk.Avail, disk.Total),
			Voice:  "I'm sweating oil. Pick something on the left and hit d before I rust.",
		})
	}

	if len(out) == 0 {
		out = append(out, Suggestion{
			Kind:   "ok",
			Title:  "Nothing spicy to report",
			Detail: "No stale Downloads, no huge cache piles, empty quarantine.",
			Voice:  "I'm 40% disk, 40% beer, 20% sass — and you're fine. For now.",
		})
	}
	return out
}

func collectNamed(n *scanner.Node, names []string, count *int, bytes *int64) {
	if n == nil {
		return
	}
	base := filepath.Base(n.Path)
	for _, name := range names {
		if base == name {
			*count++
			*bytes += n.Size
			return
		}
	}
	for _, c := range n.Children {
		collectNamed(c, names, count, bytes)
	}
}

func ScanFlagDirs(root string, names []string) (count int, bytes int64) {
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || !d.IsDir() {
			return nil
		}
		base := d.Name()
		for _, name := range names {
			if base == name {
				count++
				info, ierr := d.Info()
				if ierr == nil {
					bytes += info.Size()
				}
				return filepath.SkipDir
			}
		}
		return nil
	})
	return count, bytes
}
