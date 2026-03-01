package analyzer

import azuretype "github.com/ppiankov/azurespectre/internal/azure"

// Summary contains aggregated scan statistics.
type Summary struct {
	TotalResourcesScanned int            `json:"total_resources_scanned"`
	TotalFindings         int            `json:"total_findings"`
	TotalMonthlyWaste     float64        `json:"total_monthly_waste"`
	BySeverity            map[string]int `json:"by_severity"`
	ByResourceType        map[string]int `json:"by_resource_type"`
	ByFindingType         map[string]int `json:"by_finding_type"`
}

// AnalysisResult holds filtered findings and summary.
type AnalysisResult struct {
	Findings []azuretype.Finding `json:"findings"`
	Summary  Summary             `json:"summary"`
	Errors   []string            `json:"errors,omitempty"`
}

// AnalyzerConfig controls analysis behavior.
type AnalyzerConfig struct {
	MinMonthlyCost float64
}
