package azure

import (
	"context"
	"fmt"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// IPScanner detects unused Azure public IP addresses.
type IPScanner struct {
	network      NetworkAPI
	subscription string
}

// NewIPScanner creates a public IP scanner.
func NewIPScanner(network NetworkAPI, subscription string) *IPScanner {
	return &IPScanner{network: network, subscription: subscription}
}

// Type returns the resource type.
func (s *IPScanner) Type() ResourceType { return ResourcePublicIP }

// Scan checks for unused public IPs.
func (s *IPScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	ips, err := s.network.ListPublicIPs(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list public IPs: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(ips)}

	for _, ip := range ips {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[ip.ID] {
			continue
		}
		if shouldExcludeTags(ip.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && ip.ResourceGroup != cfg.ResourceGroup {
			continue
		}

		if ip.AssociatedResource == "" {
			cost := pricing.MonthlyPublicIPCost(ip.Location)
			result.Findings = append(result.Findings, Finding{
				ID:                    FindingUnusedIP,
				Severity:              SeverityMedium,
				ResourceType:          ResourcePublicIP,
				ResourceID:            ip.ID,
				ResourceName:          ip.Name,
				Subscription:          s.subscription,
				Region:                ip.Location,
				ResourceGroup:         ip.ResourceGroup,
				Message:               fmt.Sprintf("Public IP not associated with any resource (%s)", ip.AllocationMethod),
				EstimatedMonthlyWaste: cost,
				Metadata: map[string]any{
					"allocation_method": ip.AllocationMethod,
					"ip_address":        ip.IPAddress,
				},
			})
		}
	}

	return result, nil
}
