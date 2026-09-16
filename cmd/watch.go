package cmd

import (
	"github.com/spf13/cobra"
	"vroom/internal/tui"
)

var watchCmd = &cobra.Command{
	Use:   "watch [path]",
	Short: "Interactive TUI with animated disk-health mascot (default)",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runWatch,
}

func runWatch(cmd *cobra.Command, args []string) error {
	root, err := resolveRoot(args, true)
	if err != nil {
		return err
	}
	return tui.Start(root)
}
