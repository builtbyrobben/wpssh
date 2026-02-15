# wpssh (wpgo)

A CLI tool for managing WordPress sites over SSH built with Go. Run WP-CLI commands, manage plugins, themes, databases, and more across multiple WordPress sites from a single interface.

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap builtbyrobben/tap
brew install wpssh
```

### Download Binary

Download the latest release from [GitHub Releases](https://github.com/builtbyrobben/wpssh/releases).

### Build from Source

```bash
git clone https://github.com/builtbyrobben/wpssh.git
cd wpssh
make build
```

## Configuration

wpgo discovers sites from your SSH config (`~/.ssh/config`) and enriches them with optional metadata overlays. Run the interactive setup to get started.

```bash
wpgo setup
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `WPGO_SITE` | Default target site alias |

### Site Registry

Sites are auto-discovered from SSH config. You can add metadata overlays to enrich them:

```bash
# List all sites
wpgo sites list

# Show site details
wpgo sites show mysite

# Add metadata (WordPress path, host type, groups)
wpgo sites add mysite --wp-path /var/www/html --host-type wpengine --add-group production

# Remove metadata overlay
wpgo sites remove mysite

# List site groups
wpgo sites groups

# Test SSH connectivity
wpgo sites test mysite
wpgo sites test --all
```

## Commands

### Targeting Sites

Most commands require a target site:

```bash
# Target a single site
wpgo -s mysite plugin list

# Target multiple sites (batch mode)
wpgo --sites mysite1,mysite2 plugin list

# Target a site group
wpgo -g production plugin list

# Use WPGO_SITE env var as default
export WPGO_SITE=mysite
wpgo plugin list
```

### sites -- Site registry management

```bash
wpgo sites list                         # List all registered sites
wpgo sites show <alias>                 # Show site details
wpgo sites groups                       # List configured groups
wpgo sites add <alias> [--wp-path ...] [--host-type ...] [--add-group ...]
wpgo sites remove <alias>               # Remove metadata overlay
wpgo sites test <alias>                 # Test SSH connectivity
wpgo sites test --all                   # Test all sites
```

### plugin -- Plugin management

```bash
wpgo -s mysite plugin list                   # List all plugins
wpgo -s mysite plugin list --status active   # Filter by status
wpgo -s mysite plugin install woocommerce    # Install a plugin
wpgo -s mysite plugin install woocommerce --activate  # Install and activate
wpgo -s mysite plugin activate woocommerce   # Activate a plugin
wpgo -s mysite plugin deactivate woocommerce # Deactivate a plugin
wpgo -s mysite plugin delete woocommerce     # Delete a plugin
wpgo -s mysite plugin update woocommerce     # Update a plugin
wpgo -s mysite plugin update --all           # Update all plugins
wpgo -s mysite plugin search "seo"           # Search WordPress.org
wpgo -s mysite plugin get woocommerce        # Get plugin details
wpgo -s mysite plugin is-active woocommerce  # Check if active
wpgo -s mysite plugin is-installed woocommerce  # Check if installed
wpgo -s mysite plugin status                 # Show plugin status
wpgo -s mysite plugin verify-checksums       # Verify all plugin checksums
wpgo -s mysite plugin auto-updates enable woocommerce   # Enable auto-updates
wpgo -s mysite plugin auto-updates disable woocommerce  # Disable auto-updates
wpgo -s mysite plugin auto-updates status    # Show auto-update status
```

### theme -- Theme management

```bash
wpgo -s mysite theme list
wpgo -s mysite theme install flavor
wpgo -s mysite theme activate flavor
wpgo -s mysite theme delete flavor
wpgo -s mysite theme update --all
```

### core -- WordPress core management

```bash
wpgo -s mysite core version
wpgo -s mysite core update
wpgo -s mysite core verify-checksums
```

### db -- Database operations

```bash
wpgo -s mysite db export
wpgo -s mysite db import dump.sql
wpgo -s mysite db query "SELECT COUNT(*) FROM wp_posts"
wpgo -s mysite db size
wpgo -s mysite db tables
wpgo -s mysite db optimize
wpgo -s mysite db repair
```

### user -- User management

```bash
wpgo -s mysite user list
wpgo -s mysite user get admin
wpgo -s mysite user create --email new@example.com --role editor
```

### post -- Post management

```bash
wpgo -s mysite post list
wpgo -s mysite post get 42
wpgo -s mysite post delete 42
```

### option -- Options management

```bash
wpgo -s mysite option get siteurl
wpgo -s mysite option update blogdescription "My Site"
```

### search-replace -- Database search and replace

```bash
wpgo -s mysite search-replace "http://old.example.com" "https://new.example.com"
```

### cache -- Object cache management

```bash
wpgo -s mysite cache flush
wpgo -s mysite cache type
```

### transient -- Transients management

```bash
wpgo -s mysite transient delete --all
wpgo -s mysite transient get my_transient
```

### cron -- WP-Cron management

```bash
wpgo -s mysite cron event list
wpgo -s mysite cron event run
```

### rewrite -- Rewrite rules management

```bash
wpgo -s mysite rewrite flush
wpgo -s mysite rewrite list
```

### comment -- Comment management

```bash
wpgo -s mysite comment list
wpgo -s mysite comment approve 15
wpgo -s mysite comment delete 15
```

### menu -- Menu management

```bash
wpgo -s mysite menu list
```

### config -- wp-config.php management

```bash
wpgo -s mysite config get DB_NAME
wpgo -s mysite config set WP_DEBUG true
```

### role -- Role management

```bash
wpgo -s mysite role list
```

### maintenance -- Maintenance mode

```bash
wpgo -s mysite maintenance enable
wpgo -s mysite maintenance disable
```

### eval -- Execute arbitrary PHP

```bash
wpgo -s mysite eval "echo get_option('siteurl');"
```

### raw -- Pass-through to wp-cli

```bash
wpgo -s mysite raw "wp option list"
```

### Shortcut Commands

```bash
wpgo -s mysite health         # Full site health check
wpgo -s mysite status         # Quick site status overview
wpgo -s mysite backup         # Database backup
wpgo -s mysite backup "Pre-update snapshot"  # Backup with description
wpgo -s mysite update-all -y  # Update core + plugins + themes
wpgo -s mysite clear-cache    # Full cache clear
```

### setup and help

```bash
wpgo setup                    # Interactive onboarding
wpgo help                     # Guided help by topic
wpgo version                  # Show version
```

## Global Flags

| Flag | Description |
|------|-------------|
| `-s`, `--site` | Target site alias |
| `--sites` | Multiple target sites (comma-separated, batch mode) |
| `-g`, `--group` | Target site group |
| `--json` | Output as JSON |
| `--plain` | Output as plain text |
| `-v`, `--verbose` | Verbose output |
| `--dry-run` | Show commands without executing |
| `--no-cache` | Bypass cache |
| `--fields` | Comma-separated fields to display |
| `-y`, `--yes` | Skip confirmation prompts |
| `--ack-destructive` | Acknowledge destructive batch operations |
| `--concurrency` | Max parallel executions in batch mode (default: 1) |

## License

MIT
