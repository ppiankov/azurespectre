package azure

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// SQLScanner detects idle Azure SQL databases.
type SQLScanner struct {
	sqlAPI       SQLAPI
	monitor      MonitorAPI
	subscription string
}

// NewSQLScanner creates a SQL database scanner.
func NewSQLScanner(sqlAPI SQLAPI, monitor MonitorAPI, subscription string) *SQLScanner {
	return &SQLScanner{sqlAPI: sqlAPI, monitor: monitor, subscription: subscription}
}

// Type returns the resource type.
func (s *SQLScanner) Type() ResourceType { return ResourceSQLDatabase }

// Scan checks SQL databases for low DTU/vCore utilization.
func (s *SQLScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	databases, err := s.sqlAPI.ListSQLDatabases(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list SQL databases: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(databases)}

	idleCPU := cfg.IdleCPU
	if idleCPU == 0 {
		idleCPU = 5.0
	}
	idleDays := cfg.IdleDays
	if idleDays == 0 {
		idleDays = 7
	}

	if len(databases) == 0 || s.monitor == nil {
		return result, nil
	}

	var uris []string
	uriMap := make(map[string]SQLDatabase)
	for _, db := range databases {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[db.ID] {
			continue
		}
		if shouldExcludeTags(db.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && db.ResourceGroup != cfg.ResourceGroup {
			continue
		}
		uris = append(uris, db.ID)
		uriMap[db.ID] = db
	}

	if len(uris) == 0 {
		return result, nil
	}

	cpuMap, err := s.monitor.FetchMetricMean(ctx, uris, "dtu_consumption_percent", idleDays)
	if err != nil {
		slog.Warn("failed to fetch SQL DTU metrics", "subscription", s.subscription, "error", err)
		return result, nil
	}

	for _, uri := range uris {
		avgDTU, ok := cpuMap[uri]
		if !ok {
			continue
		}
		if avgDTU < idleCPU {
			db := uriMap[uri]
			cost := pricing.MonthlySQLCost(db.SKUTier, db.Capacity, db.Location)
			result.Findings = append(result.Findings, Finding{
				ID:                    FindingIdleSQL,
				Severity:              SeverityHigh,
				ResourceType:          ResourceSQLDatabase,
				ResourceID:            db.ID,
				ResourceName:          fmt.Sprintf("%s/%s", db.ServerName, db.Name),
				Subscription:          s.subscription,
				Region:                db.Location,
				ResourceGroup:         db.ResourceGroup,
				Message:               fmt.Sprintf("DTU utilization %.1f%% over %d days", avgDTU, idleDays),
				EstimatedMonthlyWaste: cost,
				Metadata: map[string]any{
					"server_name":     db.ServerName,
					"sku_name":        db.SKUName,
					"sku_tier":        db.SKUTier,
					"capacity":        db.Capacity,
					"avg_dtu_percent": avgDTU,
				},
			})
		}
	}

	return result, nil
}
