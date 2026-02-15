package cache

const createCacheEntries = `CREATE TABLE IF NOT EXISTS cache_entries (
    site_alias  TEXT NOT NULL,
    cache_key   TEXT NOT NULL,
    command     TEXT NOT NULL,
    data        TEXT NOT NULL,
    fetched_at  DATETIME NOT NULL DEFAULT (datetime('now')),
    ttl_seconds INTEGER NOT NULL,
    category    TEXT NOT NULL,
    PRIMARY KEY (site_alias, cache_key)
)`

const createSiteSnapshots = `CREATE TABLE IF NOT EXISTS site_snapshots (
    site_alias    TEXT PRIMARY KEY,
    core_version  TEXT,
    php_version   TEXT,
    plugin_count  INTEGER,
    theme_count   INTEGER,
    db_size       TEXT,
    last_checked  DATETIME
)`

const createIdxCategory = `CREATE INDEX IF NOT EXISTS idx_cache_category ON cache_entries(site_alias, category)`
const createIdxExpiry = `CREATE INDEX IF NOT EXISTS idx_cache_expiry ON cache_entries(fetched_at, ttl_seconds)`

const queryGetEntry = `SELECT site_alias, cache_key, command, data, fetched_at, ttl_seconds, category
    FROM cache_entries
    WHERE site_alias = ? AND cache_key = ?
      AND (julianday('now') - julianday(fetched_at)) * 86400 < ttl_seconds`

const querySetEntry = `INSERT OR REPLACE INTO cache_entries
    (site_alias, cache_key, command, data, fetched_at, ttl_seconds, category)
    VALUES (?, ?, ?, ?, ?, ?, ?)`

const queryInvalidateSite = `DELETE FROM cache_entries WHERE site_alias = ?`

const queryInvalidateAll = `DELETE FROM cache_entries`

const queryPrune = `DELETE FROM cache_entries
    WHERE (julianday('now') - julianday(fetched_at)) * 86400 >= ttl_seconds`

const queryStatsTotal = `SELECT COUNT(*) FROM cache_entries`

const queryStatsExpired = `SELECT COUNT(*) FROM cache_entries
    WHERE (julianday('now') - julianday(fetched_at)) * 86400 >= ttl_seconds`

const queryStatsSites = `SELECT COUNT(DISTINCT site_alias) FROM cache_entries`

const queryStatsCategories = `SELECT category, COUNT(*) FROM cache_entries GROUP BY category`

const queryGetSnapshot = `SELECT site_alias, core_version, php_version, plugin_count, theme_count, db_size, last_checked
    FROM site_snapshots WHERE site_alias = ?`

const querySetSnapshot = `INSERT OR REPLACE INTO site_snapshots
    (site_alias, core_version, php_version, plugin_count, theme_count, db_size, last_checked)
    VALUES (?, ?, ?, ?, ?, ?, ?)`
