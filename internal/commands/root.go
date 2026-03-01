package commands

import (
	"log/slog"

	"github.com/ppiankov/azurespectre/internal/config"
	"github.com/ppiankov/azurespectre/internal/logging"
	"github.com/spf13/cobra"
)

var (
	verbose               bool
	version, commit, date string
	cfg                   config.Config
)

var rootCmd = &cobra.Command{
	Use:   "azurespectre",
	Short: "azurespectre — Azure resource waste auditor",
	Long:  `Scans Azure subscriptions for idle VMs, unattached disks, unused public IPs, stale snapshots, unused NSGs, and more.`,
	PersistentPreRun: func(_ *cobra.Command, _ []string) {
		logging.Init(verbose)
		loaded, err := config.Load(".")
		if err != nil {
			slog.Warn("failed to load config file", "error", err)
		} else {
			cfg = loaded
		}
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	rootCmd.AddCommand(scanCmd, initCmd, versionCmd)
}

// Execute runs the root command.
func Execute(v, c, d string) error {
	version, commit, date = v, c, d
	return rootCmd.Execute()
}
