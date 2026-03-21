package cmd

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/builtbyrobben/wpssh/internal/batch"
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/safety"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Core commands

type CoreCmd struct {
	Version         CoreVersionCmd         `cmd:"" help:"Show WP version"`
	CheckUpdate     CoreCheckUpdateCmd     `cmd:"" name:"check-update" help:"Check for core updates"`
	Update          CoreUpdateCmd          `cmd:"" help:"Update WordPress core"`
	VerifyChecksums CoreVerifyChecksumsCmd `cmd:"" name:"verify-checksums" help:"Verify core file integrity"`
	IsInstalled     CoreIsInstalledCmd     `cmd:"" name:"is-installed" help:"Check if WP is installed"`
}

type (
	CoreVersionCmd     struct{}
	CoreCheckUpdateCmd struct{}
	CoreUpdateCmd      struct {
		Version string `help:"Version to update to"`
	}
)

type (
	CoreVerifyChecksumsCmd struct{}
	CoreIsInstalledCmd     struct{}
)

func (c *CoreVersionCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	sites, err := rc.ResolveSites()
	if err != nil {
		return err
	}

	if len(sites) == 1 && !g.IsBatchMode() {
		site := sites[0]
		builder := wpcli.New("core", "version")
		cacheKey := builder.CacheKey()

		if cached := rc.CacheGet(site.Alias, cacheKey); cached != "" {
			version := strings.TrimSpace(cached)
			if rc.Formatter.JSON {
				return rc.Formatter.Format(map[string]string{"version": version})
			}
			fmt.Fprintln(rc.Stdout, version)
			return nil
		}

		result, err := rc.ExecWP(context.Background(), site, builder.Build(site.WPPath))
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("wp core version: %s", result.Stderr)
		}

		version := strings.TrimSpace(result.Stdout)
		rc.CacheSet(site.Alias, cacheKey, "core version", version, cache.CategoryCore, rc.CacheTTL(cache.CategoryCore))

		if rc.Formatter.JSON {
			return rc.Formatter.Format(map[string]string{"version": version})
		}
		fmt.Fprintln(rc.Stdout, version)
		return nil
	}

	type coreVersionRow struct {
		Site    string `json:"site"`
		Version string `json:"version"`
	}

	rows := make([]coreVersionRow, 0, len(sites))
	var mu sync.Mutex
	results := batch.NewExecutor().Execute(context.Background(), sites, func(ctx context.Context, site *registry.Site) (string, error) {
		builder := wpcli.New("core", "version")
		cacheKey := builder.CacheKey()
		var version string
		if cached := rc.CacheGet(site.Alias, cacheKey); cached != "" {
			version = strings.TrimSpace(cached)
		} else {
			result, err := rc.ExecWP(ctx, site, builder.Build(site.WPPath))
			if err != nil {
				return "", err
			}
			if result.ExitCode != 0 {
				return "", fmt.Errorf("%s", strings.TrimSpace(result.Stderr))
			}
			version = strings.TrimSpace(result.Stdout)
			rc.CacheSet(site.Alias, cacheKey, "core version", version, cache.CategoryCore, rc.CacheTTL(cache.CategoryCore))
		}

		mu.Lock()
		rows = append(rows, coreVersionRow{Site: site.Alias, Version: version})
		mu.Unlock()
		return version, nil
	}, batchOptions(g, safety.Classify("core", "version"), "core version"))

	if batch.NewReport(results).HasFailures() {
		return writeBatchReport(rc, results)
	}
	return rc.Formatter.Format(rows)
}

func (c *CoreCheckUpdateCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.CoreUpdate](g, "core check-update", cache.CategoryCore, func(*registry.Site) *wpcli.Command {
		return wpcli.New("core", "check-update").Format("json")
	})
}

func (c *CoreUpdateCmd) Run(g *Globals) error {
	return runWPCommand(g, "core update", []string{cache.CategoryCore, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("core", "update")
		if c.Version != "" {
			builder.Flag("version", c.Version)
		}
		return builder
	})
}

func (c *CoreVerifyChecksumsCmd) Run(g *Globals) error {
	return runWPCommand(g, "core verify-checksums", nil, func(*registry.Site) *wpcli.Command {
		return wpcli.New("core", "verify-checksums")
	})
}

func (c *CoreIsInstalledCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	sites, err := rc.ResolveSites()
	if err != nil {
		return err
	}

	if len(sites) == 1 && !g.IsBatchMode() {
		site := sites[0]
		result, err := rc.ExecWP(context.Background(), site, wpcli.New("core", "is-installed").Build(site.WPPath))
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			fmt.Fprintln(rc.Stdout, "WordPress is not installed.")
			return nil
		}
		fmt.Fprintln(rc.Stdout, "WordPress is installed.")
		return nil
	}

	results := batch.NewExecutor().Execute(context.Background(), sites, func(ctx context.Context, site *registry.Site) (string, error) {
		result, err := rc.ExecWP(ctx, site, wpcli.New("core", "is-installed").Build(site.WPPath))
		if err != nil {
			return "", err
		}
		if result.ExitCode != 0 {
			return "WordPress is not installed", nil
		}
		return "WordPress is installed", nil
	}, batchOptions(g, safety.Classify("core", "is-installed"), "core is-installed"))
	return writeBatchReport(rc, results)
}
