#!/bin/sh
set -e

echo "Running database migrations..."
PGPASSWORD="$POSTGRES_PASSWORD" psql \
    -h "$POSTGRES_HOST" \
    -p "$POSTGRES_PORT" \
    -U "$POSTGRES_USER" \
    -d "$POSTGRES_DB" \
    -f /app/migrations/001_init.sql

echo "Starting server..."
exec ./server
