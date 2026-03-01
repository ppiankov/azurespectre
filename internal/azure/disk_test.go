package azure

import (
	"context"
	"testing"
)

func TestDiskScanner_Type(t *testing.T) {
	s := NewDiskScanner(nil, "sub-1")
	if s.Type() != ResourceDisk {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceDisk)
	}
}

func TestDiskScanner_UnattachedDisk(t *testing.T) {
	compute := &mockComputeAPI{
		disks: []ManagedDisk{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/disks/disk1",
				Name: "disk1", ResourceGroup: "rg1", Location: "eastus",
				SKU: "Premium_LRS", SizeGB: 128, DiskState: "Unattached", ManagedBy: "",
			},
		},
	}
	s := NewDiskScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].ID != FindingUnattachedDisk {
		t.Errorf("finding ID = %q, want %q", result.Findings[0].ID, FindingUnattachedDisk)
	}
}

func TestDiskScanner_AttachedDisk(t *testing.T) {
	compute := &mockComputeAPI{
		disks: []ManagedDisk{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/disks/disk1",
				Name: "disk1", ResourceGroup: "rg1", Location: "eastus",
				SKU: "Premium_LRS", SizeGB: 128, DiskState: "Attached",
				ManagedBy: "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
			},
		},
	}
	s := NewDiskScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestDiskScanner_ExcludeByID(t *testing.T) {
	diskID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/disks/disk1"
	compute := &mockComputeAPI{
		disks: []ManagedDisk{
			{ID: diskID, Name: "disk1", DiskState: "Unattached"},
		},
	}
	s := NewDiskScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		Exclude: ExcludeConfig{ResourceIDs: map[string]bool{diskID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestDiskScanner_Empty(t *testing.T) {
	compute := &mockComputeAPI{}
	s := NewDiskScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
