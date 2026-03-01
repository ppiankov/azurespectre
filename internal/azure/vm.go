package azure

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// VMScanner detects idle and stopped Azure VMs.
type VMScanner struct {
	compute      ComputeAPI
	monitor      MonitorAPI
	subscription string
}

// NewVMScanner creates a VM scanner.
func NewVMScanner(compute ComputeAPI, monitor MonitorAPI, subscription string) *VMScanner {
	return &VMScanner{compute: compute, monitor: monitor, subscription: subscription}
}

// Type returns the resource type.
func (s *VMScanner) Type() ResourceType { return ResourceVM }

// Scan checks VMs for idle CPU and long-term deallocation.
func (s *VMScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	vms, err := s.compute.ListVMs(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list VMs: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(vms)}
	now := time.Now().UTC()

	stoppedDays := cfg.StoppedDays
	if stoppedDays == 0 {
		stoppedDays = 30
	}
	idleCPU := cfg.IdleCPU
	if idleCPU == 0 {
		idleCPU = 5.0
	}
	idleDays := cfg.IdleDays
	if idleDays == 0 {
		idleDays = 7
	}

	var runningURIs []string
	runningMap := make(map[string]VirtualMachine)

	for _, vm := range vms {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[vm.ID] {
			continue
		}
		if shouldExcludeTags(vm.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && vm.ResourceGroup != cfg.ResourceGroup {
			continue
		}

		switch vm.PowerState {
		case "PowerState/deallocated":
			created := vm.TimeCreated
			if created.IsZero() {
				continue
			}
			daysStopped := int(now.Sub(created).Hours() / 24)
			if daysStopped >= stoppedDays {
				cost := pricing.MonthlyVMCost(vm.VMSize, vm.Location)
				result.Findings = append(result.Findings, Finding{
					ID:                    FindingStoppedVM,
					Severity:              SeverityHigh,
					ResourceType:          ResourceVM,
					ResourceID:            vm.ID,
					ResourceName:          vm.Name,
					Subscription:          s.subscription,
					Region:                vm.Location,
					ResourceGroup:         vm.ResourceGroup,
					Message:               fmt.Sprintf("Deallocated for %d days", daysStopped),
					EstimatedMonthlyWaste: cost,
					Metadata: map[string]any{
						"vm_size":      vm.VMSize,
						"days_stopped": daysStopped,
						"power_state":  vm.PowerState,
					},
				})
			}
		case "PowerState/running":
			runningURIs = append(runningURIs, vm.ID)
			runningMap[vm.ID] = vm
		}
	}

	if len(runningURIs) > 0 && s.monitor != nil {
		cpuMap, err := s.monitor.FetchMetricMean(ctx, runningURIs, "Percentage CPU", idleDays)
		if err != nil {
			slog.Warn("failed to fetch CPU metrics", "subscription", s.subscription, "error", err)
		} else {
			for _, uri := range runningURIs {
				avgCPU, ok := cpuMap[uri]
				if !ok {
					continue
				}
				if avgCPU < idleCPU {
					vm := runningMap[uri]
					cost := pricing.MonthlyVMCost(vm.VMSize, vm.Location)
					result.Findings = append(result.Findings, Finding{
						ID:                    FindingIdleVM,
						Severity:              SeverityHigh,
						ResourceType:          ResourceVM,
						ResourceID:            vm.ID,
						ResourceName:          vm.Name,
						Subscription:          s.subscription,
						Region:                vm.Location,
						ResourceGroup:         vm.ResourceGroup,
						Message:               fmt.Sprintf("CPU %.1f%% over %d days", avgCPU, idleDays),
						EstimatedMonthlyWaste: cost,
						Metadata: map[string]any{
							"vm_size":         vm.VMSize,
							"avg_cpu_percent": avgCPU,
							"power_state":     "running",
						},
					})
				}
			}
		}
	}

	return result, nil
}
