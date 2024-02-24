#!/bin/bash

# runs at the root of the /app 

# Define the migration directory
MIGRATION_DIR="/app/sql/schema"

# Create the migration directory if it doesn't exist
mkdir -p "$MIGRATION_DIR"

# Define the database URL
DATABASE_URL="postgres://postgres:mysecretpassword@postgres:5432/mynewdatabase?sslmode=disable"

# Wait for the PostgreSQL container to be ready
echo "Waiting for PostgreSQL to start..."
while ! pg_isready -h postgres -p 5432 -q -U postgres; do
  sleep 1
done
echo "PostgreSQL started"

# Wait for the database to be available
echo "Waiting for database to be available..."
until PGPASSWORD=mysecretpassword psql -h "postgres" -U "postgres" -c '\l' | grep "mynewdatabase" > /dev/null; do
  sleep 1
done
echo "Database is available"

# Run Goose migrations
echo "Running Goose migrations..."
goose -dir "$MIGRATION_DIR" postgres "$DATABASE_URL" up

# Check if Goose migrations were successful
if [ $? -ne 0 ]; then
  echo "Goose migrations failed"
  exit 1
else
  echo "Goose migrations applied successfully"
fi