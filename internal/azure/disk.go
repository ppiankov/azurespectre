package azure

import (
	"context"
	"fmt"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// DiskScanner detects unattached Azure managed disks.
type DiskScanner struct {
	compute      ComputeAPI
	subscription string
}

// NewDiskScanner creates a disk scanner.
func NewDiskScanner(compute ComputeAPI, subscription string) *DiskScanner {
	return &DiskScanner{compute: compute, subscription: subscription}
}

// Type returns the resource type.
func (s *DiskScanner) Type() ResourceType { return ResourceDisk }

// Scan checks for unattached managed disks.
func (s *DiskScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	disks, err := s.compute.ListDisks(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list disks: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(disks)}

	for _, disk := range disks {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[disk.ID] {
			continue
		}
		if shouldExcludeTags(disk.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && disk.ResourceGroup != cfg.ResourceGroup {
			continue
		}

		if disk.DiskState == "Unattached" && disk.ManagedBy == "" {
			cost := pricing.MonthlyDiskCost(disk.SKU, int(disk.SizeGB), disk.Location)
			result.Findings = append(result.Findings, Finding{
				ID:                    FindingUnattachedDisk,
				Severity:              SeverityHigh,
				ResourceType:          ResourceDisk,
				ResourceID:            disk.ID,
				ResourceName:          disk.Name,
				Subscription:          s.subscription,
				Region:                disk.Location,
				ResourceGroup:         disk.ResourceGroup,
				Message:               fmt.Sprintf("Unattached %s disk (%d GB)", disk.SKU, disk.SizeGB),
				EstimatedMonthlyWaste: cost,
				Metadata: map[string]any{
					"sku":     disk.SKU,
					"size_gb": disk.SizeGB,
				},
			})
		}
	}

	return result, nil
}
