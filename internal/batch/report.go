package batch

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Report summarizes the results of a batch execution.
type Report struct {
	Results []SiteResult
	Total   int
	Success int
	Failed  int
	Skipped int
}

// NewReport builds a summary report from batch results.
func NewReport(results []SiteResult) *Report {
	r := &Report{
		Results: results,
		Total:   len(results),
	}
	for _, res := range results {
		switch res.Status {
		case StatusOK:
			r.Success++
		case StatusFailed:
			r.Failed++
		case StatusSkipped:
			r.Skipped++
		}
	}
	return r
}

// HasFailures returns true if any site failed.
func (r *Report) HasFailures() bool {
	return r.Failed > 0
}

// WriteTable writes a human-readable summary table to the writer.
//
// Format:
//
//	SITE            STATUS   DURATION   DETAILS
//	sitealpha     OK       1.2s       12 plugins listed
//	sitegamma   OK       2.1s       8 plugins listed
//	sitefail     FAILED   0.3s       connection refused
//	---
//	Total: 30 | Success: 28 | Failed: 2 | Skipped: 0
func (r *Report) WriteTable(w io.Writer) {
	// Calculate column widths.
	maxSite := len("SITE")
	maxDetail := len("DETAILS")
	for _, res := range r.Results {
		alias := res.Site.Alias
		if len(alias) > maxSite {
			maxSite = len(alias)
		}
		detail := truncate(res.Detail, 60)
		if len(detail) > maxDetail {
			maxDetail = len(detail)
		}
	}

	// Header.
	fmt.Fprintf(w, "%-*s  %-7s  %-10s  %s\n", maxSite, "SITE", "STATUS", "DURATION", "DETAILS")

	// Rows.
	for _, res := range r.Results {
		durStr := formatDuration(res.Duration)
		detail := truncate(res.Detail, 60)
		fmt.Fprintf(w, "%-*s  %-7s  %-10s  %s\n", maxSite, res.Site.Alias, res.Status, durStr, detail)
	}

	// Separator and summary.
	fmt.Fprintln(w, "---")
	fmt.Fprintf(w, "Total: %d | Success: %d | Failed: %d | Skipped: %d\n",
		r.Total, r.Success, r.Failed, r.Skipped)
}

// jsonReport is the serialization format for JSON output.
type jsonReport struct {
	Results []jsonResult `json:"results"`
	Summary jsonSummary  `json:"summary"`
}

type jsonResult struct {
	Site     string  `json:"site"`
	Status   string  `json:"status"`
	Duration float64 `json:"duration_seconds"`
	Detail   string  `json:"detail"`
	Error    string  `json:"error,omitempty"`
}

type jsonSummary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// WriteJSON writes a structured JSON report to the writer.
func (r *Report) WriteJSON(w io.Writer) error {
	jr := jsonReport{
		Results: make([]jsonResult, len(r.Results)),
		Summary: jsonSummary{
			Total:   r.Total,
			Success: r.Success,
			Failed:  r.Failed,
			Skipped: r.Skipped,
		},
	}

	for i, res := range r.Results {
		jr.Results[i] = jsonResult{
			Site:     res.Site.Alias,
			Status:   res.Status.String(),
			Duration: res.Duration.Seconds(),
			Detail:   res.Detail,
		}
		if res.Err != nil {
			jr.Results[i].Error = res.Err.Error()
		}
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jr)
}

// formatDuration returns a human-readable duration string.
func formatDuration(d interface{ Seconds() float64 }) string {
	secs := d.Seconds()
	if secs < 0.01 {
		return "0.0s"
	}
	return fmt.Sprintf("%.1fs", secs)
}

// truncate shortens a string to maxLen, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 4 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
