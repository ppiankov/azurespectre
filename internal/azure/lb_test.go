package azure

import (
	"context"
	"testing"
)

func TestLBScanner_Type(t *testing.T) {
	s := NewLBScanner(&mockNetworkAPI{}, "sub-1")
	if s.Type() != ResourceLoadBalancer {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceLoadBalancer)
	}
}

func TestLBScanner_IdleLB_NoBackendPools(t *testing.T) {
	network := &mockNetworkAPI{
		lbs: []LoadBalancer{
			{
				ID:               "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/loadBalancers/lb1",
				Name:             "lb1",
				ResourceGroup:    "rg1",
				Location:         "eastus",
				SKU:              "Standard",
				BackendPoolCount: 0,
				RuleCount:        2,
			},
		},
	}

	s := NewLBScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	f := result.Findings[0]
	if f.ID != FindingIdleLB {
		t.Errorf("finding ID = %q, want %q", f.ID, FindingIdleLB)
	}
	if f.Severity != SeverityHigh {
		t.Errorf("severity = %q, want %q", f.Severity, SeverityHigh)
	}
	if f.EstimatedMonthlyWaste == 0 {
		t.Error("expected non-zero monthly waste")
	}
}

func TestLBScanner_IdleLB_NoRules(t *testing.T) {
	network := &mockNetworkAPI{
		lbs: []LoadBalancer{
			{
				ID:               "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/loadBalancers/lb2",
				Name:             "lb2",
				ResourceGroup:    "rg1",
				Location:         "eastus",
				BackendPoolCount: 1,
				RuleCount:        0,
			},
		},
	}

	s := NewLBScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].Message != "No load balancing rules configured" {
		t.Errorf("message = %q, want 'No load balancing rules configured'", result.Findings[0].Message)
	}
}

func TestLBScanner_ActiveLB(t *testing.T) {
	network := &mockNetworkAPI{
		lbs: []LoadBalancer{
			{
				ID:               "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/loadBalancers/lb3",
				Name:             "lb3",
				ResourceGroup:    "rg1",
				Location:         "eastus",
				BackendPoolCount: 2,
				RuleCount:        3,
			},
		},
	}

	s := NewLBScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestLBScanner_ExcludeByID(t *testing.T) {
	lbID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/loadBalancers/lb-excluded"
	network := &mockNetworkAPI{
		lbs: []LoadBalancer{
			{ID: lbID, Name: "lb-excluded", BackendPoolCount: 0, RuleCount: 0},
		},
	}

	s := NewLBScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		Exclude: ExcludeConfig{ResourceIDs: map[string]bool{lbID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (excluded)", len(result.Findings))
	}
}

func TestLBScanner_Empty(t *testing.T) {
	s := NewLBScanner(&mockNetworkAPI{}, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
