package adapter

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/builtbyrobben/wpssh/internal/registry"
	internalssh "github.com/builtbyrobben/wpssh/internal/ssh"
)

// StandardAdapter handles cPanel, VPS, and other standard SSH hosts.
// Commands run via "cd {wp_path} && wp {command}".
// File transfers use stdin/stdout streaming (cat).
type StandardAdapter struct{}

var _ Adapter = (*StandardAdapter)(nil)

func (a *StandardAdapter) Name() string { return "standard" }

func (a *StandardAdapter) Capabilities() AdapterCapabilities {
	return AdapterCapabilities{
		SupportsSCP:        true,
		PersistentFS:       true,
		MaxSessionDuration: 0, // No limit.
	}
}

// Exec runs a wp-cli command on a standard host.
// Wraps the command as: cd {wp_path} && wp {wpCmd}
func (a *StandardAdapter) Exec(ctx context.Context, client *internalssh.SSHClient, site *registry.Site, wpCmd string) (internalssh.ExecResult, error) {
	cfg := siteToClientConfig(site)
	cmd := fmt.Sprintf("cd %s && wp %s", shellQuote(site.WPPath), wpCmd)
	return client.Exec(ctx, cfg, site.CanonicalHost, cmd)
}

// Upload streams a local file to the remote host via stdin.
// Runs: cat > {remotePath} with the file content piped to stdin.
func (a *StandardAdapter) Upload(ctx context.Context, client *internalssh.SSHClient, site *registry.Site, localPath, remotePath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open local file: %w", err)
	}
	defer f.Close()

	cfg := siteToClientConfig(site)
	cmd := fmt.Sprintf("cat > %s", shellQuote(remotePath))
	result, err := client.ExecWithStdin(ctx, cfg, site.CanonicalHost, cmd, f)
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("upload failed (exit %d): %s", result.ExitCode, result.Stderr)
	}
	return nil
}

// Download streams a remote file to local disk via stdout.
// Runs: cat {remotePath} and writes stdout to localPath.
func (a *StandardAdapter) Download(ctx context.Context, client *internalssh.SSHClient, site *registry.Site, remotePath, localPath string) error {
	cfg := siteToClientConfig(site)
	cmd := fmt.Sprintf("cat %s", shellQuote(remotePath))
	result, err := client.Exec(ctx, cfg, site.CanonicalHost, cmd)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("download failed (exit %d): %s", result.ExitCode, result.Stderr)
	}
	if err := os.WriteFile(localPath, []byte(result.Stdout), 0o644); err != nil {
		return fmt.Errorf("write local file: %w", err)
	}
	return nil
}

// siteToClientConfig converts a registry.Site to an ssh.ClientConfig.
func siteToClientConfig(site *registry.Site) internalssh.ClientConfig {
	return internalssh.ClientConfig{
		Host:           site.Hostname,
		Port:           site.Port,
		User:           site.User,
		IdentityFile:   site.IdentityFile,
		ConnectTimeout: 30 * time.Second,
	}
}

// shellQuote wraps a string in single quotes for safe shell usage.
// Embedded single quotes are escaped with the shell idiom:
//
//	end-quote + backslash-quote + start-quote  ('”'”')
//
// For example, “it's” becomes 'it'”'”'s'.
func shellQuote(s string) string {
	result := "'"
	for _, c := range s {
		if c == '\'' {
			result += "'\\''"
		} else {
			result += string(c)
		}
	}
	result += "'"
	return result
}
