package main

import (
	"log/slog"
	"os"

	"github.com/ppiankov/azurespectre/internal/commands"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := commands.Execute(version, commit, date); err != nil {
		slog.Warn("command failed", "error", err)
		os.Exit(1)
	}
}
