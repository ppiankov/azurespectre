package azure

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSubscriptionScanner_ScanAll(t *testing.T) {
	compute := &mockComputeAPI{
		vms: []VirtualMachine{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/virtualMachines/vm1",
				Name: "vm1", ResourceGroup: "rg1", Location: "eastus",
				VMSize: "Standard_B2s", PowerState: "PowerState/deallocated",
				TimeCreated: time.Now().AddDate(0, 0, -60),
			},
		},
		disks: []ManagedDisk{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/disks/disk1",
				Name: "disk1", ResourceGroup: "rg1", Location: "eastus",
				DiskState: "Unattached",
			},
		},
	}
	network := &mockNetworkAPI{
		ips: []PublicIPAddress{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Network/publicIPAddresses/ip1",
				Name: "ip1", AssociatedResource: "",
			},
		},
	}

	scanner := NewSubscriptionScanner(compute, network, nil, "sub-1", ScanConfig{StoppedDays: 30})
	result, err := scanner.ScanAll(context.Background())
	if err != nil {
		t.Fatalf("ScanAll() error: %v", err)
	}
	if result.ResourcesScanned == 0 {
		t.Error("resources scanned should be > 0")
	}
	if len(result.Findings) < 2 {
		t.Errorf("findings = %d, want >= 2", len(result.Findings))
	}
}

func TestSubscriptionScanner_PartialFailure(t *testing.T) {
	compute := &mockComputeAPI{err: fmt.Errorf("compute API down")}
	network := &mockNetworkAPI{}

	scanner := NewSubscriptionScanner(compute, network, nil, "sub-1", ScanConfig{})
	result, err := scanner.ScanAll(context.Background())
	if err != nil {
		t.Fatalf("ScanAll() error: %v", err)
	}
	if len(result.Errors) == 0 {
		t.Error("expected errors from failed compute API")
	}
}

func TestSubscriptionScanner_Progress(t *testing.T) {
	compute := &mockComputeAPI{}
	network := &mockNetworkAPI{}

	var progress []ScanProgress
	scanner := NewSubscriptionScanner(compute, network, nil, "sub-1", ScanConfig{})
	scanner.SetProgressFn(func(p ScanProgress) {
		progress = append(progress, p)
	})

	_, err := scanner.ScanAll(context.Background())
	if err != nil {
		t.Fatalf("ScanAll() error: %v", err)
	}
	if len(progress) != 6 {
		t.Errorf("progress callbacks = %d, want 6", len(progress))
	}
}

func TestSubscriptionScanner_Empty(t *testing.T) {
	compute := &mockComputeAPI{}
	network := &mockNetworkAPI{}

	scanner := NewSubscriptionScanner(compute, network, nil, "sub-1", ScanConfig{})
	result, err := scanner.ScanAll(context.Background())
	if err != nil {
		t.Fatalf("ScanAll() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
