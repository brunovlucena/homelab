#!/bin/bash

# Bruno Site Database Migration Script
# This script runs database migrations for the Bruno site project

set -e

echo "🗄️ Running Bruno Site database migrations..."

# Check if we're running in Kubernetes environment
if [ -n "$KUBERNETES_SERVICE_HOST" ]; then
    echo "🔗 Detected Kubernetes environment"
    # In Kubernetes, use the service name
    DB_HOST="bruno-site-postgres"
    DB_PORT="5432"
else
    echo "🏠 Detected local environment"
    # In local environment, set up port forwarding to Kubernetes PostgreSQL
    echo "🔗 Setting up port forwarding to Kubernetes PostgreSQL..."
    kubectl port-forward --address 0.0.0.0 -n bruno svc/bruno-site-postgres 5432:5432 &
    PF_PID=$!
    echo "⏳ Waiting for port forwarding to establish..."
    sleep 3
    
    # Set trap to cleanup port forwarding on script exit
    trap 'echo "🛑 Cleaning up port forwarding..."; kill $PF_PID 2>/dev/null || true' EXIT
    
    DB_HOST="localhost"
    DB_PORT="5432"
fi

# Database configuration - all must be set via environment variables
if [ -z "$DB_NAME" ]; then
    echo "❌ Error: DB_NAME environment variable is required"
    exit 1
fi

if [ -z "$DB_USER" ]; then
    echo "❌ Error: DB_USER environment variable is required"
    exit 1
fi

if [ -z "$DB_PASSWORD" ]; then
    echo "❌ Error: DB_PASSWORD environment variable is required"
    exit 1
fi

MIGRATION_FILE="api/migrations/001_complete_schema.sql"

echo "📋 Database Configuration:"
echo "  Host: $DB_HOST"
echo "  Port: $DB_PORT"
echo "  Database: $DB_NAME"
echo "  User: $DB_USER"
echo "  Migration File: $MIGRATION_FILE"

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
until PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c '\q' 2>/dev/null; do
    echo "  Database not ready yet, waiting..."
    sleep 2
done
echo "✅ Database is ready!"

# Run the migration
echo "🚀 Running migration..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $MIGRATION_FILE

echo "✅ Migration completed successfully!"
