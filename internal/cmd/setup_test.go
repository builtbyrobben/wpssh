package cmd

import (
	"bytes"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/builtbyrobben/wpssh/internal/config"
)

func TestParseCSV(t *testing.T) {
	got := parseCSV(" alpha, beta,alpha , , gamma ")
	want := []string{"alpha", "beta", "gamma"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseCSV: got %#v, want %#v", got, want)
	}
}

func TestApplySetupFlags(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := &SetupCmd{
		DefaultSite:   "site-a",
		DefaultFormat: "json",
	}

	changed, err := applySetupFlags(&cfg, cmd)
	if err != nil {
		t.Fatalf("applySetupFlags: %v", err)
	}
	if !changed {
		t.Fatal("expected changed=true")
	}
	if cfg.DefaultSite != "site-a" {
		t.Fatalf("DefaultSite: got %q", cfg.DefaultSite)
	}
	if cfg.DefaultFormat != "json" {
		t.Fatalf("DefaultFormat: got %q", cfg.DefaultFormat)
	}
}

func TestApplySetupFlags_InvalidFormat(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := &SetupCmd{DefaultFormat: "xml"}

	changed, err := applySetupFlags(&cfg, cmd)
	if err == nil {
		t.Fatal("expected error")
	}
	if changed {
		t.Fatal("expected changed=false")
	}
}

func TestSetupRunWithIO_NonInteractiveSavesConfig(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)

	cmd := &SetupCmd{
		DefaultSite:    "prod-main",
		DefaultFormat:  "plain",
		NonInteractive: true,
	}
	globals := &Globals{}

	var out bytes.Buffer
	if err := cmd.runWithIO(globals, bytes.NewBuffer(nil), &out); err != nil {
		t.Fatalf("runWithIO: %v", err)
	}

	cfgPath := filepath.Join(tmp, "wpgo", "config.json")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.DefaultSite != "prod-main" {
		t.Fatalf("DefaultSite: got %q", cfg.DefaultSite)
	}
	if cfg.DefaultFormat != "plain" {
		t.Fatalf("DefaultFormat: got %q", cfg.DefaultFormat)
	}
}
