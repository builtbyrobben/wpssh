package registry

import "strings"

// GroupDef defines a site group with explicit members and auto-detection rules.
type GroupDef struct {
	Name    string   // Group name (e.g., "prime-vps", "wpengine")
	Aliases []string // Explicit site aliases in this group
	// AutoDetect is a function that returns true if a site should be auto-included.
	AutoDetect func(site *Site) bool
}

// BuiltinGroups returns the built-in group definitions.
func BuiltinGroups() []GroupDef {
	return []GroupDef{
		{
			Name: "all",
			AutoDetect: func(site *Site) bool {
				return true
			},
		},
		{
			Name: "prime-vps",
			AutoDetect: func(site *Site) bool {
				return site.Hostname == "192.0.2.10" ||
					strings.HasPrefix(site.CanonicalHost, "192.0.2.10:")
			},
		},
		{
			Name: "wpengine",
			AutoDetect: func(site *Site) bool {
				return site.HostType == "wpengine" ||
					strings.HasSuffix(site.Hostname, ".ssh.wpengine.net")
			},
		},
	}
}

// MatchGroup returns true if the site belongs to the named group, checking:
// 1. Built-in group auto-detection rules
// 2. User-defined groups from config (by alias list)
// 3. Site metadata group memberships
func MatchGroup(site *Site, groupName string, userGroups map[string][]string) bool {
	// Check built-in groups first.
	for _, g := range BuiltinGroups() {
		if g.Name == groupName {
			if g.AutoDetect != nil && g.AutoDetect(site) {
				return true
			}
			for _, alias := range g.Aliases {
				if alias == site.Alias {
					return true
				}
			}
		}
	}

	// Check user-defined groups from config.
	if aliases, ok := userGroups[groupName]; ok {
		for _, alias := range aliases {
			if alias == site.Alias {
				return true
			}
		}
	}

	// Check site's own group memberships from metadata.
	for _, g := range site.Groups {
		if g == groupName {
			return true
		}
	}

	return false
}
