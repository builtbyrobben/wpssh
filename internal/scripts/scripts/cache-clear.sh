#!/bin/bash
# cache-clear.sh — Full cache clear for wpgo.
# Executed via SSH stdin: bash -s
# Clears object cache, transients, rewrite rules, and plugin caches.
# Output: JSON with each step result.

results=""
add_result() {
    local step="$1"
    local status="$2"
    if [ -n "$results" ]; then
        results="${results},"
    fi
    results="${results}{\"step\":\"${step}\",\"status\":\"${status}\"}"
}

# Step 1: Flush object cache.
if wp cache flush 2>/dev/null; then
    add_result "cache_flush" "ok"
else
    add_result "cache_flush" "failed"
fi

# Step 2: Delete all transients.
if wp transient delete --all 2>/dev/null; then
    add_result "transient_delete" "ok"
else
    add_result "transient_delete" "failed"
fi

# Step 3: Flush rewrite rules.
if wp rewrite flush 2>/dev/null; then
    add_result "rewrite_flush" "ok"
else
    add_result "rewrite_flush" "failed"
fi

# Step 4: LiteSpeed Cache (if active).
if wp plugin is-active litespeed-cache 2>/dev/null; then
    if wp litespeed-purge all 2>/dev/null; then
        add_result "litespeed_purge" "ok"
    else
        add_result "litespeed_purge" "failed"
    fi
else
    add_result "litespeed_purge" "skipped"
fi

# Step 5: WP Rocket (if active).
if wp plugin is-active wp-rocket 2>/dev/null; then
    if wp rocket clean --confirm 2>/dev/null; then
        add_result "wp_rocket_clean" "ok"
    else
        add_result "wp_rocket_clean" "failed"
    fi
else
    add_result "wp_rocket_clean" "skipped"
fi

cat <<ENDJSON
{
  "status": "ok",
  "steps": [$results]
}
ENDJSON
