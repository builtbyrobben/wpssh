package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/builtbyrobben/wpssh/internal/config"
	"github.com/builtbyrobben/wpssh/internal/safety"
)

// SetupCmd guides users through creating/updating wpgo config.
type SetupCmd struct {
	DefaultSite    string `help:"Set default site alias" name:"default-site"`
	DefaultFormat  string `help:"Set default output format (table, json, plain)" name:"default-format"`
	NonInteractive bool   `help:"Apply provided flags without prompts" name:"non-interactive"`
}

func (c *SetupCmd) Run(g *Globals) error {
	return c.runWithIO(g, os.Stdin, os.Stdout)
}

func (c *SetupCmd) runWithIO(g *Globals, in io.Reader, out io.Writer) error {
	paths := config.DefaultPaths()
	if err := paths.EnsureDirs(); err != nil {
		return err
	}

	cfgPath := paths.ConfigFile()
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	changed, err := applySetupFlags(cfg, c)
	if err != nil {
		return err
	}
	interactive := safety.IsTTY() && !c.NonInteractive && !g.Yes

	if interactive {
		interactiveChanged, err := runInteractiveSetup(cfg, in, out)
		if err != nil {
			return err
		}
		changed = changed || interactiveChanged
	}

	exists := fileExists(cfgPath)
	if !changed && exists {
		fmt.Fprintf(out, "No changes made. Existing config: %s\n", cfgPath)
		return nil
	}

	if err := config.Save(cfg, cfgPath); err != nil {
		return err
	}

	fmt.Fprintf(out, "Saved config: %s\n", cfgPath)
	printSetupSummary(out, cfg)
	return nil
}

func runInteractiveSetup(cfg *config.Config, in io.Reader, out io.Writer) (bool, error) {
	p := newPrompter(in, out)
	changed := false

	fmt.Fprintln(out, "wpgo interactive setup")
	fmt.Fprintln(out, "Choose which config sections you want to update.")
	fmt.Fprintln(out)

	if ok, err := p.yesNo("Configure default site alias?", cfg.DefaultSite == ""); err != nil {
		return false, err
	} else if ok {
		val, err := p.input("Default site alias (blank to clear)", cfg.DefaultSite)
		if err != nil {
			return false, err
		}
		val = strings.TrimSpace(val)
		if val != cfg.DefaultSite {
			cfg.DefaultSite = val
			changed = true
		}
	}

	if ok, err := p.yesNo("Configure default output format?", true); err != nil {
		return false, err
	} else if ok {
		format, err := p.choice("Default format", cfg.DefaultFormat, []string{"table", "json", "plain"})
		if err != nil {
			return false, err
		}
		if format != cfg.DefaultFormat {
			cfg.DefaultFormat = format
			changed = true
		}
	}

	if ok, err := p.yesNo("Configure default rate limit?", false); err != nil {
		return false, err
	} else if ok {
		delay, err := p.duration("Default delay", cfg.DefaultRateLimit.Delay)
		if err != nil {
			return false, err
		}
		maxConns, err := p.intValue("Default max connections", cfg.DefaultRateLimit.MaxConns, 1)
		if err != nil {
			return false, err
		}
		if delay != cfg.DefaultRateLimit.Delay || maxConns != cfg.DefaultRateLimit.MaxConns {
			cfg.DefaultRateLimit.Delay = delay
			cfg.DefaultRateLimit.MaxConns = maxConns
			changed = true
		}
	}

	if ok, err := p.yesNo("Configure cache TTLs?", false); err != nil {
		return false, err
	} else if ok {
		plugins, err := p.duration("Cache TTL: plugins", cfg.CacheTTLs.Plugins)
		if err != nil {
			return false, err
		}
		themes, err := p.duration("Cache TTL: themes", cfg.CacheTTLs.Themes)
		if err != nil {
			return false, err
		}
		core, err := p.duration("Cache TTL: core", cfg.CacheTTLs.Core)
		if err != nil {
			return false, err
		}
		users, err := p.duration("Cache TTL: users", cfg.CacheTTLs.Users)
		if err != nil {
			return false, err
		}
		options, err := p.duration("Cache TTL: options", cfg.CacheTTLs.Options)
		if err != nil {
			return false, err
		}
		snapshot, err := p.duration("Cache TTL: snapshot", cfg.CacheTTLs.Snapshot)
		if err != nil {
			return false, err
		}
		if plugins != cfg.CacheTTLs.Plugins ||
			themes != cfg.CacheTTLs.Themes ||
			core != cfg.CacheTTLs.Core ||
			users != cfg.CacheTTLs.Users ||
			options != cfg.CacheTTLs.Options ||
			snapshot != cfg.CacheTTLs.Snapshot {
			cfg.CacheTTLs.Plugins = plugins
			cfg.CacheTTLs.Themes = themes
			cfg.CacheTTLs.Core = core
			cfg.CacheTTLs.Users = users
			cfg.CacheTTLs.Options = options
			cfg.CacheTTLs.Snapshot = snapshot
			changed = true
		}
	}

	if ok, err := p.yesNo("Add or update site groups?", len(cfg.Groups) == 0); err != nil {
		return false, err
	} else if ok {
		if cfg.Groups == nil {
			cfg.Groups = map[string]config.GroupConfig{}
		}
		for {
			name, err := p.input("Group name (blank to finish)", "")
			if err != nil {
				return false, err
			}
			name = strings.TrimSpace(name)
			if name == "" {
				break
			}

			existing := cfg.Groups[name]
			aliasesRaw, err := p.input("Aliases (comma-separated)", strings.Join(cfg.Groups[name].Aliases, ","))
			if err != nil {
				return false, err
			}

			aliases := parseCSV(aliasesRaw)
			cfg.Groups[name] = config.GroupConfig{Aliases: aliases}
			if !equalStringSlice(existing.Aliases, aliases) {
				changed = true
			}
		}
	}

	return changed, nil
}

func applySetupFlags(cfg *config.Config, c *SetupCmd) (bool, error) {
	changed := false

	if c.DefaultSite != "" {
		cfg.DefaultSite = c.DefaultSite
		changed = true
	}
	if c.DefaultFormat != "" {
		if format, ok := normalizeFormat(c.DefaultFormat); ok {
			cfg.DefaultFormat = format
			changed = true
		} else {
			return false, fmt.Errorf("invalid --default-format %q (valid: table, json, plain)", c.DefaultFormat)
		}
	}

	return changed, nil
}

func normalizeFormat(v string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "table":
		return "table", true
	case "json":
		return "json", true
	case "plain":
		return "plain", true
	default:
		return "", false
	}
}

func printSetupSummary(out io.Writer, cfg *config.Config) {
	fmt.Fprintln(out)
	fmt.Fprintln(out, "Config summary:")
	fmt.Fprintf(out, "  default_site: %q\n", cfg.DefaultSite)
	fmt.Fprintf(out, "  default_format: %s\n", cfg.DefaultFormat)
	fmt.Fprintf(out, "  default_rate_limit: delay=%s max_conns=%d\n", cfg.DefaultRateLimit.Delay, cfg.DefaultRateLimit.MaxConns)
	fmt.Fprintf(out, "  groups: %d\n", len(cfg.Groups))
}

func parseCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	return out
}

func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type prompter struct {
	r *bufio.Reader
	w io.Writer
}

func newPrompter(in io.Reader, out io.Writer) *prompter {
	return &prompter{
		r: bufio.NewReader(in),
		w: out,
	}
}

func (p *prompter) yesNo(label string, defaultYes bool) (bool, error) {
	defaultHint := "y/N"
	if defaultYes {
		defaultHint = "Y/n"
	}

	for {
		fmt.Fprintf(p.w, "%s [%s]: ", label, defaultHint)
		raw, err := p.readLine()
		if err != nil {
			return false, err
		}
		v := strings.ToLower(strings.TrimSpace(raw))
		if v == "" {
			return defaultYes, nil
		}
		if v == "y" || v == "yes" {
			return true, nil
		}
		if v == "n" || v == "no" {
			return false, nil
		}
		fmt.Fprintln(p.w, "Please enter y or n.")
	}
}

func (p *prompter) input(label, defaultValue string) (string, error) {
	if defaultValue != "" {
		fmt.Fprintf(p.w, "%s [%s]: ", label, defaultValue)
	} else {
		fmt.Fprintf(p.w, "%s: ", label)
	}

	raw, err := p.readLine()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(raw) == "" {
		return defaultValue, nil
	}
	return raw, nil
}

func (p *prompter) choice(label, defaultValue string, valid []string) (string, error) {
	validSet := map[string]bool{}
	for _, v := range valid {
		validSet[v] = true
	}

	for {
		raw, err := p.input(fmt.Sprintf("%s (%s)", label, strings.Join(valid, "/")), defaultValue)
		if err != nil {
			return "", err
		}
		v := strings.ToLower(strings.TrimSpace(raw))
		if validSet[v] {
			return v, nil
		}
		fmt.Fprintf(p.w, "Invalid value %q. Valid options: %s\n", v, strings.Join(valid, ", "))
	}
}

func (p *prompter) duration(label string, defaultValue time.Duration) (time.Duration, error) {
	for {
		raw, err := p.input(label, defaultValue.String())
		if err != nil {
			return 0, err
		}
		v, err := time.ParseDuration(strings.TrimSpace(raw))
		if err != nil {
			fmt.Fprintln(p.w, "Invalid duration. Example values: 500ms, 2s, 30m, 1h.")
			continue
		}
		return v, nil
	}
}

func (p *prompter) intValue(label string, defaultValue, min int) (int, error) {
	for {
		raw, err := p.input(label, strconv.Itoa(defaultValue))
		if err != nil {
			return 0, err
		}
		v, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || v < min {
			fmt.Fprintf(p.w, "Please enter a number >= %d.\n", min)
			continue
		}
		return v, nil
	}
}

func (p *prompter) readLine() (string, error) {
	s, err := p.r.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return strings.TrimSpace(s), nil
		}
		return "", err
	}
	return strings.TrimSpace(s), nil
}
