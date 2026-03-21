package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"slices"
	"time"

	"github.com/builtbyrobben/wpssh/internal/adapter"
	"github.com/builtbyrobben/wpssh/internal/cache"
	"github.com/builtbyrobben/wpssh/internal/config"
	"github.com/builtbyrobben/wpssh/internal/outfmt"
	"github.com/builtbyrobben/wpssh/internal/registry"
	internalssh "github.com/builtbyrobben/wpssh/internal/ssh"
)

// RunContext wires together all the services needed to execute a wp-cli command.
type RunContext struct {
	Registry  *registry.Registry
	Config    *config.Config
	SSHClient *internalssh.SSHClient
	Cache     *cache.Cache
	Formatter *outfmt.Formatter
	Globals   *Globals
	Stdout    io.Writer
	Stderr    io.Writer
}

// NewRunContext builds a fully wired RunContext from Globals.
// The caller must call Close() when done to release SSH connections and cache.
func NewRunContext(g *Globals) (*RunContext, error) {
	paths := config.DefaultPaths()
	cfg, err := config.Load(paths.ConfigFile())
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	reg, err := registry.NewRegistry(registry.RegistryOptions{
		UserGroups: cfg.UserGroups(),
	})
	if err != nil {
		return nil, fmt.Errorf("build registry: %w", err)
	}

	// Build rate limiter with config.
	hostConfigs := make(map[string]internalssh.HostConfig)
	for _, site := range reg.List() {
		hostConfigs[site.CanonicalHost] = resolveHostConfig(cfg, site)
	}
	limiter := internalssh.NewRateLimiter(hostConfigs)
	pool := internalssh.NewPool(limiter, 5*time.Minute)
	sshClient := internalssh.NewSSHClient(pool)

	var cacheStore *cache.Cache
	if !g.NoCache {
		cacheStore, err = cache.New(paths.CacheDB())
		if err != nil {
			// Non-fatal: proceed without cache.
			fmt.Fprintf(os.Stderr, "warning: cache unavailable: %v\n", err)
		}
	}

	jsonFlag, plainFlag := resolveFormatFlags(cfg, g)
	formatter := outfmt.New(jsonFlag, plainFlag, g.Fields)

	return &RunContext{
		Registry:  reg,
		Config:    cfg,
		SSHClient: sshClient,
		Cache:     cacheStore,
		Formatter: formatter,
		Globals:   g,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
	}, nil
}

// Close releases resources held by the RunContext.
func (rc *RunContext) Close() {
	if rc.SSHClient != nil {
		rc.SSHClient.Close()
	}
	if rc.Cache != nil {
		rc.Cache.Close()
	}
}

// ResolveSite resolves the target site from --site flag or config default.
func (rc *RunContext) ResolveSite() (*registry.Site, error) {
	alias := rc.Globals.Site
	if alias == "" {
		alias = rc.Config.DefaultSite
	}
	if alias == "" {
		return nil, fmt.Errorf("no site specified: use --site or set default_site in config")
	}
	return rc.Registry.Get(alias)
}

// ResolveSites resolves the current target set for single-site or batch mode.
func (rc *RunContext) ResolveSites() ([]*registry.Site, error) {
	if rc.Globals.Site != "" && (len(rc.Globals.Sites) > 0 || rc.Globals.Group != "") {
		return nil, fmt.Errorf("use either --site or batch targeting flags, not both")
	}
	if len(rc.Globals.Sites) > 0 && rc.Globals.Group != "" {
		return nil, fmt.Errorf("use either --sites or --group, not both")
	}

	if len(rc.Globals.Sites) > 0 {
		sites := make([]*registry.Site, 0, len(rc.Globals.Sites))
		seen := make(map[string]struct{}, len(rc.Globals.Sites))
		for _, alias := range rc.Globals.Sites {
			site, err := rc.Registry.Get(alias)
			if err != nil {
				return nil, err
			}
			if _, ok := seen[site.Alias]; ok {
				continue
			}
			seen[site.Alias] = struct{}{}
			sites = append(sites, site)
		}
		return sites, nil
	}

	if rc.Globals.Group != "" {
		sites := rc.Registry.Filter(registry.FilterOptions{Group: rc.Globals.Group})
		if len(sites) == 0 {
			return nil, fmt.Errorf("no sites found for group %q", rc.Globals.Group)
		}
		slices.SortFunc(sites, func(a, b *registry.Site) int {
			switch {
			case a.Alias < b.Alias:
				return -1
			case a.Alias > b.Alias:
				return 1
			default:
				return 0
			}
		})
		return sites, nil
	}

	site, err := rc.ResolveSite()
	if err != nil {
		return nil, err
	}
	return []*registry.Site{site}, nil
}

// ExecWP executes a wp-cli command on a site and returns the result.
func (rc *RunContext) ExecWP(ctx context.Context, site *registry.Site, wpCmd string) (internalssh.ExecResult, error) {
	a := adapter.ForSite(site)
	if rc.Globals.Verbose {
		fmt.Fprintf(rc.Stderr, "[adapter: %s] %s\n", a.Name(), wpCmd)
	}
	if rc.Globals.DryRun {
		fmt.Fprintf(rc.Stderr, "[dry-run] %s\n", wpCmd)
		return internalssh.ExecResult{}, nil
	}
	return a.Exec(ctx, rc.SSHClient, site, wpCmd)
}

// CacheGet checks the cache for a stored result. Returns the cached data
// string or empty string on miss. Handles nil cache gracefully.
func (rc *RunContext) CacheGet(siteAlias, commandCacheKey string) string {
	if rc.Cache == nil {
		return ""
	}
	key := cache.MakeCacheKey(siteAlias, commandCacheKey)
	entry, err := rc.Cache.Get(siteAlias, key)
	if err != nil || entry == nil {
		return ""
	}
	// Check TTL expiry.
	if time.Since(entry.FetchedAt) > time.Duration(entry.TTLSeconds)*time.Second {
		return ""
	}
	if rc.Globals.Verbose {
		fmt.Fprintf(rc.Stderr, "[cache hit] %s\n", entry.Command)
	}
	return entry.Data
}

// CacheSet stores a result in the cache. Handles nil cache gracefully.
func (rc *RunContext) CacheSet(siteAlias, commandCacheKey, command, data, category string, ttlSeconds int) {
	if rc.Cache == nil {
		return
	}
	key := cache.MakeCacheKey(siteAlias, commandCacheKey)
	entry := &cache.CacheEntry{
		SiteAlias:  siteAlias,
		CacheKey:   key,
		Command:    command,
		Data:       data,
		FetchedAt:  time.Now(),
		TTLSeconds: ttlSeconds,
		Category:   category,
	}
	if err := rc.Cache.Set(entry); err != nil && rc.Globals.Verbose {
		fmt.Fprintf(rc.Stderr, "[cache] store error: %v\n", err)
	}
}

// CacheInvalidate invalidates cache entries for a site. Handles nil cache.
func (rc *RunContext) CacheInvalidate(siteAlias string, categories []string) {
	if rc.Cache == nil {
		return
	}
	if err := rc.Cache.Invalidate(siteAlias, categories); err != nil && rc.Globals.Verbose {
		fmt.Fprintf(rc.Stderr, "[cache] invalidate error: %v\n", err)
	}
}

// CacheTTL returns the configured TTL, in seconds, for a cache category.
func (rc *RunContext) CacheTTL(category string) int {
	switch category {
	case cache.CategoryPlugins:
		return int(rc.Config.CacheTTLs.Plugins.Seconds())
	case cache.CategoryThemes:
		return int(rc.Config.CacheTTLs.Themes.Seconds())
	case cache.CategoryCore:
		return int(rc.Config.CacheTTLs.Core.Seconds())
	case cache.CategoryUsers:
		return int(rc.Config.CacheTTLs.Users.Seconds())
	case cache.CategoryOptions:
		return int(rc.Config.CacheTTLs.Options.Seconds())
	case cache.CategorySnapshot:
		return int(rc.Config.CacheTTLs.Snapshot.Seconds())
	default:
		return 0
	}
}

// SSHConfig builds an SSH ClientConfig from a registry Site.
func SSHConfig(site *registry.Site) internalssh.ClientConfig {
	return internalssh.ClientConfig{
		Host:           site.Hostname,
		Port:           site.Port,
		User:           site.User,
		IdentityFile:   site.IdentityFile,
		ConnectTimeout: 30 * time.Second,
	}
}

func resolveFormatFlags(cfg *config.Config, g *Globals) (jsonFlag, plainFlag bool) {
	if g.JSON || g.Plain {
		return g.JSON, g.Plain
	}

	switch cfg.DefaultFormat {
	case "json":
		return true, false
	case "plain":
		return false, true
	default:
		return false, false
	}
}

func resolveHostConfig(cfg *config.Config, site *registry.Site) internalssh.HostConfig {
	hostCfg := internalssh.HostConfig{
		Delay:    internalssh.DefaultHostConfig.Delay,
		MaxConns: internalssh.DefaultHostConfig.MaxConns,
	}

	if cfg.DefaultRateLimit.Delay > 0 {
		hostCfg.Delay = cfg.DefaultRateLimit.Delay
	}
	if cfg.DefaultRateLimit.MaxConns > 0 {
		hostCfg.MaxConns = cfg.DefaultRateLimit.MaxConns
	}

	if rl, ok := cfg.RateLimits[site.CanonicalHost]; ok {
		if rl.Delay > 0 {
			hostCfg.Delay = rl.Delay
		}
		if rl.MaxConns > 0 {
			hostCfg.MaxConns = rl.MaxConns
		}
	}

	if site.RateLimit != nil {
		if site.RateLimit.Delay > 0 {
			hostCfg.Delay = site.RateLimit.Delay
		}
		if site.RateLimit.MaxConns > 0 {
			hostCfg.MaxConns = site.RateLimit.MaxConns
		}
	}

	return hostCfg
}
