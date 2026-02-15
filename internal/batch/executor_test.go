package batch

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/builtbyrobben/wpssh/internal/registry"
	"github.com/builtbyrobben/wpssh/internal/safety"
)

func makeSites(n int, canonicalHost string) []*registry.Site {
	sites := make([]*registry.Site, n)
	for i := range n {
		sites[i] = &registry.Site{
			Alias:         fmt.Sprintf("site%d", i+1),
			Hostname:      "192.0.2.10",
			Port:          37980,
			User:          fmt.Sprintf("user%d", i+1),
			WPPath:        "~/public_html",
			HostType:      "standard",
			CanonicalHost: canonicalHost,
			Tags:          map[string]string{},
		}
	}
	return sites
}

func makeMixedSites() []*registry.Site {
	return []*registry.Site{
		{Alias: "site-a", CanonicalHost: "1.2.3.4:22", Tags: map[string]string{}},
		{Alias: "site-b", CanonicalHost: "1.2.3.4:22", Tags: map[string]string{}},
		{Alias: "site-c", CanonicalHost: "5.6.7.8:22", Tags: map[string]string{}},
	}
}

func TestExecutor_Sequential_AllOK(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(3, "1.2.3.4:22")

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		return fmt.Sprintf("done %s", site.Alias), nil
	}

	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency: 1,
		Tier:        safety.TierRead,
		CommandName: "test",
	})

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Status != StatusOK {
			t.Errorf("result %d: expected OK, got %s", i, r.Status)
		}
		expected := fmt.Sprintf("done site%d", i+1)
		if r.Detail != expected {
			t.Errorf("result %d detail: got %q, want %q", i, r.Detail, expected)
		}
		if r.Duration == 0 {
			t.Errorf("result %d: duration should be > 0", i)
		}
	}
}

func TestExecutor_Sequential_WithFailures(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(3, "1.2.3.4:22")

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		if site.Alias == "site2" {
			return "", fmt.Errorf("connection refused")
		}
		return "ok", nil
	}

	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency: 1,
		Tier:        safety.TierRead,
		CommandName: "test",
	})

	if results[0].Status != StatusOK {
		t.Errorf("site1: expected OK, got %s", results[0].Status)
	}
	if results[1].Status != StatusFailed {
		t.Errorf("site2: expected FAILED, got %s", results[1].Status)
	}
	if results[1].Err == nil {
		t.Error("site2: expected error to be set")
	}
	if results[2].Status != StatusOK {
		t.Errorf("site3: expected OK, got %s", results[2].Status)
	}
}

func TestExecutor_DryRun(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(3, "1.2.3.4:22")
	called := false

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		called = true
		return "should not run", nil
	}

	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency: 1,
		DryRun:      true,
		Tier:        safety.TierRead,
		CommandName: "test",
	})

	if called {
		t.Error("command function should not be called in dry-run mode")
	}

	for i, r := range results {
		if r.Status != StatusSkipped {
			t.Errorf("result %d: expected SKIPPED, got %s", i, r.Status)
		}
		if r.Detail != "dry-run" {
			t.Errorf("result %d detail: got %q", i, r.Detail)
		}
	}
}

func TestExecutor_SafetyGate_DestructiveBlocked(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(2, "1.2.3.4:22")
	called := false

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		called = true
		return "", nil
	}

	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency: 1,
		Tier:        safety.TierDestructive,
		Yes:         false,
		CommandName: "db reset",
	})

	if called {
		t.Error("command should not run when safety gate blocks")
	}

	for _, r := range results {
		if r.Status != StatusSkipped {
			t.Errorf("expected SKIPPED, got %s", r.Status)
		}
	}
}

func TestExecutor_SafetyGate_DestructiveWithYesOnly(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(2, "1.2.3.4:22")
	called := false

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		called = true
		return "", nil
	}

	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency:    1,
		Tier:           safety.TierDestructive,
		Yes:            true,
		AckDestructive: false,
		CommandName:    "db reset",
	})

	if called {
		t.Error("command should not run without --ack-destructive")
	}

	for _, r := range results {
		if r.Status != StatusSkipped {
			t.Errorf("expected SKIPPED, got %s", r.Status)
		}
	}
}

func TestExecutor_SafetyGate_DestructiveAllowed(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(2, "1.2.3.4:22")

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		return "deleted", nil
	}

	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency:    1,
		Tier:           safety.TierDestructive,
		Yes:            true,
		AckDestructive: true,
		CommandName:    "db reset",
	})

	for _, r := range results {
		if r.Status != StatusOK {
			t.Errorf("expected OK, got %s: %s", r.Status, r.Detail)
		}
	}
}

func TestExecutor_ContextCancellation(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(5, "1.2.3.4:22")

	ctx, cancel := context.WithCancel(context.Background())
	var count atomic.Int32

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		n := count.Add(1)
		if n >= 2 {
			cancel()
		}
		return "ok", nil
	}

	results := exec.Execute(ctx, sites, fn, Options{
		Concurrency: 1,
		Tier:        safety.TierRead,
		CommandName: "test",
	})

	okCount := 0
	skipCount := 0
	for _, r := range results {
		switch r.Status {
		case StatusOK:
			okCount++
		case StatusSkipped:
			skipCount++
		}
	}

	if okCount < 2 {
		t.Errorf("expected at least 2 OK results, got %d", okCount)
	}
	if skipCount == 0 {
		t.Error("expected some skipped results after cancellation")
	}
	if len(results) != 5 {
		t.Errorf("expected 5 results total, got %d", len(results))
	}
}

func TestExecutor_Parallel(t *testing.T) {
	exec := NewExecutor()
	sites := makeMixedSites()

	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		time.Sleep(50 * time.Millisecond)
		return "ok", nil
	}

	start := time.Now()
	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency: 2,
		Tier:        safety.TierRead,
		CommandName: "test",
	})
	elapsed := time.Since(start)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for _, r := range results {
		if r.Status != StatusOK {
			t.Errorf("%s: expected OK, got %s", r.Site.Alias, r.Status)
		}
	}

	// site-a+site-b sequential (100ms) parallel with site-c (50ms) = ~100ms total.
	if elapsed > 300*time.Millisecond {
		t.Errorf("parallel execution too slow: %v (expected < 300ms)", elapsed)
	}
}

func TestExecutor_SameHostSequential(t *testing.T) {
	exec := NewExecutor()
	sites := makeSites(3, "1.2.3.4:22") // all same host

	var order []string
	fn := func(ctx context.Context, site *registry.Site) (string, error) {
		order = append(order, site.Alias)
		return "ok", nil
	}

	results := exec.Execute(context.Background(), sites, fn, Options{
		Concurrency: 5, // high concurrency, but same host = sequential
		Tier:        safety.TierRead,
		CommandName: "test",
	})

	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}
	if len(order) != 3 {
		t.Errorf("expected 3 executions, got %d", len(order))
	}
}

func TestExecutor_EmptySites(t *testing.T) {
	exec := NewExecutor()
	results := exec.Execute(context.Background(), nil, nil, Options{
		Concurrency: 1,
		Tier:        safety.TierRead,
	})
	if results != nil {
		t.Errorf("expected nil for empty sites, got %v", results)
	}
}

func TestGroupByCanonicalHost(t *testing.T) {
	sites := makeMixedSites()
	groups := groupByCanonicalHost(sites)

	if len(groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(groups))
	}
	if len(groups["1.2.3.4:22"]) != 2 {
		t.Errorf("expected 2 sites in group 1.2.3.4:22, got %d", len(groups["1.2.3.4:22"]))
	}
	if len(groups["5.6.7.8:22"]) != 1 {
		t.Errorf("expected 1 site in group 5.6.7.8:22, got %d", len(groups["5.6.7.8:22"]))
	}
}

func TestGroupByCanonicalHost_FallbackAlias(t *testing.T) {
	sites := []*registry.Site{
		{Alias: "nohost", CanonicalHost: "", Tags: map[string]string{}},
	}
	groups := groupByCanonicalHost(sites)
	if _, ok := groups["nohost"]; !ok {
		t.Error("expected fallback to alias key")
	}
}

// --- Report tests ---

func TestNewReport(t *testing.T) {
	results := []SiteResult{
		{Site: &registry.Site{Alias: "s1"}, Status: StatusOK, Duration: 1 * time.Second, Detail: "done"},
		{Site: &registry.Site{Alias: "s2"}, Status: StatusFailed, Duration: 300 * time.Millisecond, Detail: "err", Err: fmt.Errorf("fail")},
		{Site: &registry.Site{Alias: "s3"}, Status: StatusSkipped, Detail: "skipped"},
		{Site: &registry.Site{Alias: "s4"}, Status: StatusOK, Duration: 2 * time.Second, Detail: "done"},
	}

	report := NewReport(results)

	if report.Total != 4 {
		t.Errorf("total: got %d, want 4", report.Total)
	}
	if report.Success != 2 {
		t.Errorf("success: got %d, want 2", report.Success)
	}
	if report.Failed != 1 {
		t.Errorf("failed: got %d, want 1", report.Failed)
	}
	if report.Skipped != 1 {
		t.Errorf("skipped: got %d, want 1", report.Skipped)
	}
	if !report.HasFailures() {
		t.Error("expected HasFailures() = true")
	}
}

func TestReport_NoFailures(t *testing.T) {
	results := []SiteResult{
		{Site: &registry.Site{Alias: "s1"}, Status: StatusOK},
	}
	report := NewReport(results)
	if report.HasFailures() {
		t.Error("expected HasFailures() = false")
	}
}

func TestReport_WriteTable(t *testing.T) {
	results := []SiteResult{
		{Site: &registry.Site{Alias: "sitealpha"}, Status: StatusOK, Duration: 1200 * time.Millisecond, Detail: "12 plugins listed"},
		{Site: &registry.Site{Alias: "sitegamma"}, Status: StatusOK, Duration: 2100 * time.Millisecond, Detail: "8 plugins listed"},
		{Site: &registry.Site{Alias: "sitefail"}, Status: StatusFailed, Duration: 300 * time.Millisecond, Detail: "connection refused", Err: fmt.Errorf("connection refused")},
	}

	report := NewReport(results)
	var buf bytes.Buffer
	report.WriteTable(&buf)
	output := buf.String()

	if !strings.Contains(output, "SITE") {
		t.Error("missing SITE header")
	}
	if !strings.Contains(output, "STATUS") {
		t.Error("missing STATUS header")
	}
	if !strings.Contains(output, "sitealpha") {
		t.Error("missing sitealpha row")
	}
	if !strings.Contains(output, "FAILED") {
		t.Error("missing FAILED status")
	}
	if !strings.Contains(output, "Total: 3") {
		t.Error("missing total summary")
	}
	if !strings.Contains(output, "Success: 2") {
		t.Error("missing success count")
	}
}

func TestReport_WriteJSON(t *testing.T) {
	results := []SiteResult{
		{Site: &registry.Site{Alias: "s1"}, Status: StatusOK, Duration: 1 * time.Second, Detail: "ok"},
		{Site: &registry.Site{Alias: "s2"}, Status: StatusFailed, Duration: 500 * time.Millisecond, Detail: "fail", Err: fmt.Errorf("error")},
	}

	report := NewReport(results)
	var buf bytes.Buffer
	if err := report.WriteJSON(&buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, `"site": "s1"`) {
		t.Error("missing s1 in JSON")
	}
	if !strings.Contains(output, `"status": "OK"`) {
		t.Error("missing OK status in JSON")
	}
	if !strings.Contains(output, `"status": "FAILED"`) {
		t.Error("missing FAILED status in JSON")
	}
	if !strings.Contains(output, `"total": 2`) {
		t.Error("missing total in summary")
	}
}

// --- Progress tests ---

func TestProgress_NewProgress(t *testing.T) {
	p := NewProgress(10, "plugin list")
	if p.total != 10 {
		t.Errorf("total: got %d, want 10", p.total)
	}
	if p.commandName != "plugin list" {
		t.Errorf("commandName: got %q", p.commandName)
	}
}

// --- ResultStatus tests ---

func TestResultStatus_String(t *testing.T) {
	tests := []struct {
		status ResultStatus
		want   string
	}{
		{StatusOK, "OK"},
		{StatusFailed, "FAILED"},
		{StatusSkipped, "SKIPPED"},
		{ResultStatus(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.status.String()
		if got != tt.want {
			t.Errorf("ResultStatus(%d).String() = %q, want %q", tt.status, got, tt.want)
		}
	}
}

// --- Helper tests ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		input  string
		maxLen int
		want   string
	}{
		{"short", 10, "short"},
		{"exactly 10", 10, "exactly 10"},
		{"this is a longer string", 10, "this is..."},
		{"ab", 1, "a"},
		{"line1\nline2", 20, "line1 line2"},
	}

	for _, tt := range tests {
		got := truncate(tt.input, tt.maxLen)
		if got != tt.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		dur  time.Duration
		want string
	}{
		{1200 * time.Millisecond, "1.2s"},
		{300 * time.Millisecond, "0.3s"},
		{5 * time.Second, "5.0s"},
		{0, "0.0s"},
	}

	for _, tt := range tests {
		got := formatDuration(tt.dur)
		if got != tt.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tt.dur, got, tt.want)
		}
	}
}
