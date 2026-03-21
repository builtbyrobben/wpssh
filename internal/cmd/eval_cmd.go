package cmd

import (
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Eval command

type EvalCmd struct {
	Code string `arg:"" help:"PHP code to evaluate"`
}

func (c *EvalCmd) Run(g *Globals) error {
	return runWPCommand(g, "eval", []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	}, func(*registry.Site) *wpcli.Command {
		return wpcli.New("eval").Arg(c.Code)
	})
}
