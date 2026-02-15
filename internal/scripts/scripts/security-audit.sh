#!/bin/bash
# security-audit.sh — Security audit for wpgo.
# Executed via SSH stdin: bash -s
# Output: JSON with audit results.

json_escape() {
    printf '%s' "$1" | sed 's/\\/\\\\/g; s/"/\\"/g; s/\t/\\t/g; s/\n/\\n/g'
}

# Core checksum verification.
checksums_output=$(wp core verify-checksums 2>&1)
if [ $? -eq 0 ]; then
    checksums_status="ok"
else
    checksums_status="failed"
fi

# Admin users.
admin_users=$(wp user list --role=administrator --fields=ID,user_login,user_email --format=json 2>/dev/null || echo "[]")

# File permissions.
wpconfig_perms="not_found"
if [ -f wp-config.php ]; then
    wpconfig_perms=$(stat -c '%a' wp-config.php 2>/dev/null || stat -f '%Lp' wp-config.php 2>/dev/null || echo "unknown")
fi

uploads_perms="not_found"
if [ -d wp-content/uploads ]; then
    uploads_perms=$(stat -c '%a' wp-content/uploads 2>/dev/null || stat -f '%Lp' wp-content/uploads 2>/dev/null || echo "unknown")
fi

# WP_DEBUG status.
wp_debug=$(wp config get WP_DEBUG 2>/dev/null || echo "not_set")
wp_debug_log=$(wp config get WP_DEBUG_LOG 2>/dev/null || echo "not_set")
wp_debug_display=$(wp config get WP_DEBUG_DISPLAY 2>/dev/null || echo "not_set")

# Table prefix (non-default = better).
db_prefix=$(wp db prefix 2>/dev/null || echo "unknown")

cat <<ENDJSON
{
  "checksums": {
    "status": "$checksums_status",
    "details": "$(json_escape "$checksums_output")"
  },
  "admin_users": $admin_users,
  "file_permissions": {
    "wp_config": "$wpconfig_perms",
    "uploads_dir": "$uploads_perms"
  },
  "debug": {
    "wp_debug": "$(json_escape "$wp_debug")",
    "wp_debug_log": "$(json_escape "$wp_debug_log")",
    "wp_debug_display": "$(json_escape "$wp_debug_display")"
  },
  "db_prefix": "$(json_escape "$db_prefix")"
}
ENDJSON
