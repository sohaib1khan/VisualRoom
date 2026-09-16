package cmd

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"vroom/internal/cleanup"
	"vroom/internal/config"
	"vroom/internal/humanize"
)

var restoreAllowElevated bool

var restoreCmd = &cobra.Command{
	Use:   "restore <id>",
	Short: "Restore a quarantine batch back to original paths",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		if err := cleanup.Restore(cfg, args[0], restoreAllowElevated); err != nil {
			return err
		}
		fmt.Println("restored batch", args[0])
		return nil
	},
}

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Permanently delete expired quarantine batches",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		force, _ := cmd.Flags().GetBool("all")
		n, err := cleanup.PurgeExpired(cfg, force)
		if err != nil {
			return err
		}
		fmt.Printf("purged %d batch(es)\n", n)
		return nil
	},
}

var quarantineCmd = &cobra.Command{
	Use:   "quarantine",
	Short: "List quarantine batches",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.MustLoad()
		batches, err := cleanup.ListBatches(cfg)
		if err != nil {
			return err
		}
		if len(batches) == 0 {
			fmt.Println("quarantine is empty")
			return nil
		}
		for _, b := range batches {
			var bytes int64
			for _, it := range b.Items {
				bytes += it.Size
			}
			fmt.Printf("%s  %s  %s  expires %s  %d item(s)\n",
				b.ID,
				humanize.Bytes(bytes),
				b.CreatedAt.Format(time.RFC3339),
				b.ExpiresAt.Format(time.RFC3339),
				len(b.Items),
			)
			for _, it := range b.Items {
				fmt.Printf("    %s\n", it.Original)
			}
		}
		return nil
	},
}

func init() {
	restoreCmd.Flags().BoolVar(&restoreAllowElevated, "allow-elevated", false, "allow running restore as root")
	purgeCmd.Flags().Bool("all", false, "purge all batches, not only expired")
}
