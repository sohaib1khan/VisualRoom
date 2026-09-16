package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"vroom/internal/about"
	"vroom/internal/config"
)

var (
	rootFlag string
	quiet    bool
)

var rootCmd = &cobra.Command{
	Use:   "vroom",
	Short: "Animated CLI disk usage monitor",
	Long: `Vroom scans disk usage, explains it in plain language, and animates
a live mood-indicator of disk health.

Author: ` + about.Author + `
Repo:   ` + about.Repo + `

Default command is interactive watch mode. Cleanup is dry-run unless --force
is passed; files are quarantined, never silently deleted.`,
	Version:       about.Version,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.MaximumNArgs(1),
	RunE:          runWatch,
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "vroom:", err)
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&rootFlag, "root", "", "scan root (default: current directory, or $HOME for watch)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "less output")
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(cleanCmd)
	rootCmd.AddCommand(restoreCmd)
	rootCmd.AddCommand(purgeCmd)
	rootCmd.AddCommand(cronCmd)
	rootCmd.AddCommand(suggestCmd)
	rootCmd.AddCommand(quarantineCmd)
	rootCmd.AddCommand(previewCmd)
}

func resolveRoot(args []string, fallbackHome bool) (string, error) {
	if rootFlag != "" {
		return absPath(config.ExpandPath(rootFlag))
	}
	if len(args) > 0 {
		return absPath(config.ExpandPath(args[0]))
	}
	if fallbackHome {
		home := config.HomeDir()
		if home == "" {
			return os.Getwd()
		}
		return home, nil
	}
	return os.Getwd()
}

func absPath(p string) (string, error) {
	if p == "" {
		return os.Getwd()
	}
	return filepath.Abs(p)
}
