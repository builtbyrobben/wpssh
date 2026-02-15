#!/bin/bash
# health-check.sh — Full site health check for wpgo.
# Executed via SSH stdin: bash -s
# Output: JSON object with all health metrics.

set -e

json_escape() {
    printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g; s/\t/\\t/g; s/\n/\\n/g'
}

# Core version
core_version=$(wp core version 2>/dev/null || echo "unknown")

# Core update check
core_update=$(wp core check-update --format=json 2>/dev/null || echo "[]")

# Plugin summary
plugin_count=$(wp plugin list --format=count 2>/dev/null || echo "0")
plugin_update_count=$(wp plugin list --update=available --format=count 2>/dev/null || echo "0")

# Theme summary
theme_count=$(wp theme list --format=count 2>/dev/null || echo "0")
active_theme=$(wp theme list --status=active --field=name 2>/dev/null || echo "unknown")

# Database
db_size=$(wp db size --size_format=mb 2>/dev/null | tail -1 || echo "unknown")
table_count=$(wp db tables --format=count 2>/dev/null || echo "0")
db_prefix=$(wp db prefix 2>/dev/null || echo "unknown")

# Admin users
admin_count=$(wp user list --role=administrator --format=count 2>/dev/null || echo "0")
admin_list=$(wp user list --role=administrator --fields=ID,user_login,user_email --format=json 2>/dev/null || echo "[]")

# Cron
cron_test="ok"
wp cron test >/dev/null 2>&1 || cron_test="failed"

# Site URL and Home URL
site_url=$(wp option get siteurl 2>/dev/null || echo "unknown")
home_url=$(wp option get home 2>/dev/null || echo "unknown")

# PHP version
php_version=$(wp eval "echo phpversion();" 2>/dev/null || echo "unknown")

# Disk usage (WordPress directory)
disk_usage=$(du -sh . 2>/dev/null | cut -f1 || echo "unknown")

cat <<ENDJSON
{
  "core_version": "$(json_escape "$core_version")",
  "core_updates": $core_update,
  "plugin_count": $plugin_count,
  "plugin_updates_available": $plugin_update_count,
  "theme_count": $theme_count,
  "active_theme": "$(json_escape "$active_theme")",
  "db_size": "$(json_escape "$db_size")",
  "db_table_count": $table_count,
  "db_prefix": "$(json_escape "$db_prefix")",
  "admin_count": $admin_count,
  "admin_users": $admin_list,
  "cron_status": "$cron_test",
  "site_url": "$(json_escape "$site_url")",
  "home_url": "$(json_escape "$home_url")",
  "php_version": "$(json_escape "$php_version")",
  "disk_usage": "$(json_escape "$disk_usage")"
}
ENDJSON
