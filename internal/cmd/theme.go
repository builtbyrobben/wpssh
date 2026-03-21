package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Theme commands

type ThemeCmd struct {
	List        ThemeListCmd        `cmd:"" help:"List themes"`
	Install     ThemeInstallCmd     `cmd:"" help:"Install a theme"`
	Activate    ThemeActivateCmd    `cmd:"" help:"Activate a theme"`
	Delete      ThemeDeleteCmd      `cmd:"" help:"Delete a theme"`
	Update      ThemeUpdateCmd      `cmd:"" help:"Update theme(s)"`
	Search      ThemeSearchCmd      `cmd:"" help:"Search WordPress.org"`
	Get         ThemeGetCmd         `cmd:"" help:"Get theme details"`
	IsActive    ThemeIsActiveCmd    `cmd:"" name:"is-active" help:"Check if active"`
	IsInstalled ThemeIsInstalledCmd `cmd:"" name:"is-installed" help:"Check if installed"`
	Status      ThemeStatusCmd      `cmd:"" help:"Show theme status"`
	AutoUpdates ThemeAutoUpdatesCmd `cmd:"" name:"auto-updates" help:"Manage auto-updates"`
}

type ThemeListCmd struct {
	StatusFilter string `help:"Filter by status" name:"status"`
}
type ThemeInstallCmd struct {
	Theme    string `arg:"" help:"Theme slug"`
	Activate bool   `help:"Activate after install"`
}
type ThemeActivateCmd struct {
	Theme string `arg:"" help:"Theme slug"`
}
type ThemeDeleteCmd struct {
	Theme string `arg:"" help:"Theme slug"`
}
type ThemeUpdateCmd struct {
	Theme string `arg:"" optional:"" help:"Theme slug"`
	All   bool   `help:"Update all themes"`
}
type ThemeSearchCmd struct {
	Term string `arg:"" help:"Search term"`
}
type ThemeGetCmd struct {
	Theme string `arg:"" help:"Theme slug"`
}
type ThemeIsActiveCmd struct {
	Theme string `arg:"" help:"Theme slug"`
}
type ThemeIsInstalledCmd struct {
	Theme string `arg:"" help:"Theme slug"`
}
type ThemeStatusCmd struct {
	Theme string `arg:"" optional:"" help:"Theme slug"`
}
type ThemeAutoUpdatesCmd struct {
	Enable  ThemeAutoUpdatesEnableCmd  `cmd:"" help:"Enable auto-updates"`
	Disable ThemeAutoUpdatesDisableCmd `cmd:"" help:"Disable auto-updates"`
	Status  ThemeAutoUpdatesStatusCmd  `cmd:"" help:"Show auto-update status"`
}
type ThemeAutoUpdatesEnableCmd struct {
	Theme string `arg:"" help:"Theme slug"`
}
type ThemeAutoUpdatesDisableCmd struct {
	Theme string `arg:"" help:"Theme slug"`
}
type ThemeAutoUpdatesStatusCmd struct {
	Theme string `arg:"" optional:"" help:"Theme slug"`
}

func (c *ThemeListCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.Theme](g, "theme list", cache.CategoryThemes, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("theme", "list").Format("json")
		if c.StatusFilter != "" {
			builder.Flag("status", c.StatusFilter)
		}
		return builder
	})
}

func (c *ThemeInstallCmd) Run(g *Globals) error {
	return runWPCommand(g, "theme install", []string{cache.CategoryThemes, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("theme", "install").Arg(c.Theme)
		if c.Activate {
			builder.BoolFlag("activate")
		}
		return builder
	})
}

func (c *ThemeActivateCmd) Run(g *Globals) error {
	return runWPCommand(g, "theme activate", []string{cache.CategoryThemes, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("theme", "activate").Arg(c.Theme)
	})
}

func (c *ThemeDeleteCmd) Run(g *Globals) error {
	return runWPCommand(g, "theme delete", []string{cache.CategoryThemes, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("theme", "delete").Arg(c.Theme)
	})
}

func (c *ThemeUpdateCmd) Run(g *Globals) error {
	return runWPCommand(g, "theme update", []string{cache.CategoryThemes, cache.CategorySnapshot}, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("theme", "update")
		if c.All {
			builder.BoolFlag("all")
		} else if c.Theme != "" {
			builder.Arg(c.Theme)
		}
		return builder
	})
}

func (c *ThemeSearchCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.Theme](g, "theme search", "", func(*registry.Site) *wpcli.Command {
		return wpcli.New("theme", "search").Arg(c.Term).Format("json")
	})
}

func (c *ThemeGetCmd) Run(g *Globals) error {
	return runStructuredSingleCommand[wpcli.Theme](g, "theme get", func(*registry.Site) *wpcli.Command {
		return wpcli.New("theme", "get").Arg(c.Theme).Format("json")
	})
}

func (c *ThemeIsActiveCmd) Run(g *Globals) error {
	return execBoolCheck(g, "theme", "is-active", c.Theme, "Theme")
}

func (c *ThemeIsInstalledCmd) Run(g *Globals) error {
	return execBoolCheck(g, "theme", "is-installed", c.Theme, "Theme")
}

func (c *ThemeStatusCmd) Run(g *Globals) error {
	return execPassthrough(g, "theme", "status", c.Theme)
}

func (c *ThemeAutoUpdatesEnableCmd) Run(g *Globals) error {
	return runWPCommand(g, "theme auto-updates enable", []string{cache.CategoryThemes}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("theme", "auto-updates", "enable").Arg(c.Theme)
	})
}

func (c *ThemeAutoUpdatesDisableCmd) Run(g *Globals) error {
	return runWPCommand(g, "theme auto-updates disable", []string{cache.CategoryThemes}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("theme", "auto-updates", "disable").Arg(c.Theme)
	})
}

func (c *ThemeAutoUpdatesStatusCmd) Run(g *Globals) error {
	return execPassthrough(g, "theme", "auto-updates", "status", c.Theme)
}
