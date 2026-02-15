package outfmt

import (
	"fmt"
	"io"
	"strings"
)

// formatTable writes data as a colored, aligned table.
func formatTable(w io.Writer, data any) error {
	rows, headers := toRows(data)
	if len(headers) == 0 || len(rows) == 0 {
		return nil
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, val := range row {
			if i >= len(widths) {
				break
			}
			s := fmt.Sprint(val)
			if len(s) > widths[i] {
				widths[i] = len(s)
			}
		}
	}

	colors := newColorHelper()

	// Print header
	headerParts := make([]string, len(headers))
	for i, h := range headers {
		headerParts[i] = padRight(strings.ToUpper(h), widths[i])
	}
	fmt.Fprintln(w, colors.Header(strings.Join(headerParts, "  ")))

	// Print separator
	sepParts := make([]string, len(headers))
	for i, width := range widths {
		sepParts[i] = strings.Repeat("-", width)
	}
	fmt.Fprintln(w, strings.Join(sepParts, "  "))

	// Print rows
	for _, row := range rows {
		parts := make([]string, len(headers))
		for i, val := range row {
			if i >= len(headers) {
				break
			}
			s := fmt.Sprint(val)
			colored := colors.StatusColor(s)
			// If color was applied, pad based on original string length
			if colored != s {
				padding := widths[i] - len(s)
				if padding > 0 {
					parts[i] = colored + strings.Repeat(" ", padding)
				} else {
					parts[i] = colored
				}
			} else {
				parts[i] = padRight(s, widths[i])
			}
		}
		fmt.Fprintln(w, strings.Join(parts, "  "))
	}

	return nil
}

// padRight pads a string with spaces to reach the desired width.
func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}
