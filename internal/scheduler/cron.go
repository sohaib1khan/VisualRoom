package scheduler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"
	"vroom/internal/cleanup"
	"vroom/internal/config"
	"vroom/internal/notify"
	"vroom/internal/scanner"
)

const marker = "# vroom-cron"

type Spec struct {
	Engine   string
	Interval string
	Mode     string
	Root     string
}

func crontabLine(binary, interval, mode, root string) (string, error) {
	sched, err := CronExpr(interval)
	if err != nil {
		return "", err
	}
	args := fmt.Sprintf("%s scan %q --quiet --notify", binary, root)
	if mode == "clean" {
		args = fmt.Sprintf("%s clean --auto --force --quiet %q", binary, root)
	}
	return fmt.Sprintf("%s %s %s", sched, args, marker), nil
}

func CronExpr(interval string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(interval)) {
	case "hourly":
		return "0 * * * *", nil
	case "daily":
		return "0 9 * * *", nil
	case "weekly":
		return "0 9 * * 1", nil
	default:
		if strings.Count(strings.TrimSpace(interval), " ") == 4 {
			return strings.TrimSpace(interval), nil
		}
		return "", fmt.Errorf("interval must be hourly, daily, weekly, or a 5-field cron expr")
	}
}

func binaryPath() string {
	bin, err := os.Executable()
	if err != nil {
		return "vroom"
	}
	if resolved, err := filepath.EvalSymlinks(bin); err == nil {
		return resolved
	}
	return bin
}

func InstallNative(spec Spec) (string, error) {
	line, err := crontabLine(binaryPath(), spec.Interval, spec.Mode, spec.Root)
	if err != nil {
		return "", err
	}
	existing := currentCrontab()
	var kept []string
	for _, l := range strings.Split(existing, "\n") {
		if strings.Contains(l, marker) {
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		kept = append(kept, l)
	}
	kept = append(kept, line)
	body := strings.Join(kept, "\n") + "\n"
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(body)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("crontab install: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	_ = cleanup.AppendAudit(fmt.Sprintf("CRON-INSTALL engine=native interval=%s mode=%s", spec.Interval, spec.Mode))
	return line, nil
}

func RemoveNative() error {
	existing := currentCrontab()
	var kept []string
	removed := 0
	for _, l := range strings.Split(existing, "\n") {
		if strings.Contains(l, marker) {
			removed++
			continue
		}
		if strings.TrimSpace(l) == "" {
			continue
		}
		kept = append(kept, l)
	}
	if removed == 0 {
		return nil
	}
	body := ""
	if len(kept) > 0 {
		body = strings.Join(kept, "\n") + "\n"
	}
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(body)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("crontab remove: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	_ = cleanup.AppendAudit("CRON-REMOVE engine=native")
	return nil
}

func currentCrontab() string {
	out, err := exec.Command("crontab", "-l").CombinedOutput()
	if err != nil {
		return ""
	}
	return string(out)
}

func StatusNative() string {
	existing := currentCrontab()
	var lines []string
	for _, l := range strings.Split(existing, "\n") {
		if strings.Contains(l, marker) {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return "no native vroom crontab entries"
	}
	return strings.Join(lines, "\n")
}

func RunJob(cfg *config.Config, mode, root string) error {
	ctx := context.Background()
	w := scanner.Default()
	disk, err := scanner.Usage(root)
	if err != nil {
		return err
	}
	node, err := w.ScanDir(ctx, root)
	if err != nil {
		return err
	}
	_ = w.SaveCache()
	_ = cleanup.AppendAudit(fmt.Sprintf("SCAN root=%s items=%d size=%d free_pct=%.1f", root, node.Count, node.Size, disk.FreePct()))
	notify.AfterScan(cfg, disk)

	if mode != "clean" {
		return nil
	}
	cutoff := time.Now().Add(-time.Duration(cfg.Cleanup.AutoClean.MinAgeDays) * 24 * time.Hour)
	var targets []string
	for _, child := range node.Children {
		if child == nil || child.IsDir {
			continue
		}
		if cleanup.IsFlagOnly(child.Path, cfg) {
			continue
		}
		if !cleanup.MatchesAutoClean(child.Path, child.IsDir, cfg) {
			continue
		}
		if child.ModTime.After(cutoff) {
			continue
		}
		if d := cleanup.IsProtected(child.Path, cfg); !d.Allowed {
			continue
		}
		targets = append(targets, child.Path)
	}
	if len(targets) == 0 {
		return nil
	}
	_, err = cleanup.Quarantine(cfg, targets, false, false)
	return err
}

func RunDaemon(cfg *config.Config, spec Spec) error {
	expr, err := CronExpr(spec.Interval)
	if err != nil {
		return err
	}
	c := cron.New()
	_, err = c.AddFunc(expr, func() {
		_ = RunJob(cfg, spec.Mode, spec.Root)
	})
	if err != nil {
		return err
	}
	c.Start()
	defer c.Stop()
	_ = cleanup.AppendAudit(fmt.Sprintf("CRON-DAEMON engine=internal interval=%s mode=%s", spec.Interval, spec.Mode))
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
	return nil
}
