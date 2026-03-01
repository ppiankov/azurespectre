package azure

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// StorageScanner detects unused Azure storage accounts.
type StorageScanner struct {
	storageAPI   StorageAPI
	monitor      MonitorAPI
	subscription string
}

// NewStorageScanner creates a storage account scanner.
func NewStorageScanner(storageAPI StorageAPI, monitor MonitorAPI, subscription string) *StorageScanner {
	return &StorageScanner{storageAPI: storageAPI, monitor: monitor, subscription: subscription}
}

// Type returns the resource type.
func (s *StorageScanner) Type() ResourceType { return ResourceStorage }

// Scan checks storage accounts for zero transactions.
func (s *StorageScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	accounts, err := s.storageAPI.ListStorageAccounts(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list storage accounts: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(accounts)}

	idleDays := cfg.IdleDays
	if idleDays == 0 {
		idleDays = 7
	}

	if len(accounts) == 0 || s.monitor == nil {
		return result, nil
	}

	var uris []string
	uriMap := make(map[string]StorageAccount)
	for _, acct := range accounts {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[acct.ID] {
			continue
		}
		if shouldExcludeTags(acct.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && acct.ResourceGroup != cfg.ResourceGroup {
			continue
		}
		uris = append(uris, acct.ID)
		uriMap[acct.ID] = acct
	}

	if len(uris) == 0 {
		return result, nil
	}

	// Azure Monitor metric for storage transactions is "Transactions"
	txnMap, err := s.monitor.FetchMetricMean(ctx, uris, "Transactions", idleDays)
	if err != nil {
		slog.Warn("failed to fetch storage transaction metrics", "subscription", s.subscription, "error", err)
		return result, nil
	}

	for _, uri := range uris {
		avgTxn, ok := txnMap[uri]
		if !ok || avgTxn > 0 {
			continue
		}
		acct := uriMap[uri]
		cost := pricing.MonthlyStorageCost(acct.SKU, acct.Location)
		result.Findings = append(result.Findings, Finding{
			ID:                    FindingUnusedStorage,
			Severity:              SeverityMedium,
			ResourceType:          ResourceStorage,
			ResourceID:            acct.ID,
			ResourceName:          acct.Name,
			Subscription:          s.subscription,
			Region:                acct.Location,
			ResourceGroup:         acct.ResourceGroup,
			Message:               fmt.Sprintf("Zero transactions over %d days", idleDays),
			EstimatedMonthlyWaste: cost,
			Metadata: map[string]any{
				"sku":  acct.SKU,
				"kind": acct.Kind,
			},
		})
	}

	return result, nil
}
