package azure

import (
	"context"
	"testing"
)

func TestIPScanner_Type(t *testing.T) {
	s := NewIPScanner(nil, "sub-1")
	if s.Type() != ResourcePublicIP {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourcePublicIP)
	}
}

func TestIPScanner_UnusedIP(t *testing.T) {
	network := &mockNetworkAPI{
		ips: []PublicIPAddress{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/publicIPAddresses/ip1",
				Name: "ip1", ResourceGroup: "rg1", Location: "eastus",
				IPAddress: "20.1.2.3", AllocationMethod: "Static",
				AssociatedResource: "",
			},
		},
	}
	s := NewIPScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].ID != FindingUnusedIP {
		t.Errorf("finding ID = %q, want %q", result.Findings[0].ID, FindingUnusedIP)
	}
}

func TestIPScanner_AssociatedIP(t *testing.T) {
	network := &mockNetworkAPI{
		ips: []PublicIPAddress{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/publicIPAddresses/ip1",
				Name: "ip1", ResourceGroup: "rg1", Location: "eastus",
				AssociatedResource: "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/networkInterfaces/nic1/ipConfigurations/ipconfig1",
			},
		},
	}
	s := NewIPScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestIPScanner_ExcludeByID(t *testing.T) {
	ipID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/publicIPAddresses/ip1"
	network := &mockNetworkAPI{
		ips: []PublicIPAddress{
			{ID: ipID, Name: "ip1", AssociatedResource: ""},
		},
	}
	s := NewIPScanner(network, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		Exclude: ExcludeConfig{ResourceIDs: map[string]bool{ipID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
