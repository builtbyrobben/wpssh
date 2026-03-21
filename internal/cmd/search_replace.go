package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Search-replace command

type SearchReplaceCmd struct {
	Old string `arg:"" help:"String to search for"`
	New string `arg:"" help:"Replacement string"`
}

func (c *SearchReplaceCmd) Run(g *Globals) error {
	return runWPCommand(g, "search-replace", []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("search-replace").Arg(c.Old).Arg(c.New)
	})
}
