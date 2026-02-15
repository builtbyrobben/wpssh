package outfmt

import (
	"io"
	"os"
)

// Formatter dispatches output formatting based on configured flags.
type Formatter struct {
	JSON   bool
	Plain  bool
	Fields string // comma-separated field names
	Writer io.Writer
}

// New creates a Formatter with the given options writing to stdout.
func New(jsonFlag, plainFlag bool, fields string) *Formatter {
	return &Formatter{
		JSON:   jsonFlag,
		Plain:  plainFlag,
		Fields: fields,
		Writer: os.Stdout,
	}
}

// NewWithWriter creates a Formatter writing to a specific writer.
func NewWithWriter(jsonFlag, plainFlag bool, fields string, w io.Writer) *Formatter {
	return &Formatter{
		JSON:   jsonFlag,
		Plain:  plainFlag,
		Fields: fields,
		Writer: w,
	}
}

// Format outputs data in the configured format.
// data should be a struct, slice of structs, or []map[string]any.
func (f *Formatter) Format(data any) error {
	// Apply field projection if requested
	projected, err := projectFields(data, f.Fields)
	if err != nil {
		return err
	}

	if f.JSON {
		return formatJSON(f.Writer, projected)
	}
	if f.Plain {
		return formatPlain(f.Writer, projected)
	}
	return formatTable(f.Writer, projected)
}
