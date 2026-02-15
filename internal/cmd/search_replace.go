package cmd

import (
	"context"
	"fmt"

	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// Search-replace command

type SearchReplaceCmd struct {
	Old string `arg:"" help:"String to search for"`
	New string `arg:"" help:"Replacement string"`
}

func (c *SearchReplaceCmd) Run(g *Globals) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()
	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New("search-replace").Arg(c.Old).Arg(c.New).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp search-replace: %s", result.Stderr)
	}
	// search-replace can modify any table — invalidate all categories.
	rc.CacheInvalidate(site.Alias, []string{
		cache.CategoryPlugins, cache.CategoryThemes, cache.CategoryCore,
		cache.CategoryUsers, cache.CategoryOptions, cache.CategorySnapshot,
	})
	fmt.Fprint(rc.Stdout, result.Stdout)
	return nil
}
