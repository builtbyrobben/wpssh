package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Cache provides SQLite-backed caching for wp-cli query results.
type Cache struct {
	db *sql.DB
}

// New opens or creates the cache database at dbPath and runs migrations.
// The parent directory is created with 0700 permissions, and the database
// file is set to 0600 (owner read/write only).
func New(dbPath string) (*Cache, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open cache db: %w", err)
	}

	// SQLite is single-writer — limit to one connection to avoid SQLITE_BUSY
	// errors from database/sql's connection pool creating multiple connections.
	db.SetMaxOpenConns(1)

	// Enable WAL mode for better concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Set busy timeout so concurrent writers wait instead of failing.
	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

	c := &Cache{db: db}
	if err := c.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate cache db: %w", err)
	}

	// Set file permissions to 0600 (owner read/write only).
	if err := os.Chmod(dbPath, 0600); err != nil {
		db.Close()
		return nil, fmt.Errorf("set cache db permissions: %w", err)
	}

	return c, nil
}

// Close closes the database connection.
func (c *Cache) Close() error {
	return c.db.Close()
}

// migrate creates tables and indexes if they don't exist.
func (c *Cache) migrate() error {
	stmts := []string{
		createCacheEntries,
		createSiteSnapshots,
		createIdxCategory,
		createIdxExpiry,
	}
	for _, stmt := range stmts {
		if _, err := c.db.Exec(stmt); err != nil {
			return fmt.Errorf("exec %q: %w", stmt[:40], err)
		}
	}
	return nil
}
