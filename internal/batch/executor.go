package batch

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/safety"
)

// CommandFunc executes a command on a single site and returns the result detail string.
// It is called by the executor for each site in the batch.
type CommandFunc func(ctx context.Context, site *registry.Site) (detail string, err error)

// Options configures batch execution.
type Options struct {
	Concurrency    int  // Max parallel executions (default 1 = sequential)
	Yes            bool // --yes flag
	AckDestructive bool // --ack-destructive flag
	DryRun         bool // --dry-run: show what would run without executing
	Tier           safety.SafetyTier
	CommandName    string // Human-readable command description for progress display
}

// SiteResult holds the outcome of a command executed on a single site.
type SiteResult struct {
	Site     *registry.Site
	Status   ResultStatus
	Detail   string
	Duration time.Duration
	Err      error
}

// ResultStatus represents the outcome of a single-site execution.
type ResultStatus int

const (
	StatusOK ResultStatus = iota
	StatusFailed
	StatusSkipped
)

// String returns a human-readable status label.
func (s ResultStatus) String() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusFailed:
		return "FAILED"
	case StatusSkipped:
		return "SKIPPED"
	default:
		return "UNKNOWN"
	}
}

// Executor runs commands across multiple sites with safety gates and rate limiting.
type Executor struct{}

// NewExecutor creates a batch executor.
func NewExecutor() *Executor {
	return &Executor{}
}

// Execute runs the given command function across all sites, respecting safety
// gates, concurrency limits, and canonical host bucketing.
//
// Same-canonical-host sites are always executed sequentially (the rate limiter
// handles spacing). Different-host sites may run in parallel up to opts.Concurrency.
func (e *Executor) Execute(ctx context.Context, sites []*registry.Site, fn CommandFunc, opts Options) []SiteResult {
	if len(sites) == 0 {
		return nil
	}

	// Safety gate check before any execution.
	if err := safety.CheckBatchSafety(opts.Tier, opts.Yes, opts.AckDestructive); err != nil {
		results := make([]SiteResult, len(sites))
		for i, site := range sites {
			results[i] = SiteResult{
				Site:   site,
				Status: StatusSkipped,
				Detail: err.Error(),
			}
		}
		return results
	}

	progress := NewProgress(len(sites), opts.CommandName)

	// Dry-run: show what would execute without doing anything.
	if opts.DryRun {
		results := make([]SiteResult, len(sites))
		for i, site := range sites {
			progress.Update(i+1, site.Alias, "dry-run (skipped)")
			results[i] = SiteResult{
				Site:   site,
				Status: StatusSkipped,
				Detail: "dry-run",
			}
		}
		progress.Done()
		return results
	}

	concurrency := opts.Concurrency
	if concurrency < 1 {
		concurrency = 1
	}

	// Group sites by canonical host for safe ordering.
	hostGroups := groupByCanonicalHost(sites)

	if concurrency == 1 {
		// Sequential execution: process all sites in order.
		return e.executeSequential(ctx, sites, fn, progress)
	}

	// Parallel execution with canonical host safety.
	return e.executeParallel(ctx, hostGroups, fn, concurrency, progress)
}

// executeSequential runs commands one at a time.
func (e *Executor) executeSequential(ctx context.Context, sites []*registry.Site, fn CommandFunc, progress *Progress) []SiteResult {
	results := make([]SiteResult, len(sites))

	for i, site := range sites {
		if ctx.Err() != nil {
			results[i] = SiteResult{
				Site:   site,
				Status: StatusSkipped,
				Detail: "context cancelled",
			}
			continue
		}

		progress.Update(i+1, site.Alias, fmt.Sprintf("running %s...", progress.commandName))

		start := time.Now()
		detail, err := fn(ctx, site)
		dur := time.Since(start)

		if err != nil {
			results[i] = SiteResult{
				Site:     site,
				Status:   StatusFailed,
				Detail:   err.Error(),
				Duration: dur,
				Err:      err,
			}
		} else {
			results[i] = SiteResult{
				Site:     site,
				Status:   StatusOK,
				Detail:   detail,
				Duration: dur,
			}
		}
	}

	progress.Done()
	return results
}

// executeParallel runs site groups in parallel, but sites within the same
// canonical host group are executed sequentially.
func (e *Executor) executeParallel(ctx context.Context, hostGroups map[string][]*registry.Site, fn CommandFunc, concurrency int, progress *Progress) []SiteResult {
	var mu sync.Mutex
	var results []SiteResult
	var completed int

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup

	// Sort host keys for deterministic ordering.
	hosts := make([]string, 0, len(hostGroups))
	for h := range hostGroups {
		hosts = append(hosts, h)
	}
	sort.Strings(hosts)

	for _, host := range hosts {
		group := hostGroups[host]
		wg.Add(1)
		go func(sites []*registry.Site) {
			defer wg.Done()

			// Acquire one of the concurrency slots.
			sem <- struct{}{}
			defer func() { <-sem }()

			// Same-host sites run sequentially within this goroutine.
			for _, site := range sites {
				if ctx.Err() != nil {
					mu.Lock()
					completed++
					results = append(results, SiteResult{
						Site:   site,
						Status: StatusSkipped,
						Detail: "context cancelled",
					})
					mu.Unlock()
					continue
				}

				mu.Lock()
				completed++
				n := completed
				mu.Unlock()

				progress.Update(n, site.Alias, fmt.Sprintf("running %s...", progress.commandName))

				start := time.Now()
				detail, err := fn(ctx, site)
				dur := time.Since(start)

				r := SiteResult{
					Site:     site,
					Duration: dur,
				}
				if err != nil {
					r.Status = StatusFailed
					r.Detail = err.Error()
					r.Err = err
				} else {
					r.Status = StatusOK
					r.Detail = detail
				}

				mu.Lock()
				results = append(results, r)
				mu.Unlock()
			}
		}(group)
	}

	wg.Wait()
	progress.Done()

	return results
}

// groupByCanonicalHost buckets sites by their resolved canonical host.
func groupByCanonicalHost(sites []*registry.Site) map[string][]*registry.Site {
	groups := make(map[string][]*registry.Site)
	for _, s := range sites {
		host := s.CanonicalHost
		if host == "" {
			host = s.Alias // fallback
		}
		groups[host] = append(groups[host], s)
	}
	return groups
}
