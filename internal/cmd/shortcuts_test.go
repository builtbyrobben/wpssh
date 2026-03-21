package cmd

import (
	"strings"
	"testing"
)

func TestBuildScriptCommand(t *testing.T) {
	got := buildScriptCommand("/srv/www/site path", "echo hello", []string{"Client Name", "pre-deploy"})

	if !strings.Contains(got, "cd '/srv/www/site path' && bash -s -- 'Client Name' 'pre-deploy'") {
		t.Fatalf("script command missing expected cd/args: %q", got)
	}
	if !strings.Contains(got, "<<'WPGO_SCRIPT'\necho hello\nWPGO_SCRIPT") {
		t.Fatalf("script heredoc missing: %q", got)
	}
}
