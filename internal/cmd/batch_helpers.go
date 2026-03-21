package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/builtbyrobben/wpssh/internal/batch"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/safety"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

func writeBatchReport(rc *RunContext, results []batch.SiteResult) error {
	report := batch.NewReport(results)
	if rc.Formatter.JSON {
		if err := report.WriteJSON(rc.Stdout); err != nil {
			return err
		}
	} else {
		report.WriteTable(rc.Stdout)
	}

	if report.HasFailures() {
		return fmt.Errorf("batch command failed on %d site(s)", report.Failed)
	}
	return nil
}

func batchOptions(g *Globals, tier safety.SafetyTier, commandName string) batch.Options {
	return batch.Options{
		Concurrency:    g.Concurrency,
		Yes:            g.Yes,
		AckDestructive: g.AckDestructive,
		DryRun:         g.DryRun,
		Tier:           tier,
		CommandName:    commandName,
	}
}

func summarizeOutput(stdout string) string {
	text := strings.TrimSpace(stdout)
	if text == "" {
		return "ok"
	}
	line := strings.Split(text, "\n")[0]
	if len(line) > 96 {
		return line[:93] + "..."
	}
	return line
}

func appendRowsWithSite[T any](dst *[]map[string]any, siteAlias string, rows []T) error {
	for _, row := range rows {
		mapped, err := valueWithSite(siteAlias, row)
		if err != nil {
			return err
		}
		*dst = append(*dst, mapped)
	}
	return nil
}

func valueWithSite(siteAlias string, value any) (map[string]any, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	var mapped map[string]any
	if err := json.Unmarshal(data, &mapped); err != nil {
		return nil, err
	}
	if mapped == nil {
		mapped = make(map[string]any)
	}
	mapped["site"] = siteAlias
	return mapped, nil
}

func runStructuredListCommand[T any](g *Globals, commandName string, cacheCategory string, builderFactory func(*registry.Site) *wpcli.Command) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	sites, err := rc.ResolveSites()
	if err != nil {
		return err
	}

	if len(sites) == 1 && !g.IsBatchMode() {
		rows, err := fetchStructuredList[T](context.Background(), rc, sites[0], cacheCategory, commandName, builderFactory)
		if err != nil {
			return err
		}
		return rc.Formatter.Format(rows)
	}

	var (
		mu        sync.Mutex
		flattened []map[string]any
	)

	results := batch.NewExecutor().Execute(context.Background(), sites, func(ctx context.Context, site *registry.Site) (string, error) {
		rows, err := fetchStructuredList[T](ctx, rc, site, cacheCategory, commandName, builderFactory)
		if err != nil {
			return "", err
		}

		mu.Lock()
		defer mu.Unlock()
		if err := appendRowsWithSite(&flattened, site.Alias, rows); err != nil {
			return "", err
		}
		return fmt.Sprintf("%d row(s)", len(rows)), nil
	}, batchOptions(g, safety.Classify(strings.Fields(commandName)...), commandName))

	report := batch.NewReport(results)
	if report.HasFailures() {
		return writeBatchReport(rc, results)
	}

	sort.Slice(flattened, func(i, j int) bool {
		left, _ := flattened[i]["site"].(string)
		right, _ := flattened[j]["site"].(string)
		if left == right {
			return fmt.Sprint(flattened[i]) < fmt.Sprint(flattened[j])
		}
		return left < right
	})
	return rc.Formatter.Format(flattened)
}

func fetchStructuredList[T any](ctx context.Context, rc *RunContext, site *registry.Site, cacheCategory, commandName string, builderFactory func(*registry.Site) *wpcli.Command) ([]T, error) {
	builder := builderFactory(site)
	cacheKey := builder.CacheKey()

	if cacheCategory != "" {
		if cached := rc.CacheGet(site.Alias, cacheKey); cached != "" {
			rows, err := wpcli.ParseJSON[T](cached)
			if err == nil {
				return rows, nil
			}
		}
	}

	result, err := rc.ExecWP(ctx, site, builder.Build(site.WPPath))
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("wp %s: %s", commandName, strings.TrimSpace(result.Stderr))
	}

	rows, err := wpcli.ParseJSON[T](result.Stdout)
	if err != nil {
		return nil, err
	}

	if cacheCategory != "" {
		if data, err := json.Marshal(rows); err == nil {
			rc.CacheSet(site.Alias, cacheKey, commandName, string(data), cacheCategory, rc.CacheTTL(cacheCategory))
		}
	}

	return rows, nil
}

func runStructuredSingleCommand[T any](g *Globals, commandName string, builderFactory func(*registry.Site) *wpcli.Command) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	sites, err := rc.ResolveSites()
	if err != nil {
		return err
	}

	if len(sites) == 1 && !g.IsBatchMode() {
		value, err := fetchStructuredSingle[T](context.Background(), rc, sites[0], commandName, builderFactory)
		if err != nil {
			return err
		}
		return rc.Formatter.Format(value)
	}

	var (
		mu        sync.Mutex
		flattened []map[string]any
	)

	results := batch.NewExecutor().Execute(context.Background(), sites, func(ctx context.Context, site *registry.Site) (string, error) {
		value, err := fetchStructuredSingle[T](ctx, rc, site, commandName, builderFactory)
		if err != nil {
			return "", err
		}
		mapped, err := valueWithSite(site.Alias, value)
		if err != nil {
			return "", err
		}

		mu.Lock()
		flattened = append(flattened, mapped)
		mu.Unlock()
		return summarizeOutput(fmt.Sprint(mapped)), nil
	}, batchOptions(g, safety.Classify(strings.Fields(commandName)...), commandName))

	report := batch.NewReport(results)
	if report.HasFailures() {
		return writeBatchReport(rc, results)
	}

	sort.Slice(flattened, func(i, j int) bool {
		left, _ := flattened[i]["site"].(string)
		right, _ := flattened[j]["site"].(string)
		return left < right
	})
	return rc.Formatter.Format(flattened)
}

func fetchStructuredSingle[T any](ctx context.Context, rc *RunContext, site *registry.Site, commandName string, builderFactory func(*registry.Site) *wpcli.Command) (*T, error) {
	result, err := rc.ExecWP(ctx, site, builderFactory(site).Build(site.WPPath))
	if err != nil {
		return nil, err
	}
	if result.ExitCode != 0 {
		return nil, fmt.Errorf("wp %s: %s", commandName, strings.TrimSpace(result.Stderr))
	}
	return wpcli.ParseSingle[T](result.Stdout)
}

func runWPCommand(g *Globals, commandName string, invalidateCategories []string, builderFactory func(*registry.Site) *wpcli.Command) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	sites, err := rc.ResolveSites()
	if err != nil {
		return err
	}

	if len(sites) == 1 && !g.IsBatchMode() {
		site := sites[0]
		result, err := rc.ExecWP(context.Background(), site, builderFactory(site).Build(site.WPPath))
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("wp %s: %s", commandName, strings.TrimSpace(result.Stderr))
		}
		if len(invalidateCategories) > 0 {
			rc.CacheInvalidate(site.Alias, invalidateCategories)
		}
		fmt.Fprint(rc.Stdout, result.Stdout)
		return nil
	}

	results := batch.NewExecutor().Execute(context.Background(), sites, func(ctx context.Context, site *registry.Site) (string, error) {
		result, err := rc.ExecWP(ctx, site, builderFactory(site).Build(site.WPPath))
		if err != nil {
			return "", err
		}
		if result.ExitCode != 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(result.Stderr))
		}
		if len(invalidateCategories) > 0 {
			rc.CacheInvalidate(site.Alias, invalidateCategories)
		}
		return summarizeOutput(result.Stdout), nil
	}, batchOptions(g, safety.Classify(strings.Fields(commandName)...), commandName))

	return writeBatchReport(rc, results)
}
