package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// DB commands

type DBCmd struct {
	Export   DBExportCmd   `cmd:"" help:"Export database"`
	Import   DBImportCmd   `cmd:"" help:"Import database"`
	Query    DBQueryCmd    `cmd:"" help:"Run SQL query"`
	Search   DBSearchCmd   `cmd:"" help:"Search database"`
	Size     DBSizeCmd     `cmd:"" help:"Show database size"`
	Tables   DBTablesCmd   `cmd:"" help:"List tables"`
	Prefix   DBPrefixCmd   `cmd:"" help:"Show table prefix"`
	Check    DBCheckCmd    `cmd:"" help:"Check database"`
	Repair   DBRepairCmd   `cmd:"" help:"Repair database"`
	Optimize DBOptimizeCmd `cmd:"" help:"Optimize database"`
	Reset    DBResetCmd    `cmd:"" help:"Reset database"`
}

type DBExportCmd struct {
	File string `arg:"" optional:"" help:"Export file path"`
}
type DBImportCmd struct {
	File string `arg:"" help:"Import file path"`
}
type DBQueryCmd struct {
	SQL string `arg:"" help:"SQL query"`
}
type DBSearchCmd struct {
	Term string `arg:"" help:"Search string"`
}
type (
	DBSizeCmd     struct{}
	DBTablesCmd   struct{}
	DBPrefixCmd   struct{}
	DBCheckCmd    struct{}
	DBRepairCmd   struct{}
	DBOptimizeCmd struct{}
	DBResetCmd    struct{}
)

func (c *DBExportCmd) Run(g *Globals) error {
	return runWPCommand(g, "db export", nil, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("db", "export")
		if c.File != "" {
			builder.Arg(c.File)
		}
		return builder
	})
}

func (c *DBImportCmd) Run(g *Globals) error {
	return runWPCommand(g, "db import", []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("db", "import").Arg(c.File)
	})
}

func (c *DBQueryCmd) Run(g *Globals) error {
	return runWPCommand(g, "db query", nil, func(*registry.Site) *wpcli.Command {
		return wpcli.New("db", "query").Arg(c.SQL)
	})
}

func (c *DBSearchCmd) Run(g *Globals) error {
	return execPassthrough(g, "db", "search", c.Term)
}

func (c *DBSizeCmd) Run(g *Globals) error {
	return execPassthrough(g, "db", "size")
}

func (c *DBTablesCmd) Run(g *Globals) error {
	return execPassthrough(g, "db", "tables")
}

func (c *DBPrefixCmd) Run(g *Globals) error {
	return execPassthrough(g, "db", "prefix")
}

func (c *DBCheckCmd) Run(g *Globals) error {
	return execPassthrough(g, "db", "check")
}

func (c *DBRepairCmd) Run(g *Globals) error {
	return runWPCommand(g, "db repair", []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("db", "repair")
	})
}

func (c *DBOptimizeCmd) Run(g *Globals) error {
	return runWPCommand(g, "db optimize", []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("db", "optimize")
	})
}

func (c *DBResetCmd) Run(g *Globals) error {
	return runWPCommand(g, "db reset", []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("db", "reset")
	})
}
