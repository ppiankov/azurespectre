# azurespectre

Azure resource waste auditor. Scans Azure subscriptions for idle VMs, unattached disks, unused public IPs, stale snapshots, unused NSGs, and more.

## What It Is

A CLI tool that audits Azure resources for waste and risk. Produces findings with estimated monthly cost savings.

## What It Is NOT

- NOT an Azure cost management tool — shows waste, not full billing
- NOT Azure Advisor — deterministic checks, not ML-based recommendations
- NOT a remediation tool — presents evidence, does not modify resources

## Quick Start

```bash
# Install
brew install ppiankov/tap/azurespectre

# Login to Azure
az login

# Scan
azurespectre scan --subscription <subscription-id>
```

## Usage

```bash
# Scan with defaults
azurespectre scan --subscription <id>

# JSON output
azurespectre scan --subscription <id> --format json

# SpectreHub format
azurespectre scan --subscription <id> --format spectrehub

# Limit to resource group
azurespectre scan --subscription <id> --resource-group mygroup

# Custom thresholds
azurespectre scan --subscription <id> --idle-days 14 --stale-days 60

# Generate config
azurespectre init
```

## Findings

| Resource | Finding | Signal | Severity |
|----------|---------|--------|----------|
| Virtual Machines | `IDLE_VM` | CPU < 5% over idle window (Azure Monitor) | high |
| Virtual Machines | `STOPPED_VM` | Deallocated > 30 days | high |
| Managed Disks | `UNATTACHED_DISK` | Not attached to any VM | high |
| Public IPs | `UNUSED_IP` | Not associated with any resource | medium |
| Snapshots | `STALE_SNAPSHOT` | Older than threshold | medium |
| NSGs | `UNUSED_NSG` | No associated subnets or NICs | low |
| Load Balancers | `IDLE_LB` | Zero backend pool members | high |
| Azure SQL | `IDLE_SQL` | DTU/vCore < 5% | high |
| App Service | `IDLE_APP_SERVICE` | Zero requests | high |
| Storage | `UNUSED_STORAGE` | Zero transactions | medium |

## Output Formats

**Text** (default): Human-readable terminal table.

**JSON** (`--format json`): `spectre/v1` envelope for programmatic consumption.

**SARIF** (`--format sarif`): SARIF v2.1.0 for GitHub Security tab integration.

**SpectreHub** (`--format spectrehub`): `spectre/v1` envelope for SpectreHub ingestion.

## Authentication

AzureSpectre uses `DefaultAzureCredential` which supports:
- Azure CLI (`az login`)
- Environment variables (`AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_CLIENT_SECRET`)
- Managed Identity
- Workload Identity

Requires **Reader** role on the subscription. Run `azurespectre init` to generate the role definition.

## Architecture

```
azurespectre/
├── cmd/azurespectre/main.go     # Entry point
├── internal/
│   ├── azure/                   # Azure SDK wrappers + scanners
│   ├── commands/                # Cobra CLI
│   ├── config/                  # YAML config loader
│   ├── analyzer/                # Finding filter + summary
│   ├── report/                  # Output formatters
│   ├── pricing/                 # Embedded pricing data
│   └── logging/                 # Structured logging
```

## License

MIT
