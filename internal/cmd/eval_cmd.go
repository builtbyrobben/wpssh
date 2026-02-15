package cmd

import (
	"context"
	"fmt"

	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Eval command

type EvalCmd struct {
	Code string `arg:"" help:"PHP code to evaluate"`
}

func (c *EvalCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()
	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("eval").Arg(c.Code).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp eval: %s", result.Stderr)
	}
	// eval can execute arbitrary PHP — invalidate all categories.
	rc.CacheInvalidate(site.Alias, []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	})
	fmt.Fprint(rc.Stdout, result.Stdout)
	return nil
}
