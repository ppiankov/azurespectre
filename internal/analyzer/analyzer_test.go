package analyzer

import (
	"testing"

	azuretype "github.com/ppiankov/azurespectre/internal/azure"
)

func TestAnalyze(t *testing.T) {
	result := &azuretype.ScanResult{
		Findings: []azuretype.Finding{
			{ID: azuretype.FindingStoppedVM, Severity: azuretype.SeverityHigh, ResourceType: azuretype.ResourceVM, EstimatedMonthlyWaste: 30.0},
			{ID: azuretype.FindingUnattachedDisk, Severity: azuretype.SeverityHigh, ResourceType: azuretype.ResourceDisk, EstimatedMonthlyWaste: 17.0},
			{ID: azuretype.FindingUnusedNSG, Severity: azuretype.SeverityLow, ResourceType: azuretype.ResourceNSG, EstimatedMonthlyWaste: 0},
		},
		ResourcesScanned: 20,
	}

	analysis := Analyze(result, AnalyzerConfig{MinMonthlyCost: 5.0})

	if analysis.Summary.TotalFindings != 3 {
		t.Errorf("findings = %d, want 3", analysis.Summary.TotalFindings)
	}
	if analysis.Summary.TotalMonthlyWaste != 47.0 {
		t.Errorf("waste = %.2f, want 47.00", analysis.Summary.TotalMonthlyWaste)
	}
	if analysis.Summary.BySeverity["high"] != 2 {
		t.Errorf("high severity = %d, want 2", analysis.Summary.BySeverity["high"])
	}
}

func TestAnalyzeFiltersLowCost(t *testing.T) {
	result := &azuretype.ScanResult{
		Findings: []azuretype.Finding{
			{ID: azuretype.FindingStoppedVM, EstimatedMonthlyWaste: 30.0},
			{ID: azuretype.FindingUnusedIP, EstimatedMonthlyWaste: 2.0},
		},
		ResourcesScanned: 10,
	}

	analysis := Analyze(result, AnalyzerConfig{MinMonthlyCost: 5.0})

	if analysis.Summary.TotalFindings != 1 {
		t.Errorf("findings = %d, want 1 (filtered low cost)", analysis.Summary.TotalFindings)
	}
}

func TestAnalyzeKeepsZeroCost(t *testing.T) {
	result := &azuretype.ScanResult{
		Findings: []azuretype.Finding{
			{ID: azuretype.FindingUnusedNSG, EstimatedMonthlyWaste: 0},
		},
		ResourcesScanned: 5,
	}

	analysis := Analyze(result, AnalyzerConfig{MinMonthlyCost: 5.0})

	if analysis.Summary.TotalFindings != 1 {
		t.Errorf("findings = %d, want 1 (zero cost kept)", analysis.Summary.TotalFindings)
	}
}

func TestAnalyzeEmpty(t *testing.T) {
	result := &azuretype.ScanResult{}
	analysis := Analyze(result, AnalyzerConfig{})
	if analysis.Summary.TotalFindings != 0 {
		t.Errorf("findings = %d, want 0", analysis.Summary.TotalFindings)
	}
}

func TestAnalyzeErrors(t *testing.T) {
	result := &azuretype.ScanResult{
		Errors: []string{"compute failed"},
	}
	analysis := Analyze(result, AnalyzerConfig{})
	if len(analysis.Errors) != 1 {
		t.Errorf("errors = %d, want 1", len(analysis.Errors))
	}
}
