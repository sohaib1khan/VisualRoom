package notify

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gen2brain/beeep"
	"vroom/internal/config"
	"vroom/internal/humanize"
	"vroom/internal/scanner"
	"vroom/internal/theme"
)

func Desktop(title, message string) error {
	return beeep.Notify(title, message, "")
}

func WriteStatus(cfg *config.Config, disk scanner.DiskUsage, headline string) error {
	path := cfg.Notifications.StatusFile
	if path == "" {
		path = filepath.Join(config.DataDir(), "status.txt")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	body := fmt.Sprintf("vroom %s\n%s\n%s\n",
		time.Now().Format(time.RFC3339),
		humanize.DiskSummary(disk.Avail, disk.Total),
		headline,
	)
	return os.WriteFile(path, []byte(body), 0o600)
}

func AfterScan(cfg *config.Config, disk scanner.DiskUsage) {
	st := theme.StateFromFreePct(disk.FreePct(), cfg.Thresholds.Relaxed, cfg.Thresholds.Sarcastic, cfg.Thresholds.Sweating, cfg.Thresholds.Panicking)
	phrases := theme.Catchphrases(st)
	headline := phrases[0]
	_ = WriteStatus(cfg, disk, headline)
	if cfg.Notifications.Desktop {
		title := "Vroom: " + st.String()
		_ = Desktop(title, humanize.DiskSummary(disk.Avail, disk.Total)+"\n"+headline)
	}
}
