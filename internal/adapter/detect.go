package adapter

import (
	"strings"

	"github.com/builtbyrobben/wpssh/internal/registry"
)

// ForSite returns the appropriate adapter for a site based on its host type
// or auto-detection from hostname patterns.
//
// Detection priority:
//  1. Explicit host_type in site metadata ("standard", "wpengine")
//  2. Hostname pattern matching (*.ssh.wpengine.net -> WP Engine)
//  3. Default: StandardAdapter
func ForSite(site *registry.Site) Adapter {
	switch strings.ToLower(site.HostType) {
	case "wpengine":
		return &WPEngineAdapter{}
	case "standard":
		return &StandardAdapter{}
	case "", "auto":
		return detectFromHostname(site.Hostname)
	default:
		return &StandardAdapter{}
	}
}

// detectFromHostname determines the adapter from hostname patterns.
func detectFromHostname(hostname string) Adapter {
	hostname = strings.ToLower(hostname)

	// WP Engine SSH hostnames match *.ssh.wpengine.net
	if strings.HasSuffix(hostname, ".ssh.wpengine.net") {
		return &WPEngineAdapter{}
	}

	return &StandardAdapter{}
}
