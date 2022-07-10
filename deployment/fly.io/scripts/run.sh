#!/bin/bash
set -e

# Restore the database if it does not already exist.
if [ -f /data/castellers.db ]; then
	echo "Database already exists, skipping restore"
else
	echo "No database found, restoring from replica if exists"
	litestream restore -v -if-replica-exists -o /data/castellers.db "${REPLICA_URL}"
fi

# Run litestream with your app as the subprocess.
exec litestream replicate -exec "/app -dsn /data/castellers.db"