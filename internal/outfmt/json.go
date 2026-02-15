package outfmt

import (
	"encoding/json"
	"fmt"
	"io"
)

// formatJSON writes data as pretty-printed JSON.
func formatJSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("format JSON: %w", err)
	}
	return nil
}
