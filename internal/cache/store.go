package cache

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"
)

// parseTime parses a datetime string from SQLite, supporting both
// "2006-01-02 15:04:05" and RFC3339 formats.
func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", s)
}

// MakeCacheKey generates a deterministic cache key from the site alias and
// the normalized command form (from wpcli.Command.CacheKey()).
// key = SHA256(siteAlias + ":" + commandCacheKey)
func MakeCacheKey(siteAlias, commandCacheKey string) string {
	h := sha256.Sum256([]byte(siteAlias + ":" + commandCacheKey))
	return fmt.Sprintf("%x", h)
}

// Get retrieves a cached entry if it exists and hasn't expired.
// Returns nil, nil if not found or expired (caller should fetch fresh).
func (c *Cache) Get(siteAlias, cacheKey string) (*CacheEntry, error) {
	row := c.db.QueryRow(queryGetEntry, siteAlias, cacheKey)

	var e CacheEntry
	var fetchedAt string
	err := row.Scan(&e.SiteAlias, &e.CacheKey, &e.Command, &e.Data, &fetchedAt, &e.TTLSeconds, &e.Category)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("cache get: %w", err)
	}

	t, err := parseTime(fetchedAt)
	if err != nil {
		return nil, fmt.Errorf("parse fetched_at: %w", err)
	}
	e.FetchedAt = t

	return &e, nil
}

// Set stores a cache entry with the given TTL.
func (c *Cache) Set(entry *CacheEntry) error {
	fetchedAt := entry.FetchedAt.UTC().Format("2006-01-02 15:04:05")
	_, err := c.db.Exec(querySetEntry,
		entry.SiteAlias,
		entry.CacheKey,
		entry.Command,
		entry.Data,
		fetchedAt,
		entry.TTLSeconds,
		entry.Category,
	)
	if err != nil {
		return fmt.Errorf("cache set: %w", err)
	}
	return nil
}

// Invalidate removes all cache entries for a site in the given categories.
// If categories is nil, invalidates ALL entries for the site.
func (c *Cache) Invalidate(siteAlias string, categories []string) error {
	if len(categories) == 0 {
		_, err := c.db.Exec(queryInvalidateSite, siteAlias)
		if err != nil {
			return fmt.Errorf("cache invalidate site: %w", err)
		}
		return nil
	}

	// Build parameterized query for category list.
	placeholders := make([]string, len(categories))
	args := make([]any, 0, len(categories)+1)
	args = append(args, siteAlias)
	for i, cat := range categories {
		placeholders[i] = "?"
		args = append(args, cat)
	}
	query := fmt.Sprintf(
		"DELETE FROM cache_entries WHERE site_alias = ? AND category IN (%s)",
		strings.Join(placeholders, ", "),
	)
	_, err := c.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("cache invalidate categories: %w", err)
	}
	return nil
}

// InvalidateCategory removes all cache entries for a site in a single category.
func (c *Cache) InvalidateCategory(siteAlias, category string) error {
	return c.Invalidate(siteAlias, []string{category})
}

// InvalidateSite removes ALL cache entries for a site (all categories).
func (c *Cache) InvalidateSite(siteAlias string) error {
	return c.Invalidate(siteAlias, nil)
}

// InvalidateAll removes all cache entries.
func (c *Cache) InvalidateAll() error {
	_, err := c.db.Exec(queryInvalidateAll)
	if err != nil {
		return fmt.Errorf("cache invalidate all: %w", err)
	}
	return nil
}

// Stats returns cache statistics.
func (c *Cache) Stats() (*CacheStats, error) {
	s := &CacheStats{
		Categories: make(map[string]int),
	}

	if err := c.db.QueryRow(queryStatsTotal).Scan(&s.TotalEntries); err != nil {
		return nil, fmt.Errorf("stats total: %w", err)
	}
	if err := c.db.QueryRow(queryStatsExpired).Scan(&s.ExpiredEntries); err != nil {
		return nil, fmt.Errorf("stats expired: %w", err)
	}
	if err := c.db.QueryRow(queryStatsSites).Scan(&s.SiteCount); err != nil {
		return nil, fmt.Errorf("stats sites: %w", err)
	}

	rows, err := c.db.Query(queryStatsCategories)
	if err != nil {
		return nil, fmt.Errorf("stats categories: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var cat string
		var count int
		if err := rows.Scan(&cat, &count); err != nil {
			return nil, fmt.Errorf("stats category scan: %w", err)
		}
		s.Categories[cat] = count
	}

	// Get database file size from the pragma.
	var pageCount, pageSize int64
	if err := c.db.QueryRow("PRAGMA page_count").Scan(&pageCount); err == nil {
		if err := c.db.QueryRow("PRAGMA page_size").Scan(&pageSize); err == nil {
			s.DBSizeBytes = pageCount * pageSize
		}
	}

	return s, nil
}

// Prune removes all expired entries and returns the number of rows deleted.
func (c *Cache) Prune() (int, error) {
	res, err := c.db.Exec(queryPrune)
	if err != nil {
		return 0, fmt.Errorf("cache prune: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("prune rows affected: %w", err)
	}
	return int(n), nil
}

// GetSnapshot retrieves a site snapshot.
// Returns nil, nil if no snapshot exists for the site.
func (c *Cache) GetSnapshot(siteAlias string) (*SiteSnapshot, error) {
	row := c.db.QueryRow(queryGetSnapshot, siteAlias)

	var snap SiteSnapshot
	var lastChecked string
	err := row.Scan(&snap.SiteAlias, &snap.CoreVersion, &snap.PHPVersion,
		&snap.PluginCount, &snap.ThemeCount, &snap.DBSize, &lastChecked)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get snapshot: %w", err)
	}

	t, err := parseTime(lastChecked)
	if err != nil {
		return nil, fmt.Errorf("parse last_checked: %w", err)
	}
	snap.LastChecked = t

	return &snap, nil
}

// SetSnapshot stores a site snapshot.
func (c *Cache) SetSnapshot(snapshot *SiteSnapshot) error {
	lastChecked := snapshot.LastChecked.UTC().Format("2006-01-02 15:04:05")
	_, err := c.db.Exec(querySetSnapshot,
		snapshot.SiteAlias,
		snapshot.CoreVersion,
		snapshot.PHPVersion,
		snapshot.PluginCount,
		snapshot.ThemeCount,
		snapshot.DBSize,
		lastChecked,
	)
	if err != nil {
		return fmt.Errorf("set snapshot: %w", err)
	}
	return nil
}

// DBSizeBytes returns the database file size in bytes, or 0 if unavailable.
func DBSizeBytes(dbPath string) int64 {
	info, err := os.Stat(dbPath)
	if err != nil {
		return 0
	}
	return info.Size()
}
