#!/bin/bash
set -e

# Change ownership of the data directory
chown -R 999:999 /var/lib/postgresql/data-replics

# Run the original PostgreSQL entrypoint
exec /usr/local/bin/docker-entrypoint.sh "$@"
