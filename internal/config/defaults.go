package config

import "time"

// DefaultConfig returns the default application configuration.
func DefaultConfig() Config {
	return Config{
		DefaultFormat: "table",
		DefaultSite:   "",
		RateLimits: map[string]RateLimitEntry{
			"192.0.2.10": {
				Delay:    3 * time.Second,
				MaxConns: 1,
			},
		},
		DefaultRateLimit: RateLimitEntry{
			Delay:    500 * time.Millisecond,
			MaxConns: 3,
		},
		CacheTTLs: CacheTTLConfig{
			Plugins:  1 * time.Hour,
			Themes:   1 * time.Hour,
			Core:     24 * time.Hour,
			Users:    6 * time.Hour,
			Options:  30 * time.Minute,
			Snapshot: 4 * time.Hour,
		},
		Groups: map[string]GroupConfig{},
	}
}
