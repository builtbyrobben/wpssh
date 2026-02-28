package cache

import "time"

// Default TTLs for cacheable command categories.
const (
	TTLPlugins  = 3600  // 1 hour
	TTLThemes   = 3600  // 1 hour
	TTLCore     = 86400 // 24 hours
	TTLUsers    = 21600 // 6 hours
	TTLOptions  = 1800  // 30 minutes
	TTLSnapshot = 14400 // 4 hours
)

// Cache categories for invalidation grouping.
const (
	CategoryPlugins  = "plugins"
	CategoryThemes   = "themes"
	CategoryCore     = "core"
	CategoryUsers    = "users"
	CategoryOptions  = "options"
	CategorySnapshot = "snapshot"
)

// CacheEntry represents a single cached wp-cli query result.
type CacheEntry struct {
	SiteAlias  string
	CacheKey   string
	Command    string // Human-readable: "plugin list --status=active"
	Data       string // JSON blob
	FetchedAt  time.Time
	TTLSeconds int
	Category   string // "plugins", "themes", "core", "users", "options", "snapshot"
}

// SiteSnapshot holds aggregate site state for quick overview.
type SiteSnapshot struct {
	SiteAlias   string
	CoreVersion string
	PHPVersion  string
	PluginCount int
	ThemeCount  int
	DBSize      string
	LastChecked time.Time
}

// CacheStats holds cache statistics for reporting.
type CacheStats struct {
	TotalEntries   int
	ExpiredEntries int
	SiteCount      int
	DBSizeBytes    int64
	Categories     map[string]int // count per category
}
