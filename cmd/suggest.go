package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"vroom/internal/config"
	"vroom/internal/scanner"
	"vroom/internal/suggest"
)

var suggestCmd = &cobra.Command{
	Use:   "suggest [path]",
	Short: "Print rule-based cleanup suggestions in Bender's voice",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root, err := resolveRoot(args, true)
		if err != nil {
			return err
		}
		cfg := config.MustLoad()
		w := scanner.Default()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		node, err := w.ScanDir(ctx, root)
		if err != nil {
			return err
		}
		disk, err := scanner.Usage(root)
		if err != nil {
			return err
		}
		for _, s := range suggest.Analyze(cfg, node, disk) {
			fmt.Printf("• %s\n  %s\n", s.Title, s.Voice)
			if s.Detail != "" {
				fmt.Printf("  %s\n", s.Detail)
			}
			fmt.Println()
		}
		return nil
	},
}
