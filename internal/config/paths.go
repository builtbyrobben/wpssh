package config

import (
	"os"
	"path/filepath"
)

// Paths holds XDG-compliant directory paths for wpgo.
type Paths struct {
	ConfigDir string // ~/.config/wpgo/
	CacheDir  string // ~/.cache/wpgo/
	DataDir   string // ~/.local/share/wpgo/
}

// DefaultPaths returns XDG-compliant paths, respecting XDG_* environment variables.
func DefaultPaths() Paths {
	return Paths{
		ConfigDir: xdgDir("XDG_CONFIG_HOME", ".config", "wpgo"),
		CacheDir:  xdgDir("XDG_CACHE_HOME", ".cache", "wpgo"),
		DataDir:   xdgDir("XDG_DATA_HOME", filepath.Join(".local", "share"), "wpgo"),
	}
}

// EnsureDirs creates all wpgo directories if they don't exist.
func (p Paths) EnsureDirs() error {
	for _, dir := range []string{p.ConfigDir, p.CacheDir, p.DataDir} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}

// ConfigFile returns the path to config.json.
func (p Paths) ConfigFile() string {
	return filepath.Join(p.ConfigDir, "config.json")
}

// SitesFile returns the path to sites.json (metadata overlay).
func (p Paths) SitesFile() string {
	return filepath.Join(p.ConfigDir, "sites.json")
}

// CacheDB returns the path to the SQLite cache database.
func (p Paths) CacheDB() string {
	return filepath.Join(p.CacheDir, "cache.db")
}

func xdgDir(envVar, fallbackSuffix, appName string) string {
	if dir := os.Getenv(envVar); dir != "" {
		return filepath.Join(dir, appName)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, fallbackSuffix, appName)
}
