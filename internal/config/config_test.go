package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.DefaultFormat != "table" {
		t.Errorf("default format: got %q, want %q", cfg.DefaultFormat, "table")
	}
	if cfg.DefaultRateLimit.Delay != 500*time.Millisecond {
		t.Errorf("default rate limit delay: got %v", cfg.DefaultRateLimit.Delay)
	}
	if cfg.DefaultRateLimit.MaxConns != 3 {
		t.Errorf("default rate limit max_conns: got %d", cfg.DefaultRateLimit.MaxConns)
	}

	primeVPS, ok := cfg.RateLimits["192.0.2.10"]
	if !ok {
		t.Fatal("missing Prime VPS rate limit")
	}
	if primeVPS.Delay != 3*time.Second {
		t.Errorf("Prime VPS delay: got %v, want 3s", primeVPS.Delay)
	}
	if primeVPS.MaxConns != 1 {
		t.Errorf("Prime VPS max_conns: got %d, want 1", primeVPS.MaxConns)
	}

	// Cache TTLs.
	if cfg.CacheTTLs.Plugins != 1*time.Hour {
		t.Errorf("plugins TTL: got %v", cfg.CacheTTLs.Plugins)
	}
	if cfg.CacheTTLs.Themes != 1*time.Hour {
		t.Errorf("themes TTL: got %v", cfg.CacheTTLs.Themes)
	}
	if cfg.CacheTTLs.Core != 24*time.Hour {
		t.Errorf("core TTL: got %v", cfg.CacheTTLs.Core)
	}
	if cfg.CacheTTLs.Users != 6*time.Hour {
		t.Errorf("users TTL: got %v", cfg.CacheTTLs.Users)
	}
	if cfg.CacheTTLs.Options != 30*time.Minute {
		t.Errorf("options TTL: got %v", cfg.CacheTTLs.Options)
	}
	if cfg.CacheTTLs.Snapshot != 4*time.Hour {
		t.Errorf("snapshot TTL: got %v", cfg.CacheTTLs.Snapshot)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.json")
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	// Should return defaults.
	if cfg.DefaultFormat != "table" {
		t.Errorf("expected default format 'table', got %q", cfg.DefaultFormat)
	}
}

func TestLoad_FromFile(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	content := `{
  "default_format": "json",
  "default_site": "sitealpha",
  "groups": {
    "staging": {"aliases": ["stage1", "stage2"]}
  }
}`
	if err := os.WriteFile(cfgPath, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.DefaultFormat != "json" {
		t.Errorf("default_format: got %q, want %q", cfg.DefaultFormat, "json")
	}
	if cfg.DefaultSite != "sitealpha" {
		t.Errorf("default_site: got %q", cfg.DefaultSite)
	}

	// Merged groups.
	g, ok := cfg.Groups["staging"]
	if !ok {
		t.Fatal("missing staging group")
	}
	if len(g.Aliases) != 2 {
		t.Errorf("staging aliases: got %d", len(g.Aliases))
	}

	// Default rate limit should be merged in.
	if cfg.DefaultRateLimit.MaxConns != 3 {
		t.Errorf("default rate limit not merged: got %d", cfg.DefaultRateLimit.MaxConns)
	}
}

func TestSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	original := DefaultConfig()
	original.DefaultSite = "testsite"
	original.Groups["mygroup"] = GroupConfig{Aliases: []string{"a", "b"}}

	if err := Save(&original, cfgPath); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.DefaultSite != "testsite" {
		t.Errorf("default_site: got %q", loaded.DefaultSite)
	}
	if len(loaded.Groups["mygroup"].Aliases) != 2 {
		t.Errorf("group aliases: got %d", len(loaded.Groups["mygroup"].Aliases))
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.json")

	cfg := DefaultConfig()
	if err := Save(&cfg, cfgPath); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Tmp file should not remain.
	if _, err := os.Stat(cfgPath + ".tmp"); !os.IsNotExist(err) {
		t.Error("temp file should not exist after save")
	}

	// Config file should exist.
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file should exist: %v", err)
	}
}

func TestUserGroups(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Groups["staging"] = GroupConfig{Aliases: []string{"s1", "s2"}}
	cfg.Groups["production"] = GroupConfig{Aliases: []string{"p1"}}

	groups := cfg.UserGroups()
	if len(groups) != 2 {
		t.Errorf("expected 2 groups, got %d", len(groups))
	}
	if len(groups["staging"]) != 2 {
		t.Errorf("staging: got %d aliases", len(groups["staging"]))
	}
	if len(groups["production"]) != 1 {
		t.Errorf("production: got %d aliases", len(groups["production"]))
	}
}

func TestDefaultPaths(t *testing.T) {
	paths := DefaultPaths()

	if paths.ConfigDir == "" {
		t.Error("ConfigDir should not be empty")
	}
	if paths.CacheDir == "" {
		t.Error("CacheDir should not be empty")
	}
	if paths.DataDir == "" {
		t.Error("DataDir should not be empty")
	}

	if paths.ConfigFile() == "" {
		t.Error("ConfigFile should not be empty")
	}
	if paths.SitesFile() == "" {
		t.Error("SitesFile should not be empty")
	}
	if paths.CacheDB() == "" {
		t.Error("CacheDB should not be empty")
	}
}

func TestDefaultPaths_XDGOverride(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	paths := DefaultPaths()
	expected := filepath.Join(tmpDir, "wpgo")
	if paths.ConfigDir != expected {
		t.Errorf("ConfigDir: got %q, want %q", paths.ConfigDir, expected)
	}
}

func TestEnsureDirs(t *testing.T) {
	tmpDir := t.TempDir()
	paths := Paths{
		ConfigDir: filepath.Join(tmpDir, "config", "wpgo"),
		CacheDir:  filepath.Join(tmpDir, "cache", "wpgo"),
		DataDir:   filepath.Join(tmpDir, "data", "wpgo"),
	}

	if err := paths.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	for _, dir := range []string{paths.ConfigDir, paths.CacheDir, paths.DataDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("dir %s should exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s should be a directory", dir)
		}
	}
}

func TestMergeConfig(t *testing.T) {
	defaults := DefaultConfig()

	// User config with partial overrides.
	user := Config{
		DefaultFormat: "json",
		// Leave other fields empty — should be filled from defaults.
	}

	merged := mergeConfig(defaults, user)

	if merged.DefaultFormat != "json" {
		t.Errorf("format should be user's 'json', got %q", merged.DefaultFormat)
	}
	if merged.DefaultRateLimit.MaxConns != defaults.DefaultRateLimit.MaxConns {
		t.Errorf("default rate limit should be from defaults, got %d", merged.DefaultRateLimit.MaxConns)
	}
	if merged.CacheTTLs.Plugins != defaults.CacheTTLs.Plugins {
		t.Errorf("cache TTLs should be from defaults")
	}
}
