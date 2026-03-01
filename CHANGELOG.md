# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0).

## [0.1.0] - 2026-03-01

### Added

- Initial release
- Azure resource scanning: VMs, managed disks, public IPs, snapshots, NSGs
- 6 finding types: IDLE_VM, STOPPED_VM, UNATTACHED_DISK, UNUSED_IP, STALE_SNAPSHOT, UNUSED_NSG
- Azure Monitor integration for CPU utilization metrics
- Embedded on-demand pricing data for cost estimation
- Analyzer with minimum cost filtering and summary aggregation
- 4 output formats: text (terminal table), JSON (`spectre/v1` envelope), SARIF (v2.1.0), SpectreHub (`spectre/v1`)
- Configuration via `.azurespectre.yaml` with `azurespectre init` generator
- Azure RBAC role definition generator
- Resource exclusion by ID and tags
- GoReleaser config for multi-platform releases (Linux, macOS, Windows; amd64, arm64)
- Homebrew formula via ppiankov/homebrew-tap
