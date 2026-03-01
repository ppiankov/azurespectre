package azure

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ppiankov/azurespectre/internal/pricing"
)

// AppServiceScanner detects idle Azure App Service apps.
type AppServiceScanner struct {
	appServiceAPI AppServiceAPI
	monitor       MonitorAPI
	subscription  string
}

// NewAppServiceScanner creates an App Service scanner.
func NewAppServiceScanner(appServiceAPI AppServiceAPI, monitor MonitorAPI, subscription string) *AppServiceScanner {
	return &AppServiceScanner{appServiceAPI: appServiceAPI, monitor: monitor, subscription: subscription}
}

// Type returns the resource type.
func (s *AppServiceScanner) Type() ResourceType { return ResourceAppService }

// Scan checks App Service apps for zero requests.
func (s *AppServiceScanner) Scan(ctx context.Context, cfg ScanConfig) (*ScanResult, error) {
	apps, err := s.appServiceAPI.ListAppServiceApps(ctx, s.subscription)
	if err != nil {
		return nil, fmt.Errorf("list app service apps: %w", err)
	}

	result := &ScanResult{ResourcesScanned: len(apps)}

	idleDays := cfg.IdleDays
	if idleDays == 0 {
		idleDays = 7
	}

	if len(apps) == 0 || s.monitor == nil {
		return result, nil
	}

	var uris []string
	uriMap := make(map[string]AppServiceApp)
	for _, app := range apps {
		if cfg.Exclude.ResourceIDs != nil && cfg.Exclude.ResourceIDs[app.ID] {
			continue
		}
		if shouldExcludeTags(app.Tags, cfg.Exclude.Tags) {
			continue
		}
		if cfg.ResourceGroup != "" && app.ResourceGroup != cfg.ResourceGroup {
			continue
		}
		// Only check running apps
		if app.State != "Running" {
			continue
		}
		uris = append(uris, app.ID)
		uriMap[app.ID] = app
	}

	if len(uris) == 0 {
		return result, nil
	}

	// Azure Monitor metric for App Service HTTP requests is "Requests"
	reqMap, err := s.monitor.FetchMetricMean(ctx, uris, "Requests", idleDays)
	if err != nil {
		slog.Warn("failed to fetch App Service request metrics", "subscription", s.subscription, "error", err)
		return result, nil
	}

	for _, uri := range uris {
		avgReqs, ok := reqMap[uri]
		if !ok || avgReqs > 0 {
			continue
		}
		app := uriMap[uri]
		cost := pricing.MonthlyAppServiceCost(app.Kind, app.Location)
		result.Findings = append(result.Findings, Finding{
			ID:                    FindingIdleAppService,
			Severity:              SeverityHigh,
			ResourceType:          ResourceAppService,
			ResourceID:            app.ID,
			ResourceName:          app.Name,
			Subscription:          s.subscription,
			Region:                app.Location,
			ResourceGroup:         app.ResourceGroup,
			Message:               fmt.Sprintf("Zero requests over %d days", idleDays),
			EstimatedMonthlyWaste: cost,
			Metadata: map[string]any{
				"kind":  app.Kind,
				"state": app.State,
			},
		})
	}

	return result, nil
}
