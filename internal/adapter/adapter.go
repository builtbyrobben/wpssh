package adapter

import (
	"context"
	"time"

	"github.com/builtbyrobben/wpssh/internal/registry"
	internalssh "github.com/builtbyrobben/wpssh/internal/ssh"
)

// AdapterCapabilities describes what a host adapter supports.
type AdapterCapabilities struct {
	SupportsSCP        bool
	PersistentFS       bool
	MaxSessionDuration time.Duration // 0 means no limit.
}

// Adapter defines how commands and file transfers are executed on a host.
// Different hosting environments (cPanel, WP Engine) implement this interface.
type Adapter interface {
	// Exec runs a shell command on the remote site.
	// Callers are responsible for building the full remote command string.
	Exec(ctx context.Context, client *internalssh.SSHClient, site *registry.Site, wpCmd string) (internalssh.ExecResult, error)

	// Upload transfers a local file to the remote host.
	Upload(ctx context.Context, client *internalssh.SSHClient, site *registry.Site, localPath, remotePath string) error

	// Download transfers a remote file to local disk.
	Download(ctx context.Context, client *internalssh.SSHClient, site *registry.Site, remotePath, localPath string) error

	// Capabilities returns what this adapter supports.
	Capabilities() AdapterCapabilities

	// Name returns the adapter identifier (e.g., "standard", "wpengine").
	Name() string
}
