#!/bin/bash
export PATH="/opt/homebrew/bin:$PATH"

# Remote MySQL credentials
REMOTE_HOST="lucas-remigio-dev.pt"  
REMOTE_PORT=""                 
REMOTE_USER=""
REMOTE_PASS=""
DB_NAME=""

# Local backup directory and file naming
LOCAL_BACKUP_DIR="/Users/lucasremigio/Developer/wallet_tracket_backup/"
DATE=$(date +'%d-%m-%Y_%H-%M-%S')
DUMP_FILE="${LOCAL_BACKUP_DIR}/${DATE}_${DB_NAME}.sql"

# Create local backup directory if it doesn't exist
mkdir -p "$LOCAL_BACKUP_DIR"

# Dump the remote MySQL database to a local file
mysqldump -h "$REMOTE_HOST" -P "$REMOTE_PORT" -u "$REMOTE_USER" -p"$REMOTE_PASS" "$DB_NAME" > "$DUMP_FILE"

# Check if the dump succeeded
if [ $? -eq 0 ]; then
  echo "Database dump successful: $DUMP_FILE"
else
  echo "Error during database dump" >&2
  exit 1
fi
