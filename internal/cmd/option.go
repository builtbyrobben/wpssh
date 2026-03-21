package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Option commands

type OptionCmd struct {
	Get    OptionGetCmd    `cmd:"" help:"Get option value"`
	Update OptionUpdateCmd `cmd:"" help:"Update option"`
	Add    OptionAddCmd    `cmd:"" help:"Add option"`
	Delete OptionDeleteCmd `cmd:"" help:"Delete option"`
	List   OptionListCmd   `cmd:"" help:"List options"`
	Patch  OptionPatchCmd  `cmd:"" help:"Patch option"`
	Pluck  OptionPluckCmd  `cmd:"" help:"Pluck option value"`
}

type OptionGetCmd struct {
	Key string `arg:"" help:"Option name"`
}
type OptionUpdateCmd struct {
	Key   string `arg:"" help:"Option name"`
	Value string `arg:"" help:"Option value"`
}
type OptionAddCmd struct {
	Key   string `arg:"" help:"Option name"`
	Value string `arg:"" help:"Option value"`
}
type OptionDeleteCmd struct {
	Key string `arg:"" help:"Option name"`
}
type OptionListCmd struct {
	Search string `help:"Search pattern"`
}
type OptionPatchCmd struct {
	Action  string `arg:"" help:"Patch action"`
	Key     string `arg:"" help:"Option name"`
	KeyPath string `arg:"" help:"Key path"`
	Value   string `arg:"" optional:"" help:"Value"`
}
type OptionPluckCmd struct {
	Key     string `arg:"" help:"Option name"`
	KeyPath string `arg:"" help:"Key path"`
}

func (c *OptionGetCmd) Run(g *Globals) error {
	return execPassthrough(g, "option", "get", c.Key)
}

func (c *OptionUpdateCmd) Run(g *Globals) error {
	return runWPCommand(g, "option update", []string{cache.CategoryOptions}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("option", "update").Arg(c.Key).Arg(c.Value)
	})
}

func (c *OptionAddCmd) Run(g *Globals) error {
	return runWPCommand(g, "option add", []string{cache.CategoryOptions}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("option", "add").Arg(c.Key).Arg(c.Value)
	})
}

func (c *OptionDeleteCmd) Run(g *Globals) error {
	return runWPCommand(g, "option delete", []string{cache.CategoryOptions}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("option", "delete").Arg(c.Key)
	})
}

func (c *OptionListCmd) Run(g *Globals) error {
	return runStructuredListCommand[wpcli.Option](g, "option list", cache.CategoryOptions, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("option", "list").Format("json")
		if c.Search != "" {
			builder.Flag("search", c.Search)
		}
		return builder
	})
}

func (c *OptionPatchCmd) Run(g *Globals) error {
	return runWPCommand(g, "option patch", []string{cache.CategoryOptions}, func(*registry.Site) *wpcli.Command {
		builder := wpcli.New("option", "patch").Arg(c.Action).Arg(c.Key).Arg(c.KeyPath)
		if c.Value != "" {
			builder.Arg(c.Value)
		}
		return builder
	})
}

func (c *OptionPluckCmd) Run(g *Globals) error {
	return execPassthrough(g, "option", "pluck", c.Key, c.KeyPath)
}
