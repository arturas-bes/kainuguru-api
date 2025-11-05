#!/bin/bash

# Kainuguru API Database Backup Script
# Usage: ./scripts/backup.sh [environment]

set -e

ENVIRONMENT="${1:-development}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${SCRIPT_DIR}/../backups"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")

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

BACKUP_FILE="${BACKUP_DIR}/${ENVIRONMENT}_kainuguru_${TIMESTAMP}.sql"
BACKUP_FILE_COMPRESSED="${BACKUP_FILE}.gz"

echo "=================================="
echo "  Kainuguru Database Backup"
echo "=================================="
echo "Environment: $ENVIRONMENT"
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "User: $DB_USER"
echo "Backup file: $BACKUP_FILE_COMPRESSED"
echo "=================================="

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Check if PostgreSQL tools are available
if ! command -v pg_dump &> /dev/null; then
    echo "❌ Error: pg_dump is not installed"
    echo "Please install PostgreSQL client tools"
    exit 1
fi

# Perform the backup
echo "📦 Creating database backup..."
if pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
   --verbose --clean --if-exists --create \
   --format=plain --no-owner --no-privileges > "$BACKUP_FILE"; then

    echo "✅ Database backup created successfully"

    # Compress the backup
    echo "🗜️  Compressing backup..."
    if gzip "$BACKUP_FILE"; then
        echo "✅ Backup compressed successfully"

        # Display backup information
        BACKUP_SIZE=$(du -h "$BACKUP_FILE_COMPRESSED" | cut -f1)
        echo "📊 Backup size: $BACKUP_SIZE"

        # Cleanup old backups (keep last 10)
        echo "🧹 Cleaning up old backups..."
        find "$BACKUP_DIR" -name "${ENVIRONMENT}_kainuguru_*.sql.gz" -type f | \
        sort -r | tail -n +11 | xargs -r rm -f

        echo "✅ Backup completed successfully: $BACKUP_FILE_COMPRESSED"

        # Optional: Upload to cloud storage
        if [ ! -z "$AWS_S3_BACKUP_BUCKET" ]; then
            echo "☁️  Uploading to AWS S3..."
            if aws s3 cp "$BACKUP_FILE_COMPRESSED" "s3://$AWS_S3_BACKUP_BUCKET/database-backups/"; then
                echo "✅ Backup uploaded to S3"
            else
                echo "⚠️  Failed to upload to S3"
            fi
        fi

    else
        echo "❌ Failed to compress backup"
        exit 1
    fi
else
    echo "❌ Database backup failed"
    exit 1
fi

echo "🎉 Backup process completed!"