package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Raw command — pass-through escape hatch to wp-cli.

type RawCmd struct {
	Args []string `arg:"" optional:"" help:"Arguments to pass to wp-cli"`
}

func (c *RawCmd) Run(g *Globals) error {
	return runWPCommand(g, "raw", []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	}, func(*registry.Site) *wpcli.Command {
		return wpcli.New(c.Args...)
	})
}
