package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/builtbyrobben/wpssh/internal/cache"
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

type CoreVersionCmd struct{}
type CoreCheckUpdateCmd struct{}
type CoreUpdateCmd struct{ Version string `help:"Version to update to"` }
type CoreVerifyChecksumsCmd struct{}
type CoreIsInstalledCmd struct{}

func (c *CoreVersionCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("core", "version")
	cacheKey := builder.CacheKey()

	// Check cache first.
	if cached := rc.CacheGet(site.Alias, cacheKey); cached != "" {
		version := strings.TrimSpace(cached)
		if g.JSON {
			return rc.Formatter.Format(map[string]string{"version": version})
		}
		fmt.Fprintln(rc.Stdout, version)
		return nil
	}

	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp core version: %s", result.Stderr)
	}

	version := strings.TrimSpace(result.Stdout)

	// Store in cache.
	rc.CacheSet(site.Alias, cacheKey, "core version", version, cache.CategoryCore, cache.TTLCore)

	if g.JSON {
		return rc.Formatter.Format(map[string]string{"version": version})
	}
	fmt.Fprintln(rc.Stdout, version)
	return nil
}

func (c *CoreCheckUpdateCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("core", "check-update").Format("json")
	cacheKey := builder.CacheKey()

	// Check cache first.
	if cached := rc.CacheGet(site.Alias, cacheKey); cached != "" {
		updates, err := wpcli.ParseJSON[wpcli.CoreUpdate](cached)
		if err == nil {
			return rc.Formatter.Format(updates)
		}
	}

	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp core check-update: %s", result.Stderr)
	}

	updates, err := wpcli.ParseJSON[wpcli.CoreUpdate](result.Stdout)
	if err != nil {
		return err
	}

	// Store in cache.
	if data, err := json.Marshal(updates); err == nil {
		rc.CacheSet(site.Alias, cacheKey, "core check-update", string(data), cache.CategoryCore, cache.TTLCore)
	}

	return rc.Formatter.Format(updates)
}

func (c *CoreUpdateCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("core", "update")
	if c.Version != "" {
		builder.Flag("version", c.Version)
	}
	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp core update: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryCore, cache.CategorySnapshot})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *CoreVerifyChecksumsCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("core", "verify-checksums").Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp core verify-checksums: %s", result.Stderr)
	}
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *CoreIsInstalledCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("core", "is-installed").Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
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
