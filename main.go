// VisualRoom (Vroom) — disk usage TUI
// Author: Sohaib Khan
// Repo:   https://github.com/sohaib1khan/VisualRoom
package main

import (
	"os"

	"vroom/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
