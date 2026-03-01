package azure

import (
	"context"
	"testing"
)

func TestSQLScanner_Type(t *testing.T) {
	s := NewSQLScanner(&mockSQLAPI{}, nil, "sub-1")
	if s.Type() != ResourceSQLDatabase {
		t.Errorf("Type() = %q, want %q", s.Type(), ResourceSQLDatabase)
	}
}

func TestSQLScanner_IdleSQL(t *testing.T) {
	dbID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Sql/servers/srv1/databases/db1"
	sqlAPI := &mockSQLAPI{
		databases: []SQLDatabase{
			{
				ID:            dbID,
				Name:          "db1",
				ResourceGroup: "rg1",
				Location:      "eastus",
				ServerName:    "srv1",
				SKUName:       "S0",
				SKUTier:       "Standard_S0",
				Capacity:      10,
			},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{dbID: 2.1},
	}

	s := NewSQLScanner(sqlAPI, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{IdleCPU: 5, IdleDays: 7})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 1 {
		t.Fatalf("findings = %d, want 1", len(result.Findings))
	}
	f := result.Findings[0]
	if f.ID != FindingIdleSQL {
		t.Errorf("finding ID = %q, want %q", f.ID, FindingIdleSQL)
	}
	if f.ResourceName != "srv1/db1" {
		t.Errorf("resource name = %q, want %q", f.ResourceName, "srv1/db1")
	}
}

func TestSQLScanner_BusySQL(t *testing.T) {
	dbID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Sql/servers/srv1/databases/db2"
	sqlAPI := &mockSQLAPI{
		databases: []SQLDatabase{
			{ID: dbID, Name: "db2", ServerName: "srv1", Location: "eastus"},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{dbID: 45.0},
	}

	s := NewSQLScanner(sqlAPI, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{IdleCPU: 5})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}

func TestSQLScanner_NilMonitor(t *testing.T) {
	sqlAPI := &mockSQLAPI{
		databases: []SQLDatabase{
			{ID: "db-1", Name: "db1", ServerName: "srv1"},
		},
	}

	s := NewSQLScanner(sqlAPI, nil, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (no monitor)", len(result.Findings))
	}
}

func TestSQLScanner_ExcludeByID(t *testing.T) {
	dbID := "/subscriptions/sub-1/resourceGroups/rg1/providers/Microsoft.Sql/servers/srv1/databases/db-excluded"
	sqlAPI := &mockSQLAPI{
		databases: []SQLDatabase{
			{ID: dbID, Name: "db-excluded", ServerName: "srv1"},
		},
	}
	monitor := &mockMonitorAPI{
		results: map[string]float64{dbID: 1.0},
	}

	s := NewSQLScanner(sqlAPI, monitor, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{
		Exclude: ExcludeConfig{ResourceIDs: map[string]bool{dbID: true}},
	})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0 (excluded)", len(result.Findings))
	}
}

func TestSQLScanner_Empty(t *testing.T) {
	s := NewSQLScanner(&mockSQLAPI{}, &mockMonitorAPI{}, "sub-1")
	result, err := s.Scan(context.Background(), ScanConfig{})
	if err != nil {
		t.Fatalf("Scan() error: %v", err)
	}
	if len(result.Findings) != 0 {
		t.Errorf("findings = %d, want 0", len(result.Findings))
	}
}
