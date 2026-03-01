package report

import (
	"io"
	"time"

	"github.com/ppiankov/azurespectre/internal/analyzer"
	azuretype "github.com/ppiankov/azurespectre/internal/azure"
)

// Reporter generates output in a specific format.
type Reporter interface {
	Generate(data Data) error
}

// Data contains everything needed to produce a report.
type Data struct {
	Tool      string              `json:"tool"`
	Version   string              `json:"version"`
	Timestamp time.Time           `json:"timestamp"`
	Target    Target              `json:"target"`
	Config    ReportConfig        `json:"config"`
	Findings  []azuretype.Finding `json:"findings"`
	Summary   analyzer.Summary    `json:"summary"`
	Errors    []string            `json:"errors,omitempty"`
}

// Target identifies the scan scope.
type Target struct {
	Type    string `json:"type"`
	URIHash string `json:"uri_hash"`
}

// ReportConfig captures scan parameters.
type ReportConfig struct {
	Subscription   string  `json:"subscription"`
	ResourceGroup  string  `json:"resource_group,omitempty"`
	IdleDays       int     `json:"idle_days"`
	StaleDays      int     `json:"stale_days"`
	StoppedDays    int     `json:"stopped_days"`
	MinMonthlyCost float64 `json:"min_monthly_cost"`
}

// TextReporter produces human-readable output.
type TextReporter struct{ Writer io.Writer }

// JSONReporter produces spectre/v1 JSON output.
type JSONReporter struct{ Writer io.Writer }

// SARIFReporter produces SARIF v2.1.0 output.
type SARIFReporter struct{ Writer io.Writer }

// SpectreHubReporter produces SpectreHub envelope output.
type SpectreHubReporter struct{ Writer io.Writer }
