package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"vroom/internal/cleanup"
	"vroom/internal/config"
	"vroom/internal/humanize"
	"vroom/internal/notify"
	"vroom/internal/scanner"
	"vroom/internal/theme"
)

var (
	scanDepth  int
	scanNotify bool
	scanJSON   bool
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "One-shot scan and print a size-sorted tree",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot(args, true)
		if err != nil {
			return err
		}
		cfg := config.MustLoad()
		opts := scanner.DefaultOptions()
		if scanDepth > 0 {
			opts.MaxDepth = scanDepth
		}
		w := scanner.New(opts)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		start := time.Now()
		node, err := w.ScanDir(ctx, root)
		if err != nil {
			return err
		}
		_ = w.SaveCache()
		disk, derr := scanner.Usage(root)
		if derr != nil {
			return derr
		}
		_ = cleanup.AppendAudit(fmt.Sprintf("SCAN root=%s duration=%s items=%d", root, time.Since(start).Round(time.Millisecond), node.Count))

		st := theme.StateFromFreePct(disk.FreePct(), cfg.Thresholds.Relaxed, cfg.Thresholds.Sarcastic, cfg.Thresholds.Sweating, cfg.Thresholds.Panicking)
		phrase := theme.Catchphrases(st)[0]

		if !quiet {
			fmt.Fprintf(os.Stderr, "Scanning %s …\n\n", root)
			printChildren(node, 0)
			fmt.Println()
			fmt.Println(humanize.DiskSummary(disk.Avail, disk.Total))
			fmt.Printf("Bender (%s): %s\n", st, phrase)
		}
		if scanNotify {
			notify.AfterScan(cfg, disk)
		}
		return nil
	},
}

func init() {
	scanCmd.Flags().IntVar(&scanDepth, "depth", 1, "listing depth (children of the root)")
	scanCmd.Flags().BoolVar(&scanNotify, "notify", false, "write status file and desktop notification")
	scanCmd.Flags().BoolVar(&scanJSON, "json", false, "reserved")
}

func printChildren(node *scanner.Node, indent int) {
	if node == nil {
		return
	}
	pad := strings.Repeat("  ", indent)
	for _, c := range node.Children {
		kind := " "
		if c.IsDir {
			kind = "/"
		}
		warn := ""
		if c.Err != nil {
			warn = "  (partial)"
		}
		fmt.Printf("%s%8s  %s%s%s\n", pad, humanize.Bytes(c.Size), c.Name, kind, warn)
	}
}
