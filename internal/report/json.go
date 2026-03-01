package report

import (
	"encoding/json"
	"fmt"
)

type jsonEnvelope struct {
	Schema string `json:"$schema"`
	Data
}

// Generate writes a spectre/v1 JSON report.
func (r *JSONReporter) Generate(data Data) error {
	envelope := jsonEnvelope{
		Schema: "spectre/v1",
		Data:   data,
	}
	enc := json.NewEncoder(r.Writer)
	enc.SetIndent("", "  ")
	if err := enc.Encode(envelope); err != nil {
		return fmt.Errorf("encode JSON: %w", err)
	}
	return nil
}
