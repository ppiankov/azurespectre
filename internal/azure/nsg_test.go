package azure

import (
	"context"
	"testing"
)

func TestNSGScanner_Type(t *testing.T) {
	s := NewNSGScanner(nil, "sub-1")
	if s.Type() != ResourceNSG {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceNSG)
	}
}

func TestNSGScanner_UnusedNSG(t *testing.T) {
	network := &mockNetworkAPI{
		nsgs: []NetworkSecurityGroup{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/networkSecurityGroups/nsg1",
				Name: "nsg1", ResourceGroup: "rg1", Location: "eastus",
				Subnets: nil, NICs: nil,
			},
		},
	}
	s := NewNSGScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].ID != FindingUnusedNSG {
		t.Errorf("finding ID = %q, want %q", result.Findings[0].ID, FindingUnusedNSG)
	}
	if result.Findings[0].EstimatedMonthlyWaste != 0 {
		t.Errorf("waste = %.2f, want 0", result.Findings[0].EstimatedMonthlyWaste)
	}
}

func TestNSGScanner_UsedNSG(t *testing.T) {
	network := &mockNetworkAPI{
		nsgs: []NetworkSecurityGroup{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/networkSecurityGroups/nsg1",
				Name: "nsg1", ResourceGroup: "rg1", Location: "eastus",
				Subnets: []string{"/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/virtualNetworks/vnet1/subnets/subnet1"},
			},
		},
	}
	s := NewNSGScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestNSGScanner_ExcludeByID(t *testing.T) {
	nsgID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/networkSecurityGroups/nsg1"
	network := &mockNetworkAPI{
		nsgs: []NetworkSecurityGroup{
			{ID: nsgID, Name: "nsg1", Subnets: nil, NICs: nil},
		},
	}
	s := NewNSGScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		Exclude: ExcludeConfig{ResourceIDs: map[string]bool{nsgID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
