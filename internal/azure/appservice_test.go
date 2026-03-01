package azure

import (
	"context"
	"testing"
)

func TestAppServiceScanner_Type(t *testing.T) {
	s := NewAppServiceScanner(&mockAppServiceAPI{}, nil, "sub-1")
	if s.Type() != ResourceAppService {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceAppService)
	}
}

func TestAppServiceScanner_IdleApp(t *testing.T) {
	appID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Web/sites/app1"
	api := &mockAppServiceAPI{
		apps: []AppServiceApp{
			{
				ID:            appID,
				Name:          "app1",
				ResourceGroup: "rg1",
				Location:      "eastus",
				Kind:          "app",
				State:         "Running",
			},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{appID: 0},
	}

	s := NewAppServiceScanner(api, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{IdleDays: 7})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	f := result.Findings[0]
	if f.ID != FindingIdleAppService {
		t.Errorf("finding ID = %q, want %q", f.ID, FindingIdleAppService)
	}
	if f.EstimatedMonthlyWaste == 0 {
		t.Error("expected non-zero monthly waste")
	}
}

func TestAppServiceScanner_ActiveApp(t *testing.T) {
	appID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Web/sites/app2"
	api := &mockAppServiceAPI{
		apps: []AppServiceApp{
			{ID: appID, Name: "app2", State: "Running", Location: "eastus"},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{appID: 150.0},
	}

	s := NewAppServiceScanner(api, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestAppServiceScanner_StoppedApp(t *testing.T) {
	api := &mockAppServiceAPI{
		apps: []AppServiceApp{
			{
				ID:    "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Web/sites/app3",
				Name:  "app3",
				State: "Stopped",
			},
		},
	}
	monitor := &mockMonitorAPI{}

	s := NewAppServiceScanner(api, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (stopped apps skipped)", len(result.Findings))
	}
}

func TestAppServiceScanner_NilMonitor(t *testing.T) {
	api := &mockAppServiceAPI{
		apps: []AppServiceApp{
			{ID: "app-1", Name: "app1", State: "Running"},
		},
	}

	s := NewAppServiceScanner(api, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (no monitor)", len(result.Findings))
	}
}

func TestAppServiceScanner_ExcludeByTag(t *testing.T) {
	appID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Web/sites/app-prod"
	api := &mockAppServiceAPI{
		apps: []AppServiceApp{
			{
				ID:    appID,
				Name:  "app-prod",
				State: "Running",
				Tags:  map[string]string{"Environment": "production"},
			},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{appID: 0},
	}

	s := NewAppServiceScanner(api, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		Exclude: ExcludeConfig{Tags: map[string]string{"Environment": "production"}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (excluded by tag)", len(result.Findings))
	}
}

func TestAppServiceScanner_Empty(t *testing.T) {
	s := NewAppServiceScanner(&mockAppServiceAPI{}, &mockMonitorAPI{}, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
