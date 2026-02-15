package wpcli

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseJSON parses wp-cli JSON array output into a typed slice.
// It strips any non-JSON lines (wp-cli warnings) before the JSON array.
func ParseJSON[T any](output string) ([]T, error) {
	cleaned := stripWarnings(output)
	if cleaned == "" {
		return nil, nil
	}

	var result []T
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse wp-cli JSON array: %w", err)
	}
	return result, nil
}

// ParseSingle parses wp-cli JSON object output into a typed struct.
// It strips any non-JSON lines (wp-cli warnings) before the JSON object.
func ParseSingle[T any](output string) (*T, error) {
	cleaned := stripWarnings(output)
	if cleaned == "" {
		return nil, fmt.Errorf("empty wp-cli output")
	}

	var result T
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, fmt.Errorf("parse wp-cli JSON object: %w", err)
	}
	return &result, nil
}

// stripWarnings removes non-JSON lines from wp-cli output.
// wp-cli sometimes prints warnings or notices before the actual JSON data.
// We look for the first line starting with '[' or '{' and return everything
// from that point onward.
func stripWarnings(output string) string {
	output = strings.TrimSpace(output)
	if output == "" {
		return ""
	}

	// Fast path: output already starts with JSON
	if output[0] == '[' || output[0] == '{' {
		return output
	}

	// Scan line by line for JSON start
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) > 0 && (trimmed[0] == '[' || trimmed[0] == '{') {
			return strings.Join(lines[i:], "\n")
		}
	}

	return ""
}
