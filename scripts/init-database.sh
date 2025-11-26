#!/bin/bash

# Database initialization script for File Hub
# This script will:
# 1. Prompt for database connection details
# 2. Create/replace the database user for File Hub
# 3. Create the database if it doesn't exist
# 4. Apply the database schema
# 5. Generate a config.yaml file with the provided details

set -e  # Exit on any error

echo "=== File Hub Database Initialization Script ==="
echo ""

# Default values
DEFAULT_DB_HOST="localhost"
DEFAULT_DB_PORT="5432"
DEFAULT_DB_USER="filehub"
DEFAULT_DB_PASSWORD="filehub"
DEFAULT_DB_NAME="filehub"
DEFAULT_CONFIG_FILE="config.yaml"

# Prompt user for database connection details with defaults
read -p "Database host [${DEFAULT_DB_HOST}]: " DB_HOST
DB_HOST=${DB_HOST:-$DEFAULT_DB_HOST}

read -p "Database port [${DEFAULT_DB_PORT}]: " DB_PORT
DB_PORT=${DB_PORT:-$DEFAULT_DB_PORT}

# Always create/replace the database user
echo ""
echo "Database user '$DEFAULT_DB_USER' will be created/replaced for File Hub."

read -s -p "Database password for File Hub user '$DEFAULT_DB_USER' (will be hidden): " DB_PASSWORD
echo ""  # New line after password input

read -p "Database name [${DEFAULT_DB_NAME}]: " DB_NAME
DB_NAME=${DB_NAME:-$DEFAULT_DB_NAME}

read -p "Config file to generate [${DEFAULT_CONFIG_FILE}]: " CONFIG_FILE
CONFIG_FILE=${CONFIG_FILE:-$DEFAULT_CONFIG_FILE}

# Set the database user to the default
DB_USER=$DEFAULT_DB_USER

echo ""
echo "Summary of configuration:"
echo "  Database host: ${DB_HOST}"
echo "  Database port: ${DB_PORT}"
echo "  Database user: ${DB_USER}"
echo "  Database name: ${DB_NAME}"
echo "  Config file: ${CONFIG_FILE}"
echo "  Database URI: postgresql://${DB_USER}:****@${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo ""

read -p "Proceed with initialization? (y/N): " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Initialization cancelled."
    exit 0
fi

echo ""
echo "Setting up database with sudo privileges..."

# Create database and user using sudo instead of prompting for superuser credentials
echo "Creating database user '$DB_USER' and database '$DB_NAME'..."
sudo -u postgres psql <<EOF
-- Create or replace the database user
DROP USER IF EXISTS $DB_USER;
CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';

-- Create the database if it doesn't exist
SELECT 'CREATE DATABASE $DB_NAME WITH OWNER $DB_USER'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec

-- Grant all privileges on database to the user
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
EOF
if [ $? -ne 0 ]; then
    echo "ERROR: Failed to create database or user. Please check your PostgreSQL setup and try again."
    exit 1
fi
echo "Database and user created successfully!"
echo ""

echo "Applying database schema..."
# Connect to the new database and apply schema
export PGPASSWORD="$DB_PASSWORD"
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "./scripts/database_schema.sql"

if [ $? -ne 0 ]; then
    echo "ERROR: Failed to apply database schema."
    exit 1
fi

echo "Database schema applied successfully!"
echo ""

echo "Generating configuration file: $CONFIG_FILE"
# Create database URI: postgresql://user:password@host:port/database
DB_URI="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
cat > "$CONFIG_FILE" <<EOF
web:
  port: 8080

database:
  uri: "$DB_URI"
EOF

echo "Configuration file generated successfully!"
echo ""

echo "=== Initialization Complete ==="
echo "1. Database '$DB_NAME' created and initialized"
echo "2. Database user '$DB_USER' created with necessary privileges"
echo "3. Configuration saved to '$CONFIG_FILE' with database URI"
echo ""
echo "To run the application:"
echo "  ./bin/file-hub"
echo ""
echo "Note: Make sure to set proper permissions on your config file if it contains sensitive information"
echo ""