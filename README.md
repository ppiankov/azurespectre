# AzureSpectre

[![ANCC](https://img.shields.io/badge/ANCC-compliant-brightgreen)](https://ancc.dev)
[![CI](https://github.com/ppiankov/azurespectre/actions/workflows/ci.yml/badge.svg)](https://github.com/ppiankov/azurespectre/actions/workflows/ci.yml)
[![Go 1.24+](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Azure resource waste auditor. Finds idle VMs, unattached disks, unused IPs, stale snapshots, and idle databases costing money for nothing.

Part of the [Spectre family](https://spectrehub.dev) of infrastructure cleanup tools.

## What it is

AzureSpectre scans your Azure subscription for resources that are running but not doing useful work. It checks Azure Monitor metrics, attachment status, and usage patterns to identify waste across Virtual Machines, Managed Disks, Public IPs, NSGs, Load Balancers, Azure SQL, App Services, Storage Accounts, and Snapshots. Each finding includes an estimated monthly cost so you can prioritize cleanup by dollar impact.

## What it is NOT

- Not a real-time monitoring tool. AzureSpectre is a point-in-time scanner, not a daemon.
- Not a remediation tool. It reports waste and lets you decide what to do.
- Not a security scanner. It checks for idle resources, not misconfigurations or vulnerabilities.
- Not a billing replacement. Cost estimates are approximations based on embedded on-demand pricing, not your actual discounted rates.
- Not a capacity planner. It flags underutilization, not rightsizing recommendations.

## Philosophy

*Principiis obsta* — resist the beginnings.

Azure subscriptions accumulate waste just like AWS accounts. Deallocated VMs with premium disks still attached, public IPs assigned to nothing, storage accounts with zero transactions — these costs compound silently. AzureSpectre surfaces them early so they can be addressed before they surprise anyone on the monthly bill.

The tool presents evidence and lets humans decide. It does not delete resources, does not guess intent, and does not use ML where deterministic checks suffice.

## Installation

```bash
# Homebrew
brew install ppiankov/tap/azurespectre

# From source
git clone https://github.com/ppiankov/azurespectre.git
cd azurespectre && make build
```

## Quick start

```bash
# Scan a subscription
azurespectre scan --subscription xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx

# Scan a specific resource group
azurespectre scan --subscription xxxx --resource-group my-rg

# Custom idle threshold
azurespectre scan --subscription xxxx --idle-days 14

# Exclude production resources by tag
azurespectre scan --subscription xxxx --exclude-tags "Environment=production"

# JSON output for automation
azurespectre scan --subscription xxxx --format json --output report.json

# SARIF output for GitHub Security tab
azurespectre scan --subscription xxxx --format sarif --output results.sarif

# Generate config and RBAC role definition
azurespectre init
```

Requires valid Azure credentials (`az login`, service principal, or managed identity).

## What it audits

| Resource | Finding | Signal | Severity |
|----------|---------|--------|----------|
| Virtual Machines | `IDLE_VM` | CPU < 5% over idle window (Azure Monitor) | high |
| Virtual Machines | `STOPPED_VM` | Deallocated > 30 days | high |
| Managed Disks | `UNATTACHED_DISK` | Not attached to any VM | high |
| Public IPs | `UNUSED_IP` | Not associated with any resource | medium |
| Snapshots | `STALE_SNAPSHOT` | Older than stale threshold | medium |
| Network Security Groups | `UNUSED_NSG` | No associated subnets or NICs | low |
| Load Balancers | `IDLE_LB` | Zero backend pool members | high |
| Azure SQL | `IDLE_SQL` | DTU/vCore utilization < 5% | high |
| App Services | `IDLE_APP_SERVICE` | Zero requests over idle window | high |
| Storage Accounts | `UNUSED_STORAGE` | Zero transactions over idle window | medium |

## Usage

```bash
azurespectre scan [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--subscription` | | Azure subscription ID (required, or env `AZURE_SUBSCRIPTION_ID`) |
| `--resource-group` | | Limit scan to specific resource group |
| `--idle-days` | `7` | Lookback window for utilization metrics |
| `--stale-days` | `90` | Age threshold for snapshots |
| `--stopped-days` | `30` | Days deallocated before flagging VMs |
| `--idle-cpu` | `5.0` | CPU % below which a VM is idle |
| `--min-monthly-cost` | `5.0` | Minimum monthly cost to report ($) |
| `--exclude-tags` | | Exclude resources by tag (`Key=Value` or `Key`, repeatable) |
| `--format` | `text` | Output format: `text`, `json`, `sarif`, `spectrehub` |
| `-o, --output` | stdout | Output file path |
| `--no-progress` | `false` | Disable progress output |
| `--timeout` | `10m` | Scan timeout |

**Other commands:**

| Command | Description |
|---------|-------------|
| `azurespectre init` | Generate `.azurespectre.yaml` config and RBAC role definition |
| `azurespectre version` | Print version, commit, and build date |

## Configuration

AzureSpectre reads `.azurespectre.yaml` from the current directory:

```yaml
subscription: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
resource_group: my-rg
idle_days: 14
stale_days: 180
stopped_days: 30
idle_cpu: 5.0
min_monthly_cost: 5.0
format: json
exclude_tags:
  - "Environment=production"
  - "azurespectre:ignore"
```

Generate a sample config with `azurespectre init`.

## Azure permissions

AzureSpectre requires read-only access. Run `azurespectre init` to generate the minimal RBAC role definition, or assign the **Reader** role on the subscription. Required resource provider permissions:

- `Microsoft.Compute/virtualMachines/read`
- `Microsoft.Compute/disks/read`
- `Microsoft.Compute/snapshots/read`
- `Microsoft.Network/publicIPAddresses/read`
- `Microsoft.Network/networkSecurityGroups/read`
- `Microsoft.Network/loadBalancers/read`
- `Microsoft.Sql/servers/read`, `Microsoft.Sql/servers/databases/read`
- `Microsoft.Web/sites/read`
- `Microsoft.Storage/storageAccounts/read`
- `Microsoft.Insights/metrics/read`

## Authentication

Uses **DefaultAzureCredential** chain:

1. Azure CLI (`az login`)
2. Environment variables (`AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`)
3. Managed Identity
4. Workload Identity

## Output formats

**Text** (default): Human-readable table with severity, resource, region, waste, and message.

**JSON** (`--format json`): `spectre/v1` envelope with findings and summary:
```json
{
  "$schema": "spectre/v1",
  "tool": "azurespectre",
  "version": "0.1.0",
  "findings": [...],
  "summary": {
    "total_resources_scanned": 85,
    "total_findings": 12,
    "total_monthly_waste": 450.00
  }
}
```

**SARIF** (`--format sarif`): SARIF v2.1.0 for GitHub Security tab integration.

**SpectreHub** (`--format spectrehub`): `spectre/v1` envelope for SpectreHub ingestion.

## Architecture

```
azurespectre/
├── cmd/azurespectre/main.go        # Entry point (LDFLAGS version injection)
├── internal/
│   ├── commands/                   # Cobra CLI: scan, init, version
│   ├── azure/                      # Azure SDK clients + 9 resource scanners
│   │   ├── types.go               # Finding, Severity, ResourceType, FindingID
│   │   ├── scanner.go             # SubscriptionScanner orchestrator (10 goroutine limit)
│   │   ├── client.go              # Azure credential setup
│   │   ├── vm.go                  # VMs: idle CPU, stopped instances
│   │   ├── disk.go                # Managed Disks: unattached
│   │   ├── ip.go                  # Public IPs: unassociated
│   │   ├── snapshot.go            # Snapshots: stale
│   │   ├── nsg.go                 # NSGs: no attachments
│   │   ├── lb.go                  # Load Balancers: empty backend pools
│   │   ├── sql.go                 # Azure SQL: idle DTU/vCore
│   │   ├── appservice.go          # App Services: zero requests
│   │   ├── storage.go             # Storage Accounts: zero transactions
│   │   └── exclude.go            # Tag/ID-based filtering
│   ├── pricing/                   # Embedded on-demand pricing
│   ├── analyzer/                  # Filter by min cost, compute summary
│   └── report/                    # Text, JSON, SARIF, SpectreHub reporters
├── Makefile
└── go.mod
```

Key design decisions:

- `cmd/azurespectre/main.go` is minimal — a single `Execute()` call with LDFLAGS version injection.
- All logic lives in `internal/` to prevent external import.
- Each resource type has its own scanner file.
- Azure Monitor metrics fetched via `azquery.Client.QueryResource()` for CPU utilization.
- Bounded concurrency: max 10 goroutines via `errgroup` with semaphore.
- Scanner errors are collected, not fatal — one scanner failure does not abort the whole scan.
- Pricing data is embedded with curated on-demand rates.

## Project status

**Status: Beta** · **v0.1.0** · Pre-1.0

| Milestone | Status |
|-----------|--------|
| 9 resource scanners (VMs, Disks, IPs, Snapshots, NSGs, LBs, SQL, App Services, Storage) | Complete |
| 10 finding types with cost estimates | Complete |
| Bounded concurrent scanning | Complete |
| 4 output formats (text, JSON, SARIF, SpectreHub) | Complete |
| Tag-based exclusion | Complete |
| Config file + init command with RBAC role generation | Complete |
| CI pipeline (test/lint/build) | Complete |
| Homebrew distribution | Complete |
| Test coverage >85% | Complete |
| Multi-subscription support | Planned |
| v1.0 release | Planned |

Pre-1.0: CLI flags and config schemas may change between minor versions. JSON output structure (`spectre/v1`) is stable.

## Known limitations

- **Approximate pricing.** Cost estimates use embedded on-demand rates, not your actual pricing (reserved instances, savings plans, spot). Treat estimates as directional, not exact.
- **Azure Monitor data lag.** Metrics may take up to 15 minutes to appear. Very recently provisioned resources may not have enough data for idle detection.
- **No multi-subscription support.** Scans a single Azure subscription at a time.
- **No rightsizing.** Flags underutilized resources but does not recommend smaller VM sizes or SKUs.
- **Single metric thresholds.** CPU < 5% is a simple heuristic. Some workloads (batch, cron) may appear idle but are not.
- **Storage account granularity.** Transaction count is checked at the account level, not per-container or per-blob.
- **SQL DTU/vCore detection.** Requires Azure Monitor metrics to be available. Serverless SQL databases may show low utilization by design.

## License

MIT License — see [LICENSE](LICENSE).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Issues and pull requests welcome.

Part of the [Spectre family](https://spectrehub.dev).
