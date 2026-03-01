package azure

import (
	"context"
	"testing"
	"time"
)

func TestVMScanner_Type(t *testing.T) {
	s := NewVMScanner(nil, nil, "sub-1")
	if s.Type() != ResourceVM {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceVM)
	}
}

func TestVMScanner_StoppedVM(t *testing.T) {
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
				Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_B2s", PowerState: "PowerState/deallocated",
				TimeCreated: time.Now().AddDate(0, 0, -60),
			},
		},
	}
	s := NewVMScanner(compute, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{StoppedDays: 30})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].ID != FindingStoppedVM {
		t.Errorf("finding ID = %q, want %q", result.Findings[0].ID, FindingStoppedVM)
	}
}

func TestVMScanner_RecentlyDeallocated(t *testing.T) {
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
				Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_B2s", PowerState: "PowerState/deallocated",
				TimeCreated: time.Now().AddDate(0, 0, -5),
			},
		},
	}
	s := NewVMScanner(compute, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{StoppedDays: 30})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestVMScanner_IdleCPU(t *testing.T) {
	vmID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID: vmID, Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_D2s_v5", PowerState: "PowerState/running",
			},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{vmID: 2.1},
	}
	s := NewVMScanner(compute, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{IdleCPU: 5.0, IdleDays: 7})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].ID != FindingIdleVM {
		t.Errorf("finding ID = %q, want %q", result.Findings[0].ID, FindingIdleVM)
	}
}

func TestVMScanner_BusyCPU(t *testing.T) {
	vmID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID: vmID, Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_D2s_v5", PowerState: "PowerState/running",
			},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{vmID: 45.0},
	}
	s := NewVMScanner(compute, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{IdleCPU: 5.0, IdleDays: 7})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestVMScanner_ExcludeByID(t *testing.T) {
	vmID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1"
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID: vmID, Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_B2s", PowerState: "PowerState/deallocated",
				TimeCreated: time.Now().AddDate(0, 0, -60),
			},
		},
	}
	s := NewVMScanner(compute, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		StoppedDays: 30,
		Exclude:     ExcludeConfig{ResourceIDs: map[string]bool{vmID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (excluded)", len(result.Findings))
	}
}

func TestVMScanner_ExcludeByTag(t *testing.T) {
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
				Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_B2s", PowerState: "PowerState/deallocated",
				TimeCreated: time.Now().AddDate(0, 0, -60),
				Tags:        map[string]string{"Environment": "production"},
			},
		},
	}
	s := NewVMScanner(compute, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		StoppedDays: 30,
		Exclude:     ExcludeConfig{Tags: map[string]string{"Environment": "production"}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (excluded by tag)", len(result.Findings))
	}
}

func TestVMScanner_Empty(t *testing.T) {
	compute := &mockComputeAPI{}
	s := NewVMScanner(compute, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestVMScanner_NilMonitor(t *testing.T) {
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
				Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_D2s_v5", PowerState: "PowerState/running",
			},
		},
	}
	s := NewVMScanner(compute, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{IdleCPU: 5.0})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (no monitor)", len(result.Findings))
	}
}

func TestVMScanner_ResourceGroupFilter(t *testing.T) {
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
				Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_B2s", PowerState: "PowerState/deallocated",
				TimeCreated: time.Now().AddDate(0, 0, -60),
			},
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg2/providers/Microsoft.Compute/virtualMachines/vm2",
				Name: "vm2", ResourceGroup: "rg2", Location: "eastus",
				VMSize: "Standard_B2s", PowerState: "PowerState/deallocated",
				TimeCreated: time.Now().AddDate(0, 0, -60),
			},
		},
	}
	s := NewVMScanner(compute, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		StoppedDays:   30,
		ResourceGroup: "rg1",
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1 (filtered to rg1)", len(result.Findings))
	}
	if result.Findings[0].ResourceName != "vm1" {
		t.Errorf("resource = %q, want vm1", result.Findings[0].ResourceName)
	}
}
