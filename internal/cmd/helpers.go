package cmd

import (
	"context"
	"fmt"
	"strings"

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

	site, err := rc.ResolveSite()
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

	builder := wpcli.New(filtered...)
	result, err := rc.ExecWP(context.Background(), site, builder.Build(site.WPPath))
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("wp %s: %s", strings.Join(filtered, " "), result.Stderr)
	}
	fmt.Fprint(rc.Stdout, result.Stdout)
	return nil
}

// execBoolCheck runs a wp-cli boolean check command (is-active, is-installed, exists)
// and prints a human-readable result.
func execBoolCheck(g *Globals, group, subCmd, target, label string) error {
	rc, err := NewRunContext(g)
	if err != nil {
		return err
	}
	defer rc.Close()

	site, err := rc.ResolveSite()
	if err != nil {
		return err
	}

	cmd := wpcli.New(group, subCmd).Arg(target).Build(site.WPPath)
	result, err := rc.ExecWP(context.Background(), site, cmd)
	if err != nil {
		return err
	}

	verb := strings.ReplaceAll(subCmd, "-", " ")
	if result.ExitCode != 0 {
		fmt.Fprintf(rc.Stdout, "%s %s is not %s.\n", label, target, verb)
	} else {
		fmt.Fprintf(rc.Stdout, "%s %s is %s.\n", label, target, verb)
	}
	return nil
}
