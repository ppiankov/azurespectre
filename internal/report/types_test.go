package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/ppiankov/azurespectre/internal/analyzer"
	azuretype "github.com/ppiankov/azurespectre/internal/azure"
)

func sampleData() Data {
	return Data{
		Tool:      "azurespectre",
		Version:   "0.1.0",
		Timestamp: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		Target:    Target{Type: "azure-subscription", URIHash: "sha256:abc"},
		Config: ReportConfig{
			Subscription:   "sub-1",
			IdleDays:       7,
			StaleDays:      90,
			StoppedDays:    30,
			MinMonthlyCost: 5.0,
		},
		Findings: []azuretype.Finding{
			{
				ID: azuretype.FindingStoppedVM, Severity: azuretype.SeverityHigh,
				ResourceType: azuretype.ResourceVM, ResourceID: "vm1-id",
				ResourceName: "vm1", Subscription: "sub-1", Region: "eastus",
				Message: "Deallocated for 60 days", EstimatedMonthlyWaste: 30.37,
			},
			{
				ID: azuretype.FindingUnattachedDisk, Severity: azuretype.SeverityHigh,
				ResourceType: azuretype.ResourceDisk, ResourceID: "disk1-id",
				ResourceName: "disk1", Subscription: "sub-1", Region: "eastus",
				Message: "Unattached Premium_LRS disk (128 GB)", EstimatedMonthlyWaste: 17.28,
			},
		},
		Summary: analyzer.Summary{
			TotalResourcesScanned: 10,
			TotalFindings:         2,
			TotalMonthlyWaste:     47.65,
			BySeverity:            map[string]int{"high": 2},
			ByResourceType:        map[string]int{"virtual_machine": 1, "managed_disk": 1},
			ByFindingType:         map[string]int{"STOPPED_VM": 1, "UNATTACHED_DISK": 1},
		},
	}
}

func TestJSONReporter(t *testing.T) {
	var buf bytes.Buffer
	r := &JSONReporter{Writer: &buf}
	if err := r.Generate(sampleData()); err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["$schema"] != "spectre/v1" {
		t.Errorf("schema = %v, want spectre/v1", parsed["$schema"])
	}
}

func TestTextReporter(t *testing.T) {
	var buf bytes.Buffer
	r := &TextReporter{Writer: &buf}
	if err := r.Generate(sampleData()); err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "azurespectre scan results") {
		t.Error("missing header")
	}
	if !strings.Contains(out, "vm1") {
		t.Error("missing finding")
	}
	if !strings.Contains(out, "Summary:") {
		t.Error("missing summary")
	}
}

func TestTextReporterNoFindings(t *testing.T) {
	data := sampleData()
	data.Findings = nil
	var buf bytes.Buffer
	r := &TextReporter{Writer: &buf}
	if err := r.Generate(data); err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !strings.Contains(buf.String(), "No findings") {
		t.Error("should show 'No findings'")
	}
}

func TestSARIFReporter(t *testing.T) {
	var buf bytes.Buffer
	r := &SARIFReporter{Writer: &buf}
	if err := r.Generate(sampleData()); err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["version"] != "2.1.0" {
		t.Errorf("version = %v, want 2.1.0", parsed["version"])
	}
}

func TestSARIFLevel(t *testing.T) {
	tests := []struct {
		severity azuretype.Severity
		want     string
	}{
		{azuretype.SeverityHigh, "error"},
		{azuretype.SeverityMedium, "warning"},
		{azuretype.SeverityLow, "note"},
		{"unknown", "none"},
	}
	for _, tt := range tests {
		got := sarifLevel(tt.severity)
		if got != tt.want {
			t.Errorf("sarifLevel(%q) = %q, want %q", tt.severity, got, tt.want)
		}
	}
}

func TestSARIFRulesDeduplicate(t *testing.T) {
	findings := []azuretype.Finding{
		{ID: azuretype.FindingUnusedIP, Severity: azuretype.SeverityMedium},
		{ID: azuretype.FindingUnusedIP, Severity: azuretype.SeverityMedium},
	}
	rules := buildSARIFRules(findings)
	if len(rules) != 1 {
		t.Errorf("rules = %d, want 1 (deduplicated)", len(rules))
	}
}

func TestSpectreHubReporter(t *testing.T) {
	var buf bytes.Buffer
	r := &SpectreHubReporter{Writer: &buf}
	if err := r.Generate(sampleData()); err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if parsed["schema"] != "spectre/v1" {
		t.Errorf("schema = %v, want spectre/v1", parsed["schema"])
	}
}

func TestSpectreHubNoFindings(t *testing.T) {
	data := sampleData()
	data.Findings = nil
	var buf bytes.Buffer
	r := &SpectreHubReporter{Writer: &buf}
	if err := r.Generate(data); err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if !strings.Contains(buf.String(), "spectre/v1") {
		t.Error("missing schema")
	}
}

func TestTextReporterWithErrors(t *testing.T) {
	data := sampleData()
	data.Errors = []string{"compute API timeout"}
	var buf bytes.Buffer
	r := &TextReporter{Writer: &buf}
	if err := r.Generate(data); err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "Errors:") {
		t.Error("missing errors section")
	}
	if !strings.Contains(out, "timeout") {
		t.Error("missing error message")
	}
}
