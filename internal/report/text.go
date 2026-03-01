package report

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// Generate writes a human-readable text report.
func (r *TextReporter) Generate(data Data) error {
	w := tabwriter.NewWriter(r.Writer, 0, 0, 2, ' ', 0)

	_, _ = fmt.Fprintf(r.Writer, "\nazurespectre scan results\n")
	_, _ = fmt.Fprintf(r.Writer, "%s\n\n", strings.Repeat("=", 50))

	if len(data.Findings) == 0 {
		_, _ = fmt.Fprintf(r.Writer, "No findings detected.\n\n")
	} else {
		_, _ = fmt.Fprintf(w, "SEVERITY\tTYPE\tRESOURCE\tREGION\tWASTE/MO\tMESSAGE\n")
		_, _ = fmt.Fprintf(w, "--------\t----\t--------\t------\t--------\t-------\n")
		for _, f := range data.Findings {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t$%.2f\t%s\n",
				f.Severity, f.ResourceType, f.ResourceName, f.Region,
				f.EstimatedMonthlyWaste, f.Message)
		}
		_ = w.Flush()
		_, _ = fmt.Fprintln(r.Writer)
	}

	if len(data.Errors) > 0 {
		_, _ = fmt.Fprintf(r.Writer, "Errors:\n")
		for _, e := range data.Errors {
			_, _ = fmt.Fprintf(r.Writer, "  - %s\n", e)
		}
		_, _ = fmt.Fprintln(r.Writer)
	}

	_, _ = fmt.Fprintf(r.Writer, "Summary:\n")
	_, _ = fmt.Fprintf(r.Writer, "  Resources scanned:       %d\n", data.Summary.TotalResourcesScanned)
	_, _ = fmt.Fprintf(r.Writer, "  Total findings:          %d\n", data.Summary.TotalFindings)
	_, _ = fmt.Fprintf(r.Writer, "  Estimated monthly waste: $%.2f\n", data.Summary.TotalMonthlyWaste)

	if len(data.Summary.BySeverity) > 0 {
		_, _ = fmt.Fprintf(r.Writer, "  By severity:             ")
		first := true
		for sev, count := range data.Summary.BySeverity {
			if !first {
				_, _ = fmt.Fprintf(r.Writer, ", ")
			}
			_, _ = fmt.Fprintf(r.Writer, "%s=%d", sev, count)
			first = false
		}
		_, _ = fmt.Fprintln(r.Writer)
	}

	if len(data.Summary.ByResourceType) > 0 {
		_, _ = fmt.Fprintf(r.Writer, "  By resource type:        ")
		first := true
		for rt, count := range data.Summary.ByResourceType {
			if !first {
				_, _ = fmt.Fprintf(r.Writer, ", ")
			}
			_, _ = fmt.Fprintf(r.Writer, "%s=%d", rt, count)
			first = false
		}
		_, _ = fmt.Fprintln(r.Writer)
	}

	return nil
}
