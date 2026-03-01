package commands

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var forceInit bool

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate sample config and Azure RBAC role definition",
	RunE:  runInit,
}

func init() {
	initCmd.Flags().BoolVar(&forceInit, "force", false, "Overwrite existing files")
}

func runInit(_ *cobra.Command, _ []string) error {
	wrote := false

	if err := writeIfNotExists(".azurespectre.yaml", sampleConfig, forceInit); err != nil {
		return err
	} else if err == nil {
		wrote = true
	}

	if err := writeIfNotExists("azurespectre-role.json", sampleRoleDefinition, forceInit); err != nil {
		return err
	} else if err == nil {
		wrote = true
	}

	if !wrote {
		slog.Info("both files already exist, use --force to overwrite")
	}

	return nil
}

func writeIfNotExists(filename, content string, force bool) error {
	if !force {
		if _, err := os.Stat(filename); err == nil {
			slog.Info("file exists, skipping", "file", filename)
			return nil
		}
	}

	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", filename, err)
	}
	_, _ = fmt.Fprintf(os.Stderr, "wrote %s\n", filename)
	return nil
}

const sampleConfig = `# azurespectre configuration
# subscription: ""              # Azure subscription ID (or use --subscription / AZURE_SUBSCRIPTION_ID)
# resource_group: ""            # Limit scan to resource group
idle_days: 7                    # CPU utilization lookback window
stale_days: 90                  # Snapshot age threshold
stopped_days: 30                # VM deallocated threshold
idle_cpu: 5.0                   # CPU% below which VM is idle
min_monthly_cost: 5.0           # Minimum monthly waste to report
format: text                    # Output format: text, json, sarif, spectrehub
# timeout: 10m
exclude:
  resource_ids: []
  tags: []
  # tags:
  #   - "Environment=production"
  #   - "DoNotDelete"
`

const sampleRoleDefinition = `{
  "Name": "AzureSpectre Reader",
  "Description": "Read-only access for azurespectre resource waste auditing",
  "Actions": [
    "Microsoft.Compute/virtualMachines/read",
    "Microsoft.Compute/virtualMachines/instanceView/read",
    "Microsoft.Compute/disks/read",
    "Microsoft.Compute/snapshots/read",
    "Microsoft.Network/publicIPAddresses/read",
    "Microsoft.Network/networkSecurityGroups/read",
    "Microsoft.Network/loadBalancers/read",
    "Microsoft.Sql/servers/read",
    "Microsoft.Sql/servers/databases/read",
    "Microsoft.Web/sites/read",
    "Microsoft.Web/serverfarms/read",
    "Microsoft.Storage/storageAccounts/read",
    "Microsoft.Insights/metrics/read"
  ],
  "NotActions": [],
  "AssignableScopes": [
    "/subscriptions/{subscription-id}"
  ]
}
`
