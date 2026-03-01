package analyzer

import azuretype "github.com/ppiankov/azurespectre/internal/azure"

// Analyze filters findings by minimum cost and computes summary statistics.
func Analyze(result *azuretype.ScanResult, cfg AnalyzerConfig) *AnalysisResult {
	var filtered []azuretype.Finding

	for _, f := range result.Findings {
		if f.EstimatedMonthlyWaste >= cfg.MinMonthlyCost || f.EstimatedMonthlyWaste == 0 {
			filtered = append(filtered, f)
		}
	}

	summary := Summary{
		TotalResourcesScanned: result.ResourcesScanned,
		TotalFindings:         len(filtered),
		BySeverity:            make(map[string]int),
		ByResourceType:        make(map[string]int),
		ByFindingType:         make(map[string]int),
	}

	for _, f := range filtered {
		summary.TotalMonthlyWaste += f.EstimatedMonthlyWaste
		summary.BySeverity[string(f.Severity)]++
		summary.ByResourceType[string(f.ResourceType)]++
		summary.ByFindingType[string(f.ID)]++
	}

	return &AnalysisResult{
		Findings: filtered,
		Summary:  summary,
		Errors:   result.Errors,
	}
}
