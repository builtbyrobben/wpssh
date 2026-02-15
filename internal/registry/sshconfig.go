package registry

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ssh_config "github.com/kevinburke/ssh_config"
)

// SSHEntry represents a single host entry parsed from an SSH config file.
type SSHEntry struct {
	Alias        string
	Hostname     string
	Port         int
	User         string
	IdentityFile string
}

// ParseSSHConfig reads and parses the user's ~/.ssh/config file.
func ParseSSHConfig() ([]SSHEntry, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, ".ssh", "config")
	return ParseSSHConfigFile(path)
}

// ParseSSHConfigFile parses an SSH config from the given file path.
func ParseSSHConfigFile(path string) ([]SSHEntry, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ParseSSHConfigReader(f)
}

// ParseSSHConfigReader parses an SSH config from an io.Reader.
func ParseSSHConfigReader(r io.Reader) ([]SSHEntry, error) {
	cfg, err := ssh_config.Decode(r)
	if err != nil {
		return nil, err
	}
	return extractEntries(cfg)
}

func extractEntries(cfg *ssh_config.Config) ([]SSHEntry, error) {
	var entries []SSHEntry

	for _, host := range cfg.Hosts {
		for _, pattern := range host.Patterns {
			alias := pattern.String()

			// Skip wildcard entries and negated patterns.
			if alias == "*" || strings.ContainsAny(alias, "*?") || strings.HasPrefix(alias, "!") {
				continue
			}

			entry := SSHEntry{
				Alias: alias,
				Port:  22, // default
			}

			// Extract key-value nodes from this host block.
			for _, node := range host.Nodes {
				kv, ok := node.(*ssh_config.KV)
				if !ok {
					continue
				}
				switch strings.ToLower(kv.Key) {
				case "hostname":
					entry.Hostname = kv.Value
				case "port":
					if p, err := strconv.Atoi(kv.Value); err == nil {
						entry.Port = p
					}
				case "user":
					entry.User = kv.Value
				case "identityfile":
					entry.IdentityFile = expandHome(kv.Value)
				}
			}

			// If no explicit Hostname, the alias itself is the hostname.
			if entry.Hostname == "" {
				entry.Hostname = alias
			}

			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// expandHome replaces ~ prefix with the user's home directory.
func expandHome(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, path[1:])
}
