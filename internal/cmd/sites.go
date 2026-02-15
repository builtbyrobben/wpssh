package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/builtbyrobben/wpssh/internal/config"
	"github.com/builtbyrobben/wpssh/internal/registry"
)

type SitesCmd struct {
	List   SitesListCmd   `cmd:"" help:"List all registered sites"`
	Show   SitesShowCmd   `cmd:"" help:"Show site details"`
	Groups SitesGroupsCmd `cmd:"" help:"List configured site groups"`
	Add    SitesAddCmd    `cmd:"" help:"Add metadata overlay for a site"`
	Remove SitesRemoveCmd `cmd:"" help:"Remove metadata overlay"`
	Test   SitesTestCmd   `cmd:"" help:"Test SSH connectivity"`
}

// buildRegistry creates a Registry from the default paths.
func buildRegistry() (*registry.Registry, *config.Config, error) {
	paths := config.DefaultPaths()
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil, nil, fmt.Errorf("load config: %w", err)
	}
	reg, err := registry.NewRegistry(registry.RegistryOptions{
		UserGroups: cfg.UserGroups(),
	})
	if err != nil {
		return nil, nil, fmt.Errorf("build registry: %w", err)
	}
	return reg, cfg, nil
}

type SitesListCmd struct{}

func (c *SitesListCmd) Run(globals *Globals) error {
	reg, _, err := buildRegistry()
	if err != nil {
		return err
	}

	var sites []*registry.Site
	if globals.Group != "" {
		sites = reg.Filter(registry.FilterOptions{Group: globals.Group})
	} else {
		sites = reg.List()
	}

	if globals.JSON {
		data, err := json.MarshalIndent(sites, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	if len(sites) == 0 {
		fmt.Println("No sites found.")
		return nil
	}

	// Table output.
	fmt.Printf("%-20s %-30s %-10s %-12s %s\n", "ALIAS", "HOSTNAME", "PORT", "TYPE", "GROUPS")
	fmt.Printf("%-20s %-30s %-10s %-12s %s\n", "-----", "--------", "----", "----", "------")
	for _, s := range sites {
		groups := "-"
		if len(s.Groups) > 0 {
			groups = strings.Join(s.Groups, ", ")
		}
		fmt.Printf("%-20s %-30s %-10d %-12s %s\n", s.Alias, s.Hostname, s.Port, s.HostType, groups)
	}
	fmt.Printf("\nTotal: %d sites\n", len(sites))
	return nil
}

type SitesShowCmd struct {
	Alias string `arg:"" help:"Site alias to show"`
}

func (c *SitesShowCmd) Run(globals *Globals) error {
	reg, _, err := buildRegistry()
	if err != nil {
		return err
	}

	site, err := reg.Get(c.Alias)
	if err != nil {
		return err
	}

	if globals.JSON {
		data, err := json.MarshalIndent(site, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("Alias:          %s\n", site.Alias)
	fmt.Printf("Hostname:       %s\n", site.Hostname)
	fmt.Printf("Port:           %d\n", site.Port)
	fmt.Printf("User:           %s\n", site.User)
	fmt.Printf("Identity File:  %s\n", site.IdentityFile)
	fmt.Printf("WP Path:        %s\n", site.WPPath)
	fmt.Printf("Host Type:      %s\n", site.HostType)
	fmt.Printf("Canonical Host: %s\n", site.CanonicalHost)
	if len(site.Groups) > 0 {
		fmt.Printf("Groups:         %s\n", strings.Join(site.Groups, ", "))
	}
	if len(site.Tags) > 0 {
		for k, v := range site.Tags {
			fmt.Printf("Tag [%s]:       %s\n", k, v)
		}
	}
	if site.RateLimit != nil {
		fmt.Printf("Rate Limit:     delay=%s, max_conns=%d\n", site.RateLimit.Delay, site.RateLimit.MaxConns)
	}
	return nil
}

type SitesGroupsCmd struct{}

func (c *SitesGroupsCmd) Run(globals *Globals) error {
	reg, _, err := buildRegistry()
	if err != nil {
		return err
	}

	// Collect all group names: built-in + from site metadata.
	groupCounts := make(map[string]int)
	for _, g := range registry.BuiltinGroups() {
		groupCounts[g.Name] = 0
	}
	for _, site := range reg.List() {
		for _, g := range site.Groups {
			groupCounts[g] = 0
		}
	}

	// Count members per group.
	for name := range groupCounts {
		members := reg.Filter(registry.FilterOptions{Group: name})
		groupCounts[name] = len(members)
	}

	if globals.JSON {
		data, err := json.MarshalIndent(groupCounts, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("%-20s %s\n", "GROUP", "SITES")
	fmt.Printf("%-20s %s\n", "-----", "-----")
	for name, count := range groupCounts {
		fmt.Printf("%-20s %d\n", name, count)
	}
	return nil
}

type SitesAddCmd struct {
	Alias    string `arg:"" help:"Site alias to add metadata for"`
	WPPath   string `help:"WordPress install path" name:"wp-path"`
	HostType string `help:"Host type (standard, wpengine)" name:"host-type"`
	AddGroup string `help:"Add to group" name:"add-group"`
}

func (c *SitesAddCmd) Run(globals *Globals) error {
	paths := config.DefaultPaths()
	if err := paths.EnsureDirs(); err != nil {
		return err
	}

	meta, err := registry.LoadMetadata(paths.SitesFile())
	if err != nil {
		return err
	}

	overlay := meta.Sites[c.Alias]
	if c.WPPath != "" {
		overlay.WPPath = c.WPPath
	}
	if c.HostType != "" {
		overlay.HostType = c.HostType
	}
	if c.AddGroup != "" {
		// Add group if not already present.
		found := false
		for _, g := range overlay.Groups {
			if g == c.AddGroup {
				found = true
				break
			}
		}
		if !found {
			overlay.Groups = append(overlay.Groups, c.AddGroup)
		}
	}
	meta.Sites[c.Alias] = overlay

	if err := registry.SaveMetadata(meta, paths.SitesFile()); err != nil {
		return err
	}
	fmt.Printf("Updated metadata for %s\n", c.Alias)
	return nil
}

type SitesRemoveCmd struct {
	Alias string `arg:"" help:"Site alias to remove metadata for"`
}

func (c *SitesRemoveCmd) Run(globals *Globals) error {
	paths := config.DefaultPaths()
	meta, err := registry.LoadMetadata(paths.SitesFile())
	if err != nil {
		return err
	}

	if _, ok := meta.Sites[c.Alias]; !ok {
		fmt.Printf("No metadata overlay found for %s\n", c.Alias)
		return nil
	}

	delete(meta.Sites, c.Alias)
	if err := registry.SaveMetadata(meta, paths.SitesFile()); err != nil {
		return err
	}
	fmt.Printf("Removed metadata for %s\n", c.Alias)
	return nil
}

type SitesTestCmd struct {
	Alias string `arg:"" optional:"" help:"Site alias to test (or --all)"`
	All   bool   `help:"Test all sites"`
}

func (c *SitesTestCmd) Run(globals *Globals) error {
	// SSH connectivity testing requires the SSH client (Phase 1).
	// For now, verify the site exists in the registry.
	reg, _, err := buildRegistry()
	if err != nil {
		return err
	}

	if c.All {
		fmt.Printf("Found %d sites in registry. SSH testing not yet implemented.\n", reg.Len())
		return nil
	}

	if c.Alias == "" {
		return fmt.Errorf("provide a site alias or use --all")
	}

	site, err := reg.Get(c.Alias)
	if err != nil {
		return err
	}
	fmt.Printf("Site %s found: %s@%s:%d (SSH testing not yet implemented)\n",
		site.Alias, site.User, site.Hostname, site.Port)
	return nil
}
