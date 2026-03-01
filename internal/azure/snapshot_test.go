package azure

import (
	"context"
	"testing"
	"time"
)

func TestSnapshotScanner_Type(t *testing.T) {
	s := NewSnapshotScanner(nil, "sub-1")
	if s.Type() != ResourceSnapshot {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceSnapshot)
	}
}

func TestSnapshotScanner_StaleSnapshot(t *testing.T) {
	compute := &mockComputeAPI{
		snapshots: []DiskSnapshot{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/snapshots/snap1",
				Name: "snap1", ResourceGroup: "rg1", Location: "eastus",
				DiskSizeGB: 128, TimeCreated: time.Now().AddDate(0, 0, -120),
			},
		},
	}
	s := NewSnapshotScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{StaleDays: 90})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	if result.Findings[0].ID != FindingStaleSnapshot {
		t.Errorf("finding ID = %q, want %q", result.Findings[0].ID, FindingStaleSnapshot)
	}
}

func TestSnapshotScanner_FreshSnapshot(t *testing.T) {
	compute := &mockComputeAPI{
		snapshots: []DiskSnapshot{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/snapshots/snap1",
				Name: "snap1", ResourceGroup: "rg1", Location: "eastus",
				DiskSizeGB: 128, TimeCreated: time.Now().AddDate(0, 0, -10),
			},
		},
	}
	s := NewSnapshotScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{StaleDays: 90})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestSnapshotScanner_ExcludeByID(t *testing.T) {
	snapID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/snapshots/snap1"
	compute := &mockComputeAPI{
		snapshots: []DiskSnapshot{
			{ID: snapID, Name: "snap1", DiskSizeGB: 128, TimeCreated: time.Now().AddDate(0, 0, -120)},
		},
	}
	s := NewSnapshotScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		StaleDays: 90,
		Exclude:   ExcludeConfig{ResourceIDs: map[string]bool{snapID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestSnapshotScanner_ExcludeByTag(t *testing.T) {
	compute := &mockComputeAPI{
		snapshots: []DiskSnapshot{
			{
				ID:   "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Compute/snapshots/snap1",
				Name: "snap1", DiskSizeGB: 128, TimeCreated: time.Now().AddDate(0, 0, -120),
				Tags: map[string]string{"keep": "true"},
			},
		},
	}
	s := NewSnapshotScanner(compute, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		StaleDays: 90,
		Exclude:   ExcludeConfig{Tags: map[string]string{"keep": "true"}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
