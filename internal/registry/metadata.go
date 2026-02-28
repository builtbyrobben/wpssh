package registry

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Metadata holds the wpgo sites.json overlay data.
type Metadata struct {
	Sites map[string]SiteOverlay `json:"sites"`
}

// SiteOverlay contains WordPress-specific metadata that overlays SSH config entries.
type SiteOverlay struct {
	WPPath    string            `json:"wp_path,omitempty"`
	HostType  string            `json:"host_type,omitempty"`
	Groups    []string          `json:"groups,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	RateLimit *RateLimitOverlay `json:"rate_limit,omitempty"`
}

// RateLimitOverlay is the JSON-serializable rate limit config for a site.
type RateLimitOverlay struct {
	DelayMs  int `json:"delay_ms,omitempty"`
	MaxConns int `json:"max_conns,omitempty"`
}

// ToRateLimitConfig converts a RateLimitOverlay to a RateLimitConfig.
func (r *RateLimitOverlay) ToRateLimitConfig() *RateLimitConfig {
	if r == nil {
		return nil
	}
	return &RateLimitConfig{
		Delay:    time.Duration(r.DelayMs) * time.Millisecond,
		MaxConns: r.MaxConns,
	}
}

// LoadMetadata reads sites.json from the given path. Returns an empty Metadata
// if the file does not exist.
func LoadMetadata(path string) (*Metadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Metadata{Sites: map[string]SiteOverlay{}}, nil
		}
		return nil, err
	}
	var m Metadata
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	if m.Sites == nil {
		m.Sites = map[string]SiteOverlay{}
	}
	return &m, nil
}

// SaveMetadata writes sites.json to the given path using atomic write
// (write to temp file, then rename).
func SaveMetadata(m *Metadata, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	// Atomic write: temp file + rename.
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
