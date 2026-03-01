package azure

import (
	"context"
	"fmt"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// LBScanner detects idle Azure load balancers.
type LBScanner struct {
	network      NetworkAPI
	subscription string
}

// NewLBScanner creates a load balancer scanner.
func NewLBScanner(network NetworkAPI, subscription string) *LBScanner {
	return &LBScanner{network: network, subscription: subscription}
}

// Type returns the resource type.
func (s *LBScanner) Type() ResourceType { return ResourceLoadBalancer }

// Scan checks load balancers for zero backend pools or zero rules.
func (s *LBScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	lbs, err := s.network.ListLoadBalancers(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list load balancers: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(lbs)}

	for _, lb := range lbs {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[lb.ID] {
			continue
		}
		if shouldExcludeTags(lb.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && lb.ResourceGroup != cfg.ResourceGroup {
			continue
		}

		if lb.BackendPoolCount == 0 || lb.RuleCount == 0 {
			cost := pricing.MonthlyLBCost(lb.Location)
			msg := "No backend pool members"
			if lb.RuleCount == 0 {
				msg = "No load balancing rules configured"
			}
			result.Findings = append(result.Findings, Finding{
				ID:                    FindingIdleLB,
				Severity:              SeverityHigh,
				ResourceType:          ResourceLoadBalancer,
				ResourceID:            lb.ID,
				ResourceName:          lb.Name,
				Subscription:          s.subscription,
				Region:                lb.Location,
				ResourceGroup:         lb.ResourceGroup,
				Message:               msg,
				EstimatedMonthlyWaste: cost,
				Metadata: map[string]any{
					"sku":                lb.SKU,
					"backend_pool_count": lb.BackendPoolCount,
					"rule_count":         lb.RuleCount,
				},
			})
		}
	}

	return result, nil
}
