#!/bin/bash
set -e

# Restore the database if it does not already exist.
if [ -f ${APP_DB_NAME} ]; then
	echo "Database already exists, skipping restore"
else
	echo "No database found, restoring from replica if exists"
	litestream restore -v -if-replica-exists -o ${APP_DB_NAME} "${REPLICA_URL}"
fi

# Run litestream with your app as the subprocess.
exec litestream replicate -exec "/app -dsn ${APP_DB_NAME}"