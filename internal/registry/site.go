package registry

import "time"

// Site represents a merged view of SSH config + metadata overlay for a WordPress site.
type Site struct {
	Alias         string            // SSH alias (from ~/.ssh/config)
	Hostname      string            // Resolved hostname
	Port          int               // SSH port (default 22)
	User          string            // SSH user
	IdentityFile  string            // Path to private key
	WPPath        string            // WordPress install path (e.g., ~/public_html)
	HostType      string            // "standard", "wpengine", "auto" (detected)
	Groups        []string          // Group memberships
	Tags          map[string]string // Arbitrary tags
	RateLimit     *RateLimitConfig  // Per-site rate limit override (nil = use host default)
	CanonicalHost string            // Resolved IP:port for rate limiting
}

// RateLimitConfig holds per-site rate limit overrides.
type RateLimitConfig struct {
	Delay    time.Duration
	MaxConns int
}
