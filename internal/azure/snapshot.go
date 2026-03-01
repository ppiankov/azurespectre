package azure

import (
	"context"
	"fmt"
	"time"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// SnapshotScanner detects stale Azure disk snapshots.
type SnapshotScanner struct {
	compute      ComputeAPI
	subscription string
}

// NewSnapshotScanner creates a snapshot scanner.
func NewSnapshotScanner(compute ComputeAPI, subscription string) *SnapshotScanner {
	return &SnapshotScanner{compute: compute, subscription: subscription}
}

// Type returns the resource type.
func (s *SnapshotScanner) Type() ResourceType { return ResourceSnapshot }

// Scan checks for snapshots older than the stale threshold.
func (s *SnapshotScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	snapshots, err := s.compute.ListSnapshots(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list snapshots: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(snapshots)}
	now := time.Now().UTC()

	staleDays := cfg.StaleDays
	if staleDays == 0 {
		staleDays = 90
	}

	for _, snap := range snapshots {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[snap.ID] {
			continue
		}
		if shouldExcludeTags(snap.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && snap.ResourceGroup != cfg.ResourceGroup {
			continue
		}

		if snap.TimeCreated.IsZero() {
			continue
		}

		age := int(now.Sub(snap.TimeCreated).Hours() / 24)
		if age >= staleDays {
			cost := pricing.MonthlySnapshotCost(int(snap.DiskSizeGB), snap.Location)
			result.Findings = append(result.Findings, Finding{
				ID:                    FindingStaleSnapshot,
				Severity:              SeverityMedium,
				ResourceType:          ResourceSnapshot,
				ResourceID:            snap.ID,
				ResourceName:          snap.Name,
				Subscription:          s.subscription,
				Region:                snap.Location,
				ResourceGroup:         snap.ResourceGroup,
				Message:               fmt.Sprintf("Snapshot is %d days old (threshold: %d)", age, staleDays),
				EstimatedMonthlyWaste: cost,
				Metadata: map[string]any{
					"age_days":    age,
					"size_gb":     snap.DiskSizeGB,
					"source_disk": snap.SourceDisk,
				},
			})
		}
	}

	return result, nil
}
