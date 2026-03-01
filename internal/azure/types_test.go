package azure

import (
	"encoding/json"
	"testing"
)

func TestFindingJSONRoundTrip(t *testing.T) {
	f := Finding{
		ID:                    FindingIdleVM,
		Severity:              SeverityHigh,
		ResourceType:          ResourceVM,
		ResourceID:            "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
		ResourceName:          "vm1",
		Subscription:          "sub-1",
		Region:                "eastus",
		ResourceGroup:         "rg1",
		Message:               "CPU 2.1% over 7 days",
		EstimatedMonthlyWaste: 70.08,
		Metadata:              map[string]any{"vm_size": "Standard_D2s_v5"},
	}

	data, err := json.Marshal(f)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var got Finding
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if got.ID != f.ID {
		t.Errorf("ID = %q, want %q", got.ID, f.ID)
	}
	if got.Severity != f.Severity {
		t.Errorf("Severity = %q, want %q", got.Severity, f.Severity)
	}
	if got.EstimatedMonthlyWaste != f.EstimatedMonthlyWaste {
		t.Errorf("Waste = %f, want %f", got.EstimatedMonthlyWaste, f.EstimatedMonthlyWaste)
	}
}

func TestResourceTypeConstants(t *testing.T) {
	types := []ResourceType{
		ResourceVM, ResourceDisk, ResourcePublicIP, ResourceSnapshot,
		ResourceNSG, ResourceLoadBalancer, ResourceSQLDatabase,
		ResourceAppService, ResourceStorage,
	}
	seen := make(map[ResourceType]bool)
	for _, rt := range types {
		if seen[rt] {
			t.Errorf("duplicate resource type: %q", rt)
		}
		seen[rt] = true
	}
}

func TestFindingIDConstants(t *testing.T) {
	ids := []FindingID{
		FindingIdleVM, FindingStoppedVM, FindingUnattachedDisk, FindingUnusedIP,
		FindingStaleSnapshot, FindingUnusedNSG, FindingIdleLB, FindingIdleSQL,
		FindingIdleAppService, FindingUnusedStorage,
	}
	seen := make(map[FindingID]bool)
	for _, id := range ids {
		if seen[id] {
			t.Errorf("duplicate finding ID: %q", id)
		}
		seen[id] = true
	}
}
