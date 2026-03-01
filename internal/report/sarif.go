package report

import (
	"encoding/json"
	"fmt"

	azuretype "github.com/ppiankov/azurespectre/internal/azure"
)

type sarifReport struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []sarifRule `json:"rules"`
}

type sarifRule struct {
	ID               string          `json:"id"`
	ShortDescription sarifMessage    `json:"shortDescription"`
	DefaultConfig    sarifRuleConfig `json:"defaultConfiguration"`
	Properties       sarifRuleProps  `json:"properties,omitempty"`
}

type sarifRuleConfig struct {
	Level string `json:"level"`
}

type sarifRuleProps struct {
	Tags []string `json:"tags,omitempty"`
}

type sarifResult struct {
	RuleID     string          `json:"ruleId"`
	Level      string          `json:"level"`
	Message    sarifMessage    `json:"message"`
	Locations  []sarifLocation `json:"locations,omitempty"`
	Properties map[string]any  `json:"properties,omitempty"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

func sarifLevel(sev azuretype.Severity) string {
	switch sev {
	case azuretype.SeverityHigh:
		return "error"
	case azuretype.SeverityMedium:
		return "warning"
	case azuretype.SeverityLow:
		return "note"
	default:
		return "none"
	}
}

func buildSARIFRules(findings []azuretype.Finding) []sarifRule {
	seen := make(map[azuretype.FindingID]bool)
	var rules []sarifRule
	for _, f := range findings {
		if seen[f.ID] {
			continue
		}
		seen[f.ID] = true
		rules = append(rules, sarifRule{
			ID:               string(f.ID),
			ShortDescription: sarifMessage{Text: string(f.ID)},
			DefaultConfig:    sarifRuleConfig{Level: sarifLevel(f.Severity)},
			Properties:       sarifRuleProps{Tags: []string{"azure", "waste"}},
		})
	}
	return rules
}

// Generate writes a SARIF v2.1.0 report.
func (r *SARIFReporter) Generate(data Data) error {
	var results []sarifResult
	for _, f := range data.Findings {
		uri := fmt.Sprintf("azure://%s/%s/%s/%s",
			f.Subscription, f.Region, f.ResourceType, f.ResourceName)
		results = append(results, sarifResult{
			RuleID:  string(f.ID),
			Level:   sarifLevel(f.Severity),
			Message: sarifMessage{Text: f.Message},
			Locations: []sarifLocation{{
				PhysicalLocation: sarifPhysicalLocation{
					ArtifactLocation: sarifArtifactLocation{URI: uri},
				},
			}},
			Properties: map[string]any{
				"estimated_monthly_waste": f.EstimatedMonthlyWaste,
				"resource_id":             f.ResourceID,
			},
		})
	}

	report := sarifReport{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{{
			Tool: sarifTool{
				Driver: sarifDriver{
					Name:    data.Tool,
					Version: data.Version,
					Rules:   buildSARIFRules(data.Findings),
				},
			},
			Results: results,
		}},
	}

	enc := json.NewEncoder(r.Writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("encode SARIF: %w", err)
	}
	return nil
}
