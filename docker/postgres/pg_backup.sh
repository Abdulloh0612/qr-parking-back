#!/bin/bash

# Configuration
CONTAINER_NAME=qr-parking-database
DB_NAME=qr-parking
DB_USER=qr-parking
BACKUP_DIR=/root/backups
DATE=$(date +%Y-%m-%d_%H-%M-%S)
SQL_FILE="${DB_NAME}_$DATE.sql"
TAR_FILE="${SQL_FILE}.tar.gz"
BACKUP_PATH="$BACKUP_DIR/$TAR_FILE"

# Create backup directory if not exists
mkdir -p "$BACKUP_DIR"

# Dump the database to a temporary .sql file
docker exec "$CONTAINER_NAME" pg_dump -U "$DB_USER" "$DB_NAME" \
  --exclude-table=logs --exclude-table=offer_analytics > "$BACKUP_DIR/$SQL_FILE"

# Compress the .sql file to .tar.gz
tar -czf "$BACKUP_PATH" -C "$BACKUP_DIR" "$SQL_FILE"

# Remove the raw .sql file after compression
rm "$BACKUP_DIR/$SQL_FILE"

# Optional: Remove backups older than 3 days
find "$BACKUP_DIR" -type f -name "*.tar.gz" -mtime +3 -exec rm {} \;
