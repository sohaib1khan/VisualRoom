package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"vroom/internal/cleanup"
	"vroom/internal/config"
	"vroom/internal/humanize"
	"vroom/internal/scanner"
	"vroom/internal/scheduler"
)

var (
	cleanForce         bool
	cleanAllowElevated bool
	cleanAuto          bool
)

var cleanCmd = &cobra.Command{
	Use:   "clean [paths...]",
	Short: "Quarantine paths (dry-run unless --force)",
	Args:  cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		dry := !cleanForce
		if len(args) == 0 && !cleanAuto {
			return fmt.Errorf("pass paths to quarantine, or --auto to apply whitelist rules")
		}

		root, err := resolveRoot(nil, true)
		if err != nil {
			return err
		}

		if scanner.InContainer() && cleanForce {
			target := root
			if len(args) > 0 {
				target = args[0]
			}
			if err := cleanup.DockerCleanBlocked(target, cleanForce); err != nil {
				return err
			}
		}

		var paths []string
		if cleanAuto {
			if err := scheduler.RunJob(cfg, "clean", root); err != nil {
				return err
			}
			if !quiet {
				fmt.Println("auto-clean finished (whitelist patterns only; see audit log)")
			}
			return nil
		}
		for _, a := range args {
			p, err := absPath(config.ExpandPath(a))
			if err != nil {
				return err
			}
			paths = append(paths, p)
		}

		res, err := cleanup.Quarantine(cfg, paths, dry, cleanAllowElevated)
		if err != nil {
			return err
		}
		if dry && !quiet {
			fmt.Println("Dry-run (pass --force to actually quarantine):")
		}
		for _, it := range res.Batch.Items {
			fmt.Printf("  %s  %s\n", humanize.Bytes(it.Size), it.Original)
		}
		for _, s := range res.Skipped {
			fmt.Fprintf(os.Stderr, "  skip %s: %s\n", s.Path, s.Reason)
		}
		if !quiet {
			verb := "would quarantine"
			if !dry {
				verb = "quarantined"
			}
			fmt.Printf("%s %d item(s), %s\n", verb, res.Moved, humanize.Bytes(res.Bytes))
			if !dry {
				fmt.Printf("batch id: %s  expires %s\n", res.Batch.ID, res.Batch.ExpiresAt.Format(time.RFC3339))
				fmt.Println("restore with: vroom restore " + res.Batch.ID)
			}
		}
		return nil
	},
}

func init() {
	cleanCmd.Flags().BoolVar(&cleanForce, "force", false, "actually move files (otherwise dry-run)")
	cleanCmd.Flags().BoolVar(&cleanAllowElevated, "allow-elevated", false, "allow running cleanup as root")
	cleanCmd.Flags().BoolVar(&cleanAuto, "auto", false, "whitelist-only auto-clean (cron mode)")
}
