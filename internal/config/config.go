package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Config holds the wpgo application configuration.
type Config struct {
	DefaultFormat    string                      `json:"default_format"`
	DefaultSite      string                      `json:"default_site,omitempty"`
	RateLimits       map[string]RateLimitEntry   `json:"rate_limits,omitempty"`
	DefaultRateLimit RateLimitEntry              `json:"default_rate_limit"`
	CacheTTLs        CacheTTLConfig              `json:"cache_ttls"`
	Groups           map[string]GroupConfig       `json:"groups,omitempty"`
}

// RateLimitEntry configures rate limiting for a canonical host.
type RateLimitEntry struct {
	Delay    time.Duration `json:"delay"`
	MaxConns int           `json:"max_conns"`
}

// CacheTTLConfig holds TTL settings for each cacheable category.
type CacheTTLConfig struct {
	Plugins  time.Duration `json:"plugins"`
	Themes   time.Duration `json:"themes"`
	Core     time.Duration `json:"core"`
	Users    time.Duration `json:"users"`
	Options  time.Duration `json:"options"`
	Snapshot time.Duration `json:"snapshot"`
}

// GroupConfig defines a user-configured site group.
type GroupConfig struct {
	Aliases []string `json:"aliases,omitempty"`
}

// Load reads the config file and merges with defaults.
func Load(path string) (*Config, error) {
	defaults := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &defaults, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Merge with defaults for missing values.
	merged := mergeConfig(defaults, cfg)
	return &merged, nil
}

// Save writes the config to disk using atomic write.
func Save(cfg *Config, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// mergeConfig fills in missing values from defaults.
func mergeConfig(defaults, user Config) Config {
	if user.DefaultFormat == "" {
		user.DefaultFormat = defaults.DefaultFormat
	}
	if user.RateLimits == nil {
		user.RateLimits = defaults.RateLimits
	}
	if user.DefaultRateLimit.MaxConns == 0 {
		user.DefaultRateLimit = defaults.DefaultRateLimit
	}
	if user.CacheTTLs == (CacheTTLConfig{}) {
		user.CacheTTLs = defaults.CacheTTLs
	}
	if user.Groups == nil {
		user.Groups = defaults.Groups
	}
	return user
}

// UserGroups returns a map of group name → aliases for use with the registry.
func (c *Config) UserGroups() map[string][]string {
	groups := make(map[string][]string, len(c.Groups))
	for name, g := range c.Groups {
		groups[name] = g.Aliases
	}
	return groups
}
