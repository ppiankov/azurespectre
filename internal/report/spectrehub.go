package report

import (
	"encoding/json"
	"fmt"
)

// Generate writes a spectre/v1 SpectreHub envelope.
func (r *SpectreHubReporter) Generate(data Data) error {
	envelope := map[string]any{
		"schema":    "spectre/v1",
		"tool":      data.Tool,
		"version":   data.Version,
		"timestamp": data.Timestamp,
		"target":    data.Target,
		"config":    data.Config,
		"findings":  data.Findings,
		"summary":   data.Summary,
		"errors":    data.Errors,
	}
	enc := json.NewEncoder(r.Writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(envelope); err != nil {
		return fmt.Errorf("encode SpectreHub: %w", err)
	}
	return nil
}
