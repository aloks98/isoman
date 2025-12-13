#!/bin/sh
set -e

# Create directories if they don't exist
mkdir -p /data/isos /data/db

# Fix ownership - only if needed and don't follow symlinks
# Only change ownership of the directories themselves, not recursively
chown isoman:isoman /data /data/isos /data/db

# Switch to isoman user and execute the main command
exec su-exec isoman "$@"
