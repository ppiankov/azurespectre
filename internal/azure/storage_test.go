package azure

import (
	"context"
	"testing"
)

func TestStorageScanner_Type(t *testing.T) {
	s := NewStorageScanner(&mockStorageAPI{}, nil, "sub-1")
	if s.Type() != ResourceStorage {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceStorage)
	}
}

func TestStorageScanner_UnusedStorage(t *testing.T) {
	acctID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/sa1"
	storageAPI := &mockStorageAPI{
		accounts: []StorageAccount{
			{
				ID:            acctID,
				Name:          "sa1",
				ResourceGroup: "rg1",
				Location:      "eastus",
				SKU:           "Standard_LRS",
				Kind:          "StorageV2",
			},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{acctID: 0},
	}

	s := NewStorageScanner(storageAPI, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{IdleDays: 7})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	f := result.Findings[0]
	if f.ID != FindingUnusedStorage {
		t.Errorf("finding ID = %q, want %q", f.ID, FindingUnusedStorage)
	}
	if f.Severity != SeverityMedium {
		t.Errorf("severity = %q, want %q", f.Severity, SeverityMedium)
	}
}

func TestStorageScanner_ActiveStorage(t *testing.T) {
	acctID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/sa2"
	storageAPI := &mockStorageAPI{
		accounts: []StorageAccount{
			{ID: acctID, Name: "sa2", Location: "eastus"},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{acctID: 500.0},
	}

	s := NewStorageScanner(storageAPI, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestStorageScanner_NilMonitor(t *testing.T) {
	storageAPI := &mockStorageAPI{
		accounts: []StorageAccount{
			{ID: "sa-1", Name: "sa1"},
		},
	}

	s := NewStorageScanner(storageAPI, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (no monitor)", len(result.Findings))
	}
}

func TestStorageScanner_ExcludeByID(t *testing.T) {
	acctID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Storage/storageAccounts/sa-excluded"
	storageAPI := &mockStorageAPI{
		accounts: []StorageAccount{
			{ID: acctID, Name: "sa-excluded"},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{acctID: 0},
	}

	s := NewStorageScanner(storageAPI, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		Exclude: ExcludeConfig{ResourceIDs: map[string]bool{acctID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (excluded)", len(result.Findings))
	}
}

func TestStorageScanner_Empty(t *testing.T) {
	s := NewStorageScanner(&mockStorageAPI{}, &mockMonitorAPI{}, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
