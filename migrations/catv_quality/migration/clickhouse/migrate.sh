#!/bin/bash

# ClickHouse connection settings
CH_HOST="192.168.7.18"
CH_PORT="30123"
#CH_USER="default"
#CH_PASS="default"
CH_DB="catv"

# Migration directory
# Get the directory where the script is located
MIGRATION_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Goose command
# Using http driver for ClickHouse migration
GOOSE_DBSTRING="http://${CH_USER}:${CH_PASS}@${CH_HOST}:${CH_PORT}/${CH_DB}"

echo "Starting ClickHouse v3 migrations..."
goose -dir "${MIGRATION_DIR}" clickhouse "${GOOSE_DBSTRING}" up

if [ $? -eq 0 ]; then
    echo "Migration completed successfully."
else
    echo "Migration failed."
    exit 1
fi
