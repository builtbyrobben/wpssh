package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func statFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test-cache.db")
	c, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	t.Cleanup(func() { c.Close() })
	return c
}

func TestCacheSetGetRoundTrip(t *testing.T) {
	c := newTestCache(t)

	entry := &CacheEntry{
		SiteAlias:  "sitealpha",
		CacheKey:   MakeCacheKey("sitealpha", "plugin list:--format=json"),
		Command:    "plugin list --format=json",
		Data:       `[{"name":"akismet","status":"active","version":"5.3"}]`,
		FetchedAt:  time.Now().UTC().Truncate(time.Second),
		TTLSeconds: 3600,
		Category:   "plugins",
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := c.Get(entry.SiteAlias, entry.CacheKey)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil, expected entry")
	}
	if got.SiteAlias != entry.SiteAlias {
		t.Errorf("SiteAlias = %q, want %q", got.SiteAlias, entry.SiteAlias)
	}
	if got.CacheKey != entry.CacheKey {
		t.Errorf("CacheKey = %q, want %q", got.CacheKey, entry.CacheKey)
	}
	if got.Command != entry.Command {
		t.Errorf("Command = %q, want %q", got.Command, entry.Command)
	}
	if got.Data != entry.Data {
		t.Errorf("Data = %q, want %q", got.Data, entry.Data)
	}
	if got.Category != entry.Category {
		t.Errorf("Category = %q, want %q", got.Category, entry.Category)
	}
	if got.TTLSeconds != entry.TTLSeconds {
		t.Errorf("TTLSeconds = %d, want %d", got.TTLSeconds, entry.TTLSeconds)
	}
}

func TestCacheGetNotFound(t *testing.T) {
	c := newTestCache(t)

	got, err := c.Get("nonexistent", "somekey")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != nil {
		t.Fatalf("Get() = %+v, want nil", got)
	}
}

func TestCacheTTLExpiry(t *testing.T) {
	c := newTestCache(t)

	entry := &CacheEntry{
		SiteAlias:  "sitealpha",
		CacheKey:   "expired-key",
		Command:    "plugin list",
		Data:       `[]`,
		FetchedAt:  time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second),
		TTLSeconds: 3600, // 1 hour TTL, but fetched 2 hours ago
		Category:   "plugins",
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := c.Get(entry.SiteAlias, entry.CacheKey)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != nil {
		t.Fatal("Get() returned entry for expired cache, expected nil")
	}
}

func TestCacheTTLNotExpired(t *testing.T) {
	c := newTestCache(t)

	entry := &CacheEntry{
		SiteAlias:  "sitealpha",
		CacheKey:   "fresh-key",
		Command:    "plugin list",
		Data:       `[]`,
		FetchedAt:  time.Now().UTC().Add(-30 * time.Minute).Truncate(time.Second),
		TTLSeconds: 3600, // 1 hour TTL, fetched 30 min ago
		Category:   "plugins",
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := c.Get(entry.SiteAlias, entry.CacheKey)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil for non-expired entry")
	}
}

func TestInvalidateByCategory(t *testing.T) {
	c := newTestCache(t)

	entries := []*CacheEntry{
		{SiteAlias: "site1", CacheKey: "k1", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
		{SiteAlias: "site1", CacheKey: "k2", Command: "theme list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "themes"},
		{SiteAlias: "site1", CacheKey: "k3", Command: "core version", Data: `"6.4"`, FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 86400, Category: "core"},
	}
	for _, e := range entries {
		if err := c.Set(e); err != nil {
			t.Fatalf("Set() error: %v", err)
		}
	}

	// Invalidate only plugins
	if err := c.Invalidate("site1", []string{"plugins"}); err != nil {
		t.Fatalf("Invalidate() error: %v", err)
	}

	// plugins should be gone
	got, err := c.Get("site1", "k1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got != nil {
		t.Error("plugins entry should have been invalidated")
	}

	// themes should remain
	got, err = c.Get("site1", "k2")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Error("themes entry should still exist")
	}

	// core should remain
	got, err = c.Get("site1", "k3")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Error("core entry should still exist")
	}
}

func TestInvalidateMultipleCategories(t *testing.T) {
	c := newTestCache(t)

	entries := []*CacheEntry{
		{SiteAlias: "site1", CacheKey: "k1", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
		{SiteAlias: "site1", CacheKey: "k2", Command: "theme list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "themes"},
		{SiteAlias: "site1", CacheKey: "k3", Command: "core version", Data: `"6.4"`, FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 86400, Category: "core"},
	}
	for _, e := range entries {
		if err := c.Set(e); err != nil {
			t.Fatalf("Set() error: %v", err)
		}
	}

	// Invalidate plugins + snapshot (simulating plugin update)
	if err := c.Invalidate("site1", []string{"plugins", "themes"}); err != nil {
		t.Fatalf("Invalidate() error: %v", err)
	}

	got, _ := c.Get("site1", "k1")
	if got != nil {
		t.Error("plugins should be invalidated")
	}
	got, _ = c.Get("site1", "k2")
	if got != nil {
		t.Error("themes should be invalidated")
	}
	got, _ = c.Get("site1", "k3")
	if got == nil {
		t.Error("core should still exist")
	}
}

func TestInvalidateAllCategoriesForSite(t *testing.T) {
	c := newTestCache(t)

	entries := []*CacheEntry{
		{SiteAlias: "site1", CacheKey: "k1", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
		{SiteAlias: "site1", CacheKey: "k2", Command: "theme list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "themes"},
		{SiteAlias: "site2", CacheKey: "k3", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
	}
	for _, e := range entries {
		if err := c.Set(e); err != nil {
			t.Fatalf("Set() error: %v", err)
		}
	}

	// Invalidate ALL for site1 (nil categories)
	if err := c.Invalidate("site1", nil); err != nil {
		t.Fatalf("Invalidate() error: %v", err)
	}

	got, _ := c.Get("site1", "k1")
	if got != nil {
		t.Error("site1 plugins should be invalidated")
	}
	got, _ = c.Get("site1", "k2")
	if got != nil {
		t.Error("site1 themes should be invalidated")
	}
	// site2 should be untouched
	got, _ = c.Get("site2", "k3")
	if got == nil {
		t.Error("site2 plugins should still exist")
	}
}

func TestInvalidateAll(t *testing.T) {
	c := newTestCache(t)

	entries := []*CacheEntry{
		{SiteAlias: "site1", CacheKey: "k1", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
		{SiteAlias: "site2", CacheKey: "k2", Command: "theme list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "themes"},
	}
	for _, e := range entries {
		if err := c.Set(e); err != nil {
			t.Fatalf("Set() error: %v", err)
		}
	}

	if err := c.InvalidateAll(); err != nil {
		t.Fatalf("InvalidateAll() error: %v", err)
	}

	got, _ := c.Get("site1", "k1")
	if got != nil {
		t.Error("site1 should be invalidated")
	}
	got, _ = c.Get("site2", "k2")
	if got != nil {
		t.Error("site2 should be invalidated")
	}
}

func TestMakeCacheKeyDeterministic(t *testing.T) {
	key1 := MakeCacheKey("sitealpha", "plugin list:--format=json --status=active")
	key2 := MakeCacheKey("sitealpha", "plugin list:--format=json --status=active")
	if key1 != key2 {
		t.Errorf("same input produced different keys: %q vs %q", key1, key2)
	}

	// Different site alias => different key
	key3 := MakeCacheKey("othersite", "plugin list:--format=json --status=active")
	if key1 == key3 {
		t.Error("different site aliases should produce different keys")
	}

	// Different command => different key
	key4 := MakeCacheKey("sitealpha", "plugin list:--format=json")
	if key1 == key4 {
		t.Error("different commands should produce different keys")
	}

	// SHA256 produces 64-char hex string
	if len(key1) != 64 {
		t.Errorf("cache key length = %d, want 64", len(key1))
	}
}

func TestPrune(t *testing.T) {
	c := newTestCache(t)

	// One expired, one fresh
	entries := []*CacheEntry{
		{SiteAlias: "site1", CacheKey: "expired", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
		{SiteAlias: "site1", CacheKey: "fresh", Command: "theme list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "themes"},
	}
	for _, e := range entries {
		if err := c.Set(e); err != nil {
			t.Fatalf("Set() error: %v", err)
		}
	}

	pruned, err := c.Prune()
	if err != nil {
		t.Fatalf("Prune() error: %v", err)
	}
	if pruned != 1 {
		t.Errorf("Prune() = %d, want 1", pruned)
	}

	// Expired should be gone (can't use Get since it filters expired anyway)
	// Verify with stats
	stats, err := c.Stats()
	if err != nil {
		t.Fatalf("Stats() error: %v", err)
	}
	if stats.TotalEntries != 1 {
		t.Errorf("TotalEntries = %d, want 1 after prune", stats.TotalEntries)
	}
}

func TestSnapshotSetGet(t *testing.T) {
	c := newTestCache(t)

	snap := &SiteSnapshot{
		SiteAlias:   "sitealpha",
		CoreVersion: "6.4.2",
		PHPVersion:  "8.2.14",
		PluginCount: 12,
		ThemeCount:  3,
		DBSize:      "45 MB",
		LastChecked: time.Now().UTC().Truncate(time.Second),
	}

	if err := c.SetSnapshot(snap); err != nil {
		t.Fatalf("SetSnapshot() error: %v", err)
	}

	got, err := c.GetSnapshot("sitealpha")
	if err != nil {
		t.Fatalf("GetSnapshot() error: %v", err)
	}
	if got == nil {
		t.Fatal("GetSnapshot() returned nil")
	}
	if got.CoreVersion != snap.CoreVersion {
		t.Errorf("CoreVersion = %q, want %q", got.CoreVersion, snap.CoreVersion)
	}
	if got.PHPVersion != snap.PHPVersion {
		t.Errorf("PHPVersion = %q, want %q", got.PHPVersion, snap.PHPVersion)
	}
	if got.PluginCount != snap.PluginCount {
		t.Errorf("PluginCount = %d, want %d", got.PluginCount, snap.PluginCount)
	}
	if got.ThemeCount != snap.ThemeCount {
		t.Errorf("ThemeCount = %d, want %d", got.ThemeCount, snap.ThemeCount)
	}
	if got.DBSize != snap.DBSize {
		t.Errorf("DBSize = %q, want %q", got.DBSize, snap.DBSize)
	}
	if !got.LastChecked.Equal(snap.LastChecked) {
		t.Errorf("LastChecked = %v, want %v", got.LastChecked, snap.LastChecked)
	}
}

func TestSnapshotNotFound(t *testing.T) {
	c := newTestCache(t)

	got, err := c.GetSnapshot("nonexistent")
	if err != nil {
		t.Fatalf("GetSnapshot() error: %v", err)
	}
	if got != nil {
		t.Fatalf("GetSnapshot() = %+v, want nil", got)
	}
}

func TestSnapshotUpsert(t *testing.T) {
	c := newTestCache(t)

	snap1 := &SiteSnapshot{
		SiteAlias:   "sitealpha",
		CoreVersion: "6.4.1",
		PHPVersion:  "8.2.14",
		PluginCount: 10,
		ThemeCount:  2,
		DBSize:      "40 MB",
		LastChecked: time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second),
	}
	if err := c.SetSnapshot(snap1); err != nil {
		t.Fatalf("SetSnapshot() error: %v", err)
	}

	snap2 := &SiteSnapshot{
		SiteAlias:   "sitealpha",
		CoreVersion: "6.4.2",
		PHPVersion:  "8.2.14",
		PluginCount: 12,
		ThemeCount:  3,
		DBSize:      "45 MB",
		LastChecked: time.Now().UTC().Truncate(time.Second),
	}
	if err := c.SetSnapshot(snap2); err != nil {
		t.Fatalf("SetSnapshot() error: %v", err)
	}

	got, err := c.GetSnapshot("sitealpha")
	if err != nil {
		t.Fatalf("GetSnapshot() error: %v", err)
	}
	if got.CoreVersion != "6.4.2" {
		t.Errorf("CoreVersion = %q, want %q (should be updated)", got.CoreVersion, "6.4.2")
	}
	if got.PluginCount != 12 {
		t.Errorf("PluginCount = %d, want 12 (should be updated)", got.PluginCount)
	}
}

func TestStatsReporting(t *testing.T) {
	c := newTestCache(t)

	entries := []*CacheEntry{
		{SiteAlias: "site1", CacheKey: "k1", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
		{SiteAlias: "site1", CacheKey: "k2", Command: "theme list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "themes"},
		{SiteAlias: "site2", CacheKey: "k3", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: "plugins"},
		// One expired entry
		{SiteAlias: "site2", CacheKey: "k4", Command: "core version", Data: `"6.4"`, FetchedAt: time.Now().UTC().Add(-25 * time.Hour).Truncate(time.Second), TTLSeconds: 86400, Category: "core"},
	}
	for _, e := range entries {
		if err := c.Set(e); err != nil {
			t.Fatalf("Set() error: %v", err)
		}
	}

	stats, err := c.Stats()
	if err != nil {
		t.Fatalf("Stats() error: %v", err)
	}

	if stats.TotalEntries != 4 {
		t.Errorf("TotalEntries = %d, want 4", stats.TotalEntries)
	}
	if stats.ExpiredEntries != 1 {
		t.Errorf("ExpiredEntries = %d, want 1", stats.ExpiredEntries)
	}
	if stats.SiteCount != 2 {
		t.Errorf("SiteCount = %d, want 2", stats.SiteCount)
	}
	if stats.Categories["plugins"] != 2 {
		t.Errorf("Categories[plugins] = %d, want 2", stats.Categories["plugins"])
	}
	if stats.Categories["themes"] != 1 {
		t.Errorf("Categories[themes] = %d, want 1", stats.Categories["themes"])
	}
	if stats.Categories["core"] != 1 {
		t.Errorf("Categories[core] = %d, want 1", stats.Categories["core"])
	}
	if stats.DBSizeBytes <= 0 {
		t.Errorf("DBSizeBytes = %d, want > 0", stats.DBSizeBytes)
	}
}

func TestSetOverwritesExisting(t *testing.T) {
	c := newTestCache(t)

	entry := &CacheEntry{
		SiteAlias:  "site1",
		CacheKey:   "k1",
		Command:    "plugin list",
		Data:       `[{"name":"old"}]`,
		FetchedAt:  time.Now().UTC().Truncate(time.Second),
		TTLSeconds: 3600,
		Category:   "plugins",
	}
	if err := c.Set(entry); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	// Overwrite with new data
	entry.Data = `[{"name":"new"}]`
	if err := c.Set(entry); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := c.Get("site1", "k1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got.Data != `[{"name":"new"}]` {
		t.Errorf("Data = %q, want updated value", got.Data)
	}
}

func TestFilePermissions(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "perms-test.db")
	c, err := New(dbPath)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	defer c.Close()

	info, err := filepath.EvalSymlinks(dbPath)
	if err != nil {
		t.Fatalf("EvalSymlinks() error: %v", err)
	}
	_ = info
	// Check the file mode using os.Stat directly on the path
	stat, err := statFile(dbPath)
	if err != nil {
		t.Fatalf("stat error: %v", err)
	}
	perm := stat.Mode().Perm()
	if perm != 0o600 {
		t.Errorf("file permissions = %o, want 0600", perm)
	}
}

func TestPruneNothingToDelete(t *testing.T) {
	c := newTestCache(t)

	pruned, err := c.Prune()
	if err != nil {
		t.Fatalf("Prune() error: %v", err)
	}
	if pruned != 0 {
		t.Errorf("Prune() = %d, want 0 (empty db)", pruned)
	}
}

func TestEmptyStats(t *testing.T) {
	c := newTestCache(t)

	stats, err := c.Stats()
	if err != nil {
		t.Fatalf("Stats() error: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("TotalEntries = %d, want 0", stats.TotalEntries)
	}
	if stats.ExpiredEntries != 0 {
		t.Errorf("ExpiredEntries = %d, want 0", stats.ExpiredEntries)
	}
	if stats.SiteCount != 0 {
		t.Errorf("SiteCount = %d, want 0", stats.SiteCount)
	}
	if len(stats.Categories) != 0 {
		t.Errorf("Categories = %v, want empty", stats.Categories)
	}
}

func TestQueryKeyedSeparation(t *testing.T) {
	c := newTestCache(t)

	// Two different queries for the same command with different flags
	// must get separate cache entries.
	key1 := MakeCacheKey("sitealpha", "plugin list:--format=json")
	key2 := MakeCacheKey("sitealpha", "plugin list:--format=json --status=active")

	if key1 == key2 {
		t.Fatal("different flags should produce different cache keys")
	}

	entry1 := &CacheEntry{
		SiteAlias:  "sitealpha",
		CacheKey:   key1,
		Command:    "plugin list --format=json",
		Data:       `[{"name":"akismet"},{"name":"jetpack"}]`,
		FetchedAt:  time.Now().UTC().Truncate(time.Second),
		TTLSeconds: TTLPlugins,
		Category:   CategoryPlugins,
	}
	entry2 := &CacheEntry{
		SiteAlias:  "sitealpha",
		CacheKey:   key2,
		Command:    "plugin list --format=json --status=active",
		Data:       `[{"name":"akismet"}]`,
		FetchedAt:  time.Now().UTC().Truncate(time.Second),
		TTLSeconds: TTLPlugins,
		Category:   CategoryPlugins,
	}

	if err := c.Set(entry1); err != nil {
		t.Fatalf("Set(entry1) error: %v", err)
	}
	if err := c.Set(entry2); err != nil {
		t.Fatalf("Set(entry2) error: %v", err)
	}

	got1, err := c.Get("sitealpha", key1)
	if err != nil {
		t.Fatalf("Get(key1) error: %v", err)
	}
	got2, err := c.Get("sitealpha", key2)
	if err != nil {
		t.Fatalf("Get(key2) error: %v", err)
	}

	if got1 == nil || got2 == nil {
		t.Fatal("both entries should exist")
	}
	if got1.Data == got2.Data {
		t.Error("entries with different keys should have different data")
	}
	if got1.Data != entry1.Data {
		t.Errorf("entry1 data mismatch: got %q, want %q", got1.Data, entry1.Data)
	}
	if got2.Data != entry2.Data {
		t.Errorf("entry2 data mismatch: got %q, want %q", got2.Data, entry2.Data)
	}
}

func TestOptionsMetadataOnly(t *testing.T) {
	// The options cache must NEVER store option values -- only metadata
	// (option_name + autoload). This test verifies the pattern callers
	// must follow and that values are never present in cached data.

	c := newTestCache(t)

	// Simulate what wpgo should cache: only name + autoload, never values.
	type OptionMeta struct {
		Name     string `json:"option_name"`
		Autoload string `json:"autoload"`
	}

	metaOnly := []OptionMeta{
		{Name: "siteurl", Autoload: "yes"},
		{Name: "blogname", Autoload: "yes"},
		{Name: "secret_key", Autoload: "no"},
	}
	data, err := json.Marshal(metaOnly)
	if err != nil {
		t.Fatalf("json.Marshal error: %v", err)
	}

	key := MakeCacheKey("sitealpha", "option list:--format=json")
	entry := &CacheEntry{
		SiteAlias:  "sitealpha",
		CacheKey:   key,
		Command:    "option list --format=json",
		Data:       string(data),
		FetchedAt:  time.Now().UTC().Truncate(time.Second),
		TTLSeconds: TTLOptions,
		Category:   CategoryOptions,
	}

	if err := c.Set(entry); err != nil {
		t.Fatalf("Set() error: %v", err)
	}

	got, err := c.Get("sitealpha", key)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if got == nil {
		t.Fatal("Get() returned nil")
	}

	// Verify the cached data contains ONLY name + autoload fields.
	var cached []map[string]any
	if err := json.Unmarshal([]byte(got.Data), &cached); err != nil {
		t.Fatalf("json.Unmarshal error: %v", err)
	}

	for i, opt := range cached {
		// Must have option_name and autoload
		if _, ok := opt["option_name"]; !ok {
			t.Errorf("option[%d] missing option_name", i)
		}
		if _, ok := opt["autoload"]; !ok {
			t.Errorf("option[%d] missing autoload", i)
		}
		// Must NOT have option_value or value
		for key := range opt {
			lower := strings.ToLower(key)
			if strings.Contains(lower, "value") {
				t.Errorf("option[%d] contains value field %q -- options cache must NEVER store values", i, key)
			}
		}
		// Must have exactly 2 fields
		if len(opt) != 2 {
			t.Errorf("option[%d] has %d fields, want exactly 2 (option_name, autoload)", i, len(opt))
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := newTestCache(t)

	const goroutines = 20
	const opsPerGoroutine = 50

	var wg sync.WaitGroup
	errs := make(chan error, goroutines*opsPerGoroutine)

	// Concurrent writers
	for g := 0; g < goroutines/2; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				entry := &CacheEntry{
					SiteAlias:  fmt.Sprintf("site%d", id),
					CacheKey:   fmt.Sprintf("key-%d-%d", id, i),
					Command:    fmt.Sprintf("plugin list #%d", i),
					Data:       fmt.Sprintf(`[{"id":%d}]`, i),
					FetchedAt:  time.Now().UTC().Truncate(time.Second),
					TTLSeconds: 3600,
					Category:   CategoryPlugins,
				}
				if err := c.Set(entry); err != nil {
					errs <- fmt.Errorf("Set goroutine=%d i=%d: %w", id, i, err)
				}
			}
		}(g)
	}

	// Concurrent readers
	for g := goroutines / 2; g < goroutines; g++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; i < opsPerGoroutine; i++ {
				_, err := c.Get(fmt.Sprintf("site%d", id%5), fmt.Sprintf("key-%d-%d", id%5, i))
				if err != nil {
					errs <- fmt.Errorf("Get goroutine=%d i=%d: %w", id, i, err)
				}
			}
		}(g)
	}

	// Concurrent stats + prune
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < opsPerGoroutine; i++ {
			if _, err := c.Stats(); err != nil {
				errs <- fmt.Errorf("Stats i=%d: %w", i, err)
			}
			if _, err := c.Prune(); err != nil {
				errs <- fmt.Errorf("Prune i=%d: %w", i, err)
			}
		}
	}()

	wg.Wait()
	close(errs)

	var errList []error
	for err := range errs {
		errList = append(errList, err)
	}
	if len(errList) > 0 {
		for _, e := range errList[:min(5, len(errList))] {
			t.Error(e)
		}
		if len(errList) > 5 {
			t.Errorf("... and %d more errors", len(errList)-5)
		}
	}
}

func TestInvalidateCategoryConvenience(t *testing.T) {
	c := newTestCache(t)

	if err := c.Set(&CacheEntry{
		SiteAlias: "site1", CacheKey: "k1", Command: "plugin list",
		Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second),
		TTLSeconds: 3600, Category: CategoryPlugins,
	}); err != nil {
		t.Fatal(err)
	}
	if err := c.Set(&CacheEntry{
		SiteAlias: "site1", CacheKey: "k2", Command: "theme list",
		Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second),
		TTLSeconds: 3600, Category: CategoryThemes,
	}); err != nil {
		t.Fatal(err)
	}

	if err := c.InvalidateCategory("site1", CategoryPlugins); err != nil {
		t.Fatalf("InvalidateCategory error: %v", err)
	}

	got, _ := c.Get("site1", "k1")
	if got != nil {
		t.Error("plugins should be invalidated")
	}
	got, _ = c.Get("site1", "k2")
	if got == nil {
		t.Error("themes should still exist")
	}
}

func TestInvalidateSiteConvenience(t *testing.T) {
	c := newTestCache(t)

	for _, e := range []*CacheEntry{
		{SiteAlias: "site1", CacheKey: "k1", Command: "plugin list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: CategoryPlugins},
		{SiteAlias: "site1", CacheKey: "k2", Command: "theme list", Data: "[]", FetchedAt: time.Now().UTC().Truncate(time.Second), TTLSeconds: 3600, Category: CategoryThemes},
	} {
		if err := c.Set(e); err != nil {
			t.Fatal(err)
		}
	}

	if err := c.InvalidateSite("site1"); err != nil {
		t.Fatalf("InvalidateSite error: %v", err)
	}

	got, _ := c.Get("site1", "k1")
	if got != nil {
		t.Error("site1 plugins should be invalidated")
	}
	got, _ = c.Get("site1", "k2")
	if got != nil {
		t.Error("site1 themes should be invalidated")
	}
}

func TestDefaultTTLConstants(t *testing.T) {
	// Verify TTL constants match the plan specification.
	if TTLPlugins != 3600 {
		t.Errorf("TTLPlugins = %d, want 3600", TTLPlugins)
	}
	if TTLThemes != 3600 {
		t.Errorf("TTLThemes = %d, want 3600", TTLThemes)
	}
	if TTLCore != 86400 {
		t.Errorf("TTLCore = %d, want 86400", TTLCore)
	}
	if TTLUsers != 21600 {
		t.Errorf("TTLUsers = %d, want 21600", TTLUsers)
	}
	if TTLOptions != 1800 {
		t.Errorf("TTLOptions = %d, want 1800", TTLOptions)
	}
	if TTLSnapshot != 14400 {
		t.Errorf("TTLSnapshot = %d, want 14400", TTLSnapshot)
	}
}
