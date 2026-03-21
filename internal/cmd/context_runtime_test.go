package cmd

import (
	"testing"
	"time"

	"github.com/builtbyrobben/wpssh/internal/config"
	"github.com/builtbyrobben/wpssh/internal/registry"
	sshclient "github.com/builtbyrobben/wpssh/internal/ssh"
)

func TestResolveFormatFlags(t *testing.T) {
	cfg := &config.Config{DefaultFormat: "plain"}

	jsonFlag, plainFlag := resolveFormatFlags(cfg, &Globals{})
	if jsonFlag || !plainFlag {
		t.Fatalf("default plain format not applied: json=%v plain=%v", jsonFlag, plainFlag)
	}

	jsonFlag, plainFlag = resolveFormatFlags(cfg, &Globals{JSON: true})
	if !jsonFlag || plainFlag {
		t.Fatalf("explicit --json should win: json=%v plain=%v", jsonFlag, plainFlag)
	}
}

func TestResolveHostConfig(t *testing.T) {
	cfg := &config.Config{
		DefaultRateLimit: config.RateLimitEntry{
			Delay:    2 * time.Second,
			MaxConns: 4,
		},
		RateLimits: map[string]config.RateLimitEntry{
			"203.0.113.10:22": {
				Delay:    5 * time.Second,
				MaxConns: 2,
			},
		},
	}

	site := &registry.Site{
		Alias:         "prod-a",
		CanonicalHost: "203.0.113.10:22",
		RateLimit: &registry.RateLimitConfig{
			Delay: 7 * time.Second,
		},
	}

	got := resolveHostConfig(cfg, site)
	if got.Delay != 7*time.Second {
		t.Fatalf("delay = %s, want 7s", got.Delay)
	}
	if got.MaxConns != 2 {
		t.Fatalf("max_conns = %d, want 2", got.MaxConns)
	}

	other := resolveHostConfig(cfg, &registry.Site{Alias: "prod-b", CanonicalHost: "198.51.100.8:22"})
	if other.Delay != 2*time.Second || other.MaxConns != 4 {
		t.Fatalf("default host config = %+v, want delay=2s max_conns=4", other)
	}

	if sshclient.DefaultHostConfig.Delay == 0 || sshclient.DefaultHostConfig.MaxConns == 0 {
		t.Fatal("default ssh host config should remain valid")
	}
}
