package azure

import (
	"context"
	"fmt"
)

// NSGScanner detects unused Azure network security groups.
type NSGScanner struct {
	network      NetworkAPI
	subscription string
}

// NewNSGScanner creates an NSG scanner.
func NewNSGScanner(network NetworkAPI, subscription string) *NSGScanner {
	return &NSGScanner{network: network, subscription: subscription}
}

// Type returns the resource type.
func (s *NSGScanner) Type() ResourceType { return ResourceNSG }

// Scan checks for NSGs not associated with any subnet or NIC.
func (s *NSGScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	nsgs, err := s.network.ListNSGs(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list NSGs: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(nsgs)}

	for _, nsg := range nsgs {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[nsg.ID] {
			continue
		}
		if shouldExcludeTags(nsg.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && nsg.ResourceGroup != cfg.ResourceGroup {
			continue
		}

		if len(nsg.Subnets) == 0 && len(nsg.NICs) == 0 {
			result.Findings = append(result.Findings, Finding{
				ID:                    FindingUnusedNSG,
				Severity:              SeverityLow,
				ResourceType:          ResourceNSG,
				ResourceID:            nsg.ID,
				ResourceName:          nsg.Name,
				Subscription:          s.subscription,
				Region:                nsg.Location,
				ResourceGroup:         nsg.ResourceGroup,
				Message:               "NSG not associated with any subnet or NIC",
				EstimatedMonthlyWaste: 0,
			})
		}
	}

	return result, nil
}
