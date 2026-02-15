package registry

import (
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
)

const hostTypeStandard = "standard"

// Registry provides a merged view of SSH config entries + metadata overlay.
type Registry struct {
	sites        map[string]*Site    // alias → merged site
	sorted       []*Site             // sorted by alias
	canonicalMap map[string]string   // alias → "IP:port"
	userGroups   map[string][]string // group name → explicit aliases (from config)
}

// RegistryOptions configures how the registry is built.
type RegistryOptions struct {
	// SSHConfigPath overrides the default ~/.ssh/config path.
	SSHConfigPath string
	// MetadataPath overrides the default ~/.config/wpgo/sites.json path.
	MetadataPath string
	// UserGroups are user-defined groups from app config (group name → aliases).
	UserGroups map[string][]string
	// SkipDNS disables DNS resolution for canonical host mapping (for testing).
	SkipDNS bool
}

// NewRegistry parses SSH config and metadata, merges them, and resolves canonical hosts.
func NewRegistry(opts RegistryOptions) (*Registry, error) {
	// Parse SSH config.
	var entries []SSHEntry
	var err error
	if opts.SSHConfigPath != "" {
		entries, err = ParseSSHConfigFile(opts.SSHConfigPath)
	} else {
		entries, err = ParseSSHConfig()
	}
	if err != nil {
		return nil, fmt.Errorf("parse ssh config: %w", err)
	}

	// Load metadata overlay.
	metaPath := opts.MetadataPath
	if metaPath == "" {
		metaPath = defaultMetadataPath()
	}
	meta, err := LoadMetadata(metaPath)
	if err != nil {
		return nil, fmt.Errorf("load metadata: %w", err)
	}

	// Build merged sites.
	reg := &Registry{
		sites:        make(map[string]*Site),
		canonicalMap: make(map[string]string),
		userGroups:   opts.UserGroups,
	}
	if reg.userGroups == nil {
		reg.userGroups = make(map[string][]string)
	}

	for _, entry := range entries {
		site := &Site{
			Alias:        entry.Alias,
			Hostname:     entry.Hostname,
			Port:         entry.Port,
			User:         entry.User,
			IdentityFile: entry.IdentityFile,
			WPPath:       "~/public_html", // sensible default
			HostType:     hostTypeStandard,
			Tags:         make(map[string]string),
		}

		// Apply metadata overlay if present.
		if overlay, ok := meta.Sites[entry.Alias]; ok {
			if overlay.WPPath != "" {
				site.WPPath = overlay.WPPath
			}
			if overlay.HostType != "" {
				site.HostType = overlay.HostType
			}
			if len(overlay.Groups) > 0 {
				site.Groups = overlay.Groups
			}
			if overlay.Tags != nil {
				site.Tags = overlay.Tags
			}
			site.RateLimit = overlay.RateLimit.ToRateLimitConfig()
		}

		// Auto-detect host type from hostname patterns.
		site.HostType = detectHostType(site)

		// Resolve canonical host for rate limiting.
		site.CanonicalHost = resolveCanonicalHost(site.Hostname, site.Port, opts.SkipDNS)

		reg.sites[entry.Alias] = site
	}

	// Build sorted list.
	reg.sorted = make([]*Site, 0, len(reg.sites))
	for _, s := range reg.sites {
		reg.sorted = append(reg.sorted, s)
		reg.canonicalMap[s.Alias] = s.CanonicalHost
	}
	sort.Slice(reg.sorted, func(i, j int) bool {
		return reg.sorted[i].Alias < reg.sorted[j].Alias
	})

	return reg, nil
}

// Get returns the merged site for the given alias.
func (r *Registry) Get(alias string) (*Site, error) {
	s, ok := r.sites[alias]
	if !ok {
		return nil, fmt.Errorf("site %q not found in registry", alias)
	}
	return s, nil
}

// List returns all known sites, sorted by alias.
func (r *Registry) List() []*Site {
	return r.sorted
}

// Filter returns sites matching the given group and/or tag.
func (r *Registry) Filter(opts FilterOptions) []*Site {
	return FilterSites(r.sorted, opts, r.userGroups)
}

// CanonicalHostMap returns a map of alias → resolved IP:port.
func (r *Registry) CanonicalHostMap() map[string]string {
	cp := make(map[string]string, len(r.canonicalMap))
	for k, v := range r.canonicalMap {
		cp[k] = v
	}
	return cp
}

// Len returns the number of registered sites.
func (r *Registry) Len() int {
	return len(r.sites)
}

// detectHostType auto-detects the host type from hostname patterns.
// Returns the overlay host_type if explicitly set, otherwise auto-detects.
func detectHostType(site *Site) string {
	// If explicitly set to something other than "auto" and "standard", respect it.
	if site.HostType != "" && site.HostType != "auto" && site.HostType != hostTypeStandard {
		return site.HostType
	}

	hostname := site.Hostname
	if strings.HasSuffix(hostname, ".ssh.wpengine.net") {
		return "wpengine"
	}
	// cPanel on Prime VPS.
	if hostname == "192.0.2.10" {
		return hostTypeStandard
	}
	return hostTypeStandard
}

// resolveCanonicalHost resolves a hostname + port to "IP:port" for rate limiting.
func resolveCanonicalHost(hostname string, port int, skipDNS bool) string {
	portStr := strconv.Itoa(port)

	// If already an IP, use directly.
	if ip := net.ParseIP(hostname); ip != nil {
		return net.JoinHostPort(hostname, portStr)
	}

	if skipDNS {
		return net.JoinHostPort(hostname, portStr)
	}

	// Resolve hostname to IP.
	ips, err := net.LookupHost(hostname)
	if err != nil || len(ips) == 0 {
		// Fall back to hostname if DNS fails.
		return net.JoinHostPort(hostname, portStr)
	}
	return net.JoinHostPort(ips[0], portStr)
}

func defaultMetadataPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return home + "/.config/wpgo/sites.json"
}
