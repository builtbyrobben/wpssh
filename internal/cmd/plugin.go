package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Plugin commands

type PluginCmd struct {
	List            PluginListCmd            `cmd:"" help:"List all plugins"`
	Install         PluginInstallCmd         `cmd:"" help:"Install a plugin"`
	Activate        PluginActivateCmd        `cmd:"" help:"Activate a plugin"`
	Deactivate      PluginDeactivateCmd      `cmd:"" help:"Deactivate a plugin"`
	Delete          PluginDeleteCmd          `cmd:"" help:"Delete a plugin"`
	Update          PluginUpdateCmd          `cmd:"" help:"Update plugin(s)"`
	Search          PluginSearchCmd          `cmd:"" help:"Search WordPress.org"`
	Get             PluginGetCmd             `cmd:"" help:"Get plugin details"`
	IsActive        PluginIsActiveCmd        `cmd:"" name:"is-active" help:"Check if active"`
	IsInstalled     PluginIsInstalledCmd     `cmd:"" name:"is-installed" help:"Check if installed"`
	Status          PluginStatusCmd          `cmd:"" help:"Show plugin status"`
	VerifyChecksums PluginVerifyChecksumsCmd `cmd:"" name:"verify-checksums" help:"Verify checksums"`
	AutoUpdates     PluginAutoUpdatesCmd     `cmd:"" name:"auto-updates" help:"Manage auto-updates"`
}

type PluginListCmd struct {
	StatusFilter string `help:"Filter by status" name:"status"`
}
type PluginInstallCmd struct {
	Plugin   string `arg:"" help:"Plugin slug"`
	Activate bool   `help:"Activate after install"`
}
type PluginActivateCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginDeactivateCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginDeleteCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginUpdateCmd struct {
	Plugin string `arg:"" optional:"" help:"Plugin slug"`
	All    bool   `help:"Update all plugins"`
}
type PluginSearchCmd struct {
	Term string `arg:"" help:"Search term"`
}
type PluginGetCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginIsActiveCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginIsInstalledCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginStatusCmd struct {
	Plugin string `arg:"" optional:"" help:"Plugin slug"`
}
type PluginVerifyChecksumsCmd struct {
	Plugin string `arg:"" optional:"" help:"Plugin slug"`
}
type PluginAutoUpdatesCmd struct {
	Enable  PluginAutoUpdatesEnableCmd  `cmd:"" help:"Enable auto-updates"`
	Disable PluginAutoUpdatesDisableCmd `cmd:"" help:"Disable auto-updates"`
	Status  PluginAutoUpdatesStatusCmd  `cmd:"" help:"Show auto-update status"`
}
type PluginAutoUpdatesEnableCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginAutoUpdatesDisableCmd struct {
	Plugin string `arg:"" help:"Plugin slug"`
}
type PluginAutoUpdatesStatusCmd struct {
	Plugin string `arg:"" optional:"" help:"Plugin slug"`
}

func (c *PluginListCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("plugin", "list").Format("json")
	if c.StatusFilter != "" {
		builder.Flag("status", c.StatusFilter)
	}
	cacheKey := builder.CacheKey()

	// Check cache first.
	if cached := rc.CacheGet(site.Alias, cacheKey); cached != "" {
		plugins, err := wpcli.ParseJSON[wpcli.Plugin](cached)
		if err == nil {
			return rc.Formatter.Format(plugins)
		}
	}

	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin list: %s", result.Stderr)
	}

	plugins, err := wpcli.ParseJSON[wpcli.Plugin](result.Stdout)
	if err != nil {
		return err
	}

	// Store in cache.
	if data, err := json.Marshal(plugins); err == nil {
		rc.CacheSet(site.Alias, cacheKey, "plugin list", string(data), cache.CategoryPlugins, cache.TTLPlugins)
	}

	return rc.Formatter.Format(plugins)
}

func (c *PluginInstallCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("plugin", "install").Arg(c.Plugin)
	if c.Activate {
		builder.BoolFlag("activate")
	}
	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin install: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryPlugins, cache.CategorySnapshot})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginActivateCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "activate").Arg(c.Plugin).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin activate: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryPlugins, cache.CategorySnapshot})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginDeactivateCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "deactivate").Arg(c.Plugin).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin deactivate: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryPlugins, cache.CategorySnapshot})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginDeleteCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "delete").Arg(c.Plugin).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin delete: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryPlugins, cache.CategorySnapshot})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginUpdateCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("plugin", "update")
	if c.All {
		builder.BoolFlag("all")
	} else if c.Plugin != "" {
		builder.Arg(c.Plugin)
	}
	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin update: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryPlugins, cache.CategorySnapshot})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginSearchCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "search").Arg(c.Term).Format("json").Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin search: %s", result.Stderr)
	}

	plugins, err := wpcli.ParseJSON[wpcli.Plugin](result.Stdout)
	if err != nil {
		return err
	}
	return rc.Formatter.Format(plugins)
}

func (c *PluginGetCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "get").Arg(c.Plugin).Format("json").Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin get: %s", result.Stderr)
	}

	plugin, err := wpcli.ParseSingle[wpcli.Plugin](result.Stdout)
	if err != nil {
		return err
	}
	return rc.Formatter.Format(plugin)
}

func (c *PluginIsActiveCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "is-active").Arg(c.Plugin).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		fmt.Fprintf(rc.Stdout, "Plugin %s is not active.\n", c.Plugin)
	} else {
		fmt.Fprintf(rc.Stdout, "Plugin %s is active.\n", c.Plugin)
	}
	return nil
}

func (c *PluginIsInstalledCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "is-installed").Arg(c.Plugin).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		fmt.Fprintf(rc.Stdout, "Plugin %s is not installed.\n", c.Plugin)
	} else {
		fmt.Fprintf(rc.Stdout, "Plugin %s is installed.\n", c.Plugin)
	}
	return nil
}

func (c *PluginStatusCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("plugin", "status")
	if c.Plugin != "" {
		builder.Arg(c.Plugin)
	}
	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin status: %s", result.Stderr)
	}
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginVerifyChecksumsCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("plugin", "verify-checksums")
	if c.Plugin != "" {
		builder.Arg(c.Plugin)
	} else {
		builder.BoolFlag("all")
	}
	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin verify-checksums: %s", result.Stderr)
	}
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginAutoUpdatesEnableCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "auto-updates", "enable").Arg(c.Plugin).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin auto-updates enable: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryPlugins})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginAutoUpdatesDisableCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("plugin", "auto-updates", "disable").Arg(c.Plugin).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin auto-updates disable: %s", result.Stderr)
	}
	rc.CacheInvalidate(site.Alias, []string{cache.CategoryPlugins})
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}

func (c *PluginAutoUpdatesStatusCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	builder := wpcli.New("plugin", "auto-updates", "status")
	if c.Plugin != "" {
		builder.Arg(c.Plugin)
	}
	cmd := builder.Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp plugin auto-updates status: %s", result.Stderr)
	}
	fmt.Fprintln(rc.Stdout, result.Stdout)
	return nil
}
