# azurespectre

[![CI](https://github.com/ppiankov/azurespectre/actions/workflows/ci.yml/badge.svg)](https://github.com/ppiankov/azurespectre/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ppiankov/azurespectre)](https://goreportcard.com/report/github.com/ppiankov/azurespectre)

**azurespectre** — Azure resource waste auditor with cost estimates. Part of [SpectreHub](https://github.com/ppiankov/spectrehub).

## What it is

- Scans VMs, managed disks, public IPs, NSGs, load balancers, Azure SQL, App Services, storage accounts, and snapshots
- Detects idle, unattached, and oversized resources using Azure Monitor metrics
- Estimates monthly waste in USD per finding
- Supports configurable thresholds and exclusions
- Outputs text, JSON, SARIF, and SpectreHub formats

## What it is NOT

- Not a real-time monitor — point-in-time scanner
- Not a remediation tool — reports only, never modifies resources
- Not a security scanner — checks utilization, not vulnerabilities
- Not a billing replacement — uses embedded on-demand pricing

## Quick start

### Homebrew

```sh
brew tap ppiankov/tap
brew install azurespectre
```

### From source

```sh
git clone https://github.com/ppiankov/azurespectre.git
cd azurespectre
make build
```

### Usage

```sh
azurespectre scan --subscription <id> --format json
```

## CLI commands

| Command | Description |
|---------|-------------|
| `azurespectre scan` | Scan Azure subscription for idle and wasteful resources |
| `azurespectre init` | Generate config file and role assignment |
| `azurespectre version` | Print version |

## SpectreHub integration

azurespectre feeds Azure resource waste findings into [SpectreHub](https://github.com/ppiankov/spectrehub) for unified visibility across your infrastructure.

```sh
spectrehub collect --tool azurespectre
```

## Safety

azurespectre operates in **read-only mode**. It inspects and reports — never modifies, deletes, or alters your resources.

## License

MIT — see [LICENSE](LICENSE).

---

Built by [Obsta Labs](https://github.com/ppiankov)
