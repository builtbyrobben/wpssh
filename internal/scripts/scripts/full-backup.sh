#!/bin/bash
# full-backup.sh — Named database backup for wpgo.
# Executed via SSH stdin: bash -s -- <client_name> <description>
# Output: JSON with status, filename, path.

set -e

CLIENT="${1:-Site}"
DESC="${2:-Manual}"
DATE=$(date +%Y%m%d_%H%M%S)

# Sanitize inputs: replace spaces/special chars with underscores.
CLIENT_SAFE=$(echo "$CLIENT" | sed 's/[^a-zA-Z0-9_-]/_/g')
DESC_SAFE=$(echo "$DESC" | sed 's/[^a-zA-Z0-9_-]/_/g')

FILENAME="${CLIENT_SAFE}_DB_${DESC_SAFE}_${DATE}.sql"

# Export to a temp location within the WordPress directory.
EXPORT_DIR="."
EXPORT_PATH="${EXPORT_DIR}/${FILENAME}"

if wp db export "$EXPORT_PATH" 2>/dev/null; then
    FILESIZE=$(du -h "$EXPORT_PATH" 2>/dev/null | cut -f1 || echo "unknown")
    FULLPATH=$(realpath "$EXPORT_PATH" 2>/dev/null || echo "$EXPORT_PATH")

    cat <<ENDJSON
{
  "status": "ok",
  "filename": "$FILENAME",
  "path": "$FULLPATH",
  "size": "$FILESIZE"
}
ENDJSON
else
    cat <<ENDJSON
{
  "status": "error",
  "filename": "$FILENAME",
  "path": "",
  "size": "",
  "error": "wp db export failed"
}
ENDJSON
    exit 1
fi
