package outfmt

import (
	"os"
	"strings"

	"github.com/muesli/termenv"
)

// colorHelper provides terminal color styling that respects NO_COLOR.
type colorHelper struct {
	profile termenv.Profile
	enabled bool
}

// newColorHelper creates a color helper that respects the NO_COLOR env var.
func newColorHelper() *colorHelper {
	_, noColor := os.LookupEnv("NO_COLOR")
	output := termenv.NewOutput(os.Stdout)
	return &colorHelper{
		profile: output.Profile,
		enabled: !noColor,
	}
}

// Header styles text as a bold header.
func (c *colorHelper) Header(s string) string {
	if !c.enabled {
		return s
	}
	output := termenv.NewOutput(os.Stdout)
	return output.String(s).Bold().String()
}

// StatusColor applies color based on common WordPress status values.
func (c *colorHelper) StatusColor(s string) string {
	if !c.enabled {
		return s
	}

	output := termenv.NewOutput(os.Stdout)
	lower := strings.ToLower(s)

	switch lower {
	case "active", "on", "available", "publish", "approved":
		return output.String(s).Foreground(output.Color("2")).String() // green
	case "inactive", "off", "none", "draft", "pending":
		return output.String(s).Foreground(output.Color("3")).String() // yellow
	case "must-use", "dropin", "parent":
		return output.String(s).Foreground(output.Color("6")).String() // cyan
	case "error", "failed", "trash", "spam":
		return output.String(s).Foreground(output.Color("1")).String() // red
	default:
		return s
	}
}

// Green colors text green.
func (c *colorHelper) Green(s string) string {
	if !c.enabled {
		return s
	}
	output := termenv.NewOutput(os.Stdout)
	return output.String(s).Foreground(output.Color("2")).String()
}

// Yellow colors text yellow.
func (c *colorHelper) Yellow(s string) string {
	if !c.enabled {
		return s
	}
	output := termenv.NewOutput(os.Stdout)
	return output.String(s).Foreground(output.Color("3")).String()
}

// Red colors text red.
func (c *colorHelper) Red(s string) string {
	if !c.enabled {
		return s
	}
	output := termenv.NewOutput(os.Stdout)
	return output.String(s).Foreground(output.Color("1")).String()
}
