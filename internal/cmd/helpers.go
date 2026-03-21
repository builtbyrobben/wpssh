package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/builtbyrobben/wpssh/internal/batch"
	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/safety"
	"github.com/builtbyrobben/wpssh/internal/wpcli"
)

// execPassthrough runs a wp-cli command and prints its stdout directly.
// Parts are the wp-cli subcommands (e.g., "theme", "status").
// If the last part is empty, it is omitted (allows optional args).
func execPassthrough(g *Globals, parts ...string) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	sites, err := rc.ResolveSites()
	if err != nil {
		return err
	}

	// Filter out empty strings (optional args that weren't provided)
	var filtered []string
	for _, p := range parts {
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	commandName := strings.Join(filtered, " ")

	if len(sites) == 1 && !g.IsBatchMode() {
		site := sites[0]
		builder := wpcli.New(filtered...)
		result, err := rc.ExecWP(context.Background(), site, builder.Build(site.WPPath))
		if err != nil {
			return err
		}
		if result.ExitCode != 0 {
			return fmt.Errorf("wp %s: %s", commandName, result.Stderr)
		}
		fmt.Fprint(rc.Stdout, result.Stdout)
		return nil
	}

	results := batch.NewExecutor().Execute(context.Background(), sites, func(ctx context.Context, site *registry.Site) (string, error) {
		result, err := rc.ExecWP(ctx, site, wpcli.New(filtered...).Build(site.WPPath))
		if err != nil {
			return "", err
		}
		if result.ExitCode != 0 {
			return "", fmt.Errorf("%s", strings.TrimSpace(result.Stderr))
		}
		return summarizeOutput(result.Stdout), nil
	}, batchOptions(g, safety.Classify(filtered...), commandName))

	return writeBatchReport(rc, results)
}

// execBoolCheck runs a wp-cli boolean check command (is-active, is-installed, exists)
// and prints a human-readable result.
func execBoolCheck(g *Globals, group, subCmd, target, label string) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	sites, err := rc.ResolveSites()
	if err != nil {
		return err
	}

	verb := strings.ReplaceAll(subCmd, "-", " ")
	commandName := strings.Join([]string{group, subCmd}, " ")

	if len(sites) == 1 && !g.IsBatchMode() {
		site := sites[0]
		cmd := wpcli.New(group, subCmd).Arg(target).Build(site.WPPath)
		result, err := rc.ExecWP(context.Background(), site, cmd)
		if err != nil {
			return err
		}

		if result.ExitCode != 0 {
			fmt.Fprintf(rc.Stdout, "%s %s is not %s.\n", label, target, verb)
		} else {
			fmt.Fprintf(rc.Stdout, "%s %s is %s.\n", label, target, verb)
		}
		return nil
	}

	results := batch.NewExecutor().Execute(context.Background(), sites, func(ctx context.Context, site *registry.Site) (string, error) {
		cmd := wpcli.New(group, subCmd).Arg(target).Build(site.WPPath)
		result, err := rc.ExecWP(ctx, site, cmd)
		if err != nil {
			return "", err
		}
		if result.ExitCode != 0 {
			return fmt.Sprintf("%s %s is not %s", label, target, verb), nil
		}
		return fmt.Sprintf("%s %s is %s", label, target, verb), nil
	}, batchOptions(g, safety.Classify(group, subCmd), commandName))

	return writeBatchReport(rc, results)
}
