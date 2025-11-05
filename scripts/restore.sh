#!/bin/bash

# Kainuguru API Database Restore Script
# Usage: ./scripts/restore.sh <backup_file> [environment]

set -e

BACKUP_FILE="$1"
ENVIRONMENT="${2:-development}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -z "$BACKUP_FILE" ]; then
    echo "❌ Error: Backup file not specified"
    echo "Usage: $0 <backup_file> [environment]"
    echo ""
    echo "Available backups:"
    find "${SCRIPT_DIR}/../backups" -name "*.sql.gz" -type f 2>/dev/null | sort -r | head -10
    exit 1
fi

# Load environment specific configuration
case $ENVIRONMENT in
  "production")
    DB_HOST="${PROD_DB_HOST:-localhost}"
    DB_PORT="${PROD_DB_PORT:-5432}"
    DB_NAME="${PROD_DB_NAME:-kainuguru}"
    DB_USER="${PROD_DB_USER:-kainuguru}"
    ;;
  "staging")
    DB_HOST="${STAGING_DB_HOST:-localhost}"
    DB_PORT="${STAGING_DB_PORT:-5432}"
    DB_NAME="${STAGING_DB_NAME:-kainuguru_staging}"
    DB_USER="${STAGING_DB_USER:-kainuguru}"
    ;;
  *)
    DB_HOST="${DB_HOST:-localhost}"
    DB_PORT="${DB_PORT:-5432}"
    DB_NAME="${DB_NAME:-kainuguru}"
    DB_USER="${DB_USER:-kainuguru}"
    ;;
esac

echo "=================================="
echo "  Kainuguru Database Restore"
echo "=================================="
echo "Environment: $ENVIRONMENT"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo "Backup file: $BACKUP_FILE"
echo "=================================="

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo "❌ Error: Backup file not found: $BACKUP_FILE"
    exit 1
fi

# Check if PostgreSQL tools are available
if ! command -v psql &> /dev/null; then
    echo "❌ Error: psql is not installed"
    echo "Please install PostgreSQL client tools"
    exit 1
fi

# Confirmation prompt for production
if [ "$ENVIRONMENT" = "production" ]; then
    echo "⚠️  WARNING: You are about to restore to PRODUCTION database!"
    echo "This will OVERWRITE all existing data in $DB_NAME"
    read -p "Are you sure you want to continue? (type 'YES' to confirm): " confirm
    if [ "$confirm" != "YES" ]; then
        echo "❌ Restore cancelled"
        exit 1
    fi
fi

# Extract backup if compressed
TEMP_SQL_FILE=""
if [[ "$BACKUP_FILE" == *.gz ]]; then
    echo "📦 Extracting compressed backup..."
    TEMP_SQL_FILE="${BACKUP_FILE%.gz}"
    if gunzip -c "$BACKUP_FILE" > "$TEMP_SQL_FILE"; then
        echo "✅ Backup extracted successfully"
        SQL_FILE="$TEMP_SQL_FILE"
    else
        echo "❌ Failed to extract backup"
        exit 1
    fi
else
    SQL_FILE="$BACKUP_FILE"
fi

# Perform the restore
echo "🔄 Restoring database..."
if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres < "$SQL_FILE"; then
    echo "✅ Database restored successfully"

    # Cleanup temporary file
    if [ ! -z "$TEMP_SQL_FILE" ] && [ -f "$TEMP_SQL_FILE" ]; then
        rm -f "$TEMP_SQL_FILE"
        echo "🧹 Temporary files cleaned up"
    fi

    echo "🎉 Restore process completed!"
else
    echo "❌ Database restore failed"

    # Cleanup temporary file on failure
    if [ ! -z "$TEMP_SQL_FILE" ] && [ -f "$TEMP_SQL_FILE" ]; then
        rm -f "$TEMP_SQL_FILE"
    fi

    exit 1
fi