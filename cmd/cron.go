package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"vroom/internal/config"
	"vroom/internal/scheduler"
)

var (
	cronInterval string
	cronMode     string
	cronEngine   string
)

var cronCmd = &cobra.Command{
	Use:   "cron",
	Short: "Install, remove, or run scheduled scans",
}

var cronInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install a scheduled scan (native crontab or internal daemon entry)",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		root, err := resolveRoot(nil, true)
		if err != nil {
			return err
		}
		spec := scheduler.Spec{
			Engine:   cronEngine,
			Interval: cronInterval,
			Mode:     cronMode,
			Root:     root,
		}
		if spec.Engine == "" {
			spec.Engine = cfg.Scheduler.Engine
		}
		if spec.Interval == "" {
			spec.Interval = cfg.Scheduler.Interval
		}
		if spec.Mode == "" {
			spec.Mode = cfg.Scheduler.Mode
		}
		switch spec.Engine {
		case "internal":
			fmt.Println("internal engine: run `vroom cron daemon` to keep the scheduler alive")
			fmt.Printf("planned: interval=%s mode=%s root=%s\n", spec.Interval, spec.Mode, spec.Root)
			return nil
		default:
			line, err := scheduler.InstallNative(spec)
			if err != nil {
				return err
			}
			fmt.Println("installed crontab:")
			fmt.Println(" ", line)
			return nil
		}
	},
}

var cronRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove the native crontab entry",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := scheduler.RemoveNative(); err != nil {
			return err
		}
		fmt.Println("removed vroom crontab entries")
		return nil
	},
}

var cronStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show installed native crontab entries",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(scheduler.StatusNative())
		return nil
	},
}

var cronDaemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run the internal robfig/cron scheduler in the foreground",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		root, err := resolveRoot(nil, true)
		if err != nil {
			return err
		}
		spec := scheduler.Spec{
			Engine:   "internal",
			Interval: cronInterval,
			Mode:     cronMode,
			Root:     root,
		}
		if spec.Interval == "" {
			spec.Interval = cfg.Scheduler.Interval
		}
		if spec.Mode == "" {
			spec.Mode = cfg.Scheduler.Mode
		}
		fmt.Printf("vroom daemon interval=%s mode=%s root=%s\n", spec.Interval, spec.Mode, spec.Root)
		return scheduler.RunDaemon(cfg, spec)
	},
}

var cronRunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run one scheduled job immediately",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		root, err := resolveRoot(nil, true)
		if err != nil {
			return err
		}
		mode := cronMode
		if mode == "" {
			mode = cfg.Scheduler.Mode
		}
		return scheduler.RunJob(cfg, mode, root)
	},
}

func init() {
	cronCmd.PersistentFlags().StringVar(&cronInterval, "interval", "daily", "hourly | daily | weekly | 5-field cron")
	cronCmd.PersistentFlags().StringVar(&cronMode, "mode", "notify", "notify | clean")
	cronCmd.PersistentFlags().StringVar(&cronEngine, "engine", "native", "native | internal")
	cronCmd.AddCommand(cronInstallCmd, cronRemoveCmd, cronStatusCmd, cronDaemonCmd, cronRunCmd)
}
