package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
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
	return runStructuredListCommand[wpcli.Plugin](g, "plugin list", cache.CategoryPlugins, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("plugin", "list").Format("json")
		if c.StatusFilter != "" {
			builder.Flag("status", c.StatusFilter)
		}
		return builder
	})
}

func (c *PluginInstallCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin install", []string{cache.CategoryPlugins, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("plugin", "install").Arg(c.Plugin)
		if c.Activate {
			builder.BoolFlag("activate")
		}
		return builder
	})
}

func (c *PluginActivateCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin activate", []string{cache.CategoryPlugins, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("plugin", "activate").Arg(c.Plugin)
	})
}

func (c *PluginDeactivateCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin deactivate", []string{cache.CategoryPlugins, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("plugin", "deactivate").Arg(c.Plugin)
	})
}

func (c *PluginDeleteCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin delete", []string{cache.CategoryPlugins, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("plugin", "delete").Arg(c.Plugin)
	})
}

func (c *PluginUpdateCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin update", []string{cache.CategoryPlugins, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("plugin", "update")
		if c.All {
			builder.BoolFlag("all")
		} else if c.Plugin != "" {
			builder.Arg(c.Plugin)
		}
		return builder
	})
}

func (c *PluginSearchCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.Plugin](g, "plugin search", "", func(*registry.Site) *wpcli.Command {
		return wpcli.New("plugin", "search").Arg(c.Term).Format("json")
	})
}

func (c *PluginGetCmd) Run(g *Globals) error {
	return runStructuredSingleCommand[wpcli.Plugin](g, "plugin get", func(*registry.Site) *wpcli.Command {
		return wpcli.New("plugin", "get").Arg(c.Plugin).Format("json")
	})
}

func (c *PluginIsActiveCmd) Run(g *Globals) error {
	return execBoolCheck(g, "plugin", "is-active", c.Plugin, "Plugin")
}

func (c *PluginIsInstalledCmd) Run(g *Globals) error {
	return execBoolCheck(g, "plugin", "is-installed", c.Plugin, "Plugin")
}

func (c *PluginStatusCmd) Run(g *Globals) error {
	return execPassthrough(g, "plugin", "status", c.Plugin)
}

func (c *PluginVerifyChecksumsCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin verify-checksums", nil, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("plugin", "verify-checksums")
		if c.Plugin != "" {
			builder.Arg(c.Plugin)
		} else {
			builder.BoolFlag("all")
		}
		return builder
	})
}

func (c *PluginAutoUpdatesEnableCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin auto-updates enable", []string{cache.CategoryPlugins}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("plugin", "auto-updates", "enable").Arg(c.Plugin)
	})
}

func (c *PluginAutoUpdatesDisableCmd) Run(g *Globals) error {
	return runWPCommand(g, "plugin auto-updates disable", []string{cache.CategoryPlugins}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("plugin", "auto-updates", "disable").Arg(c.Plugin)
	})
}

func (c *PluginAutoUpdatesStatusCmd) Run(g *Globals) error {
	return execPassthrough(g, "plugin", "auto-updates", "status", c.Plugin)
}
