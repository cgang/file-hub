#!/bin/bash

# Database initialization script for File Hub
# This script will:
# 1. Prompt for database connection details
# 2. Optionally create the database and user if they don't exist
# 3. Apply the database schema
# 4. Generate a config.yaml file with the provided details

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

# Ask user if they want to create a new database user or use an existing one
echo ""
echo "Do you want to create a new database user for File Hub or use an existing one?"
PS3="Please select an option (1 or 2): "
options=("Create a new database user" "Use an existing database user")
select opt in "${options[@]}"
do
    case $opt in
        "Create a new database user")
            CREATE_NEW_USER=true
            break
            ;;
        "Use an existing database user")
            CREATE_NEW_USER=false
            break
            ;;
        *) echo "Invalid option. Please select 1 or 2.";;
    esac
done

if [ "$CREATE_NEW_USER" = true ]; then
    echo ""
    echo "Creating a new database user requires superuser privileges."
    read -p "Database superuser [postgres]: " DB_SUPERUSER
    DB_SUPERUSER=${DB_SUPERUSER:-postgres}

    read -s -p "Database superuser password (will be hidden): " DB_SUPERUSER_PASSWORD
    echo ""  # New line after password input

    read -p "Database user to create for File Hub [${DEFAULT_DB_USER}]: " DB_USER
    DB_USER=${DB_USER:-$DEFAULT_DB_USER}
else
    echo ""
    echo "Using an existing database user. You'll need to provide the credentials."
    read -p "Existing database user for File Hub [${DEFAULT_DB_USER}]: " DB_USER
    DB_USER=${DB_USER:-$DEFAULT_DB_USER}

    # For existing users, we still need the superuser for database creation but not for user creation
    read -p "Database superuser [postgres]: " DB_SUPERUSER
    DB_SUPERUSER=${DB_SUPERUSER:-postgres}

    read -s -p "Database superuser password (will be hidden): " DB_SUPERUSER_PASSWORD
    echo ""  # New line after password input
fi

read -s -p "Database password for File Hub user '$DB_USER' (will be hidden): " DB_PASSWORD
echo ""  # New line after password input

read -p "Database name [${DEFAULT_DB_NAME}]: " DB_NAME
DB_NAME=${DB_NAME:-$DEFAULT_DB_NAME}

read -p "Config file to generate [${DEFAULT_CONFIG_FILE}]: " CONFIG_FILE
CONFIG_FILE=${CONFIG_FILE:-$DEFAULT_CONFIG_FILE}

echo ""
echo "Summary of configuration:"
echo "  Database host: ${DB_HOST}"
echo "  Database port: ${DB_PORT}"
echo "  Database superuser: ${DB_SUPERUSER}"
echo "  Database user: ${DB_USER}"
if [ "$CREATE_NEW_USER" = true ]; then
    echo "  Action: Creating new database user"
else
    echo "  Action: Using existing database user"
fi
echo "  Database name: ${DB_NAME}"
echo "  Config file: ${CONFIG_FILE}"
echo ""

read -p "Proceed with initialization? (y/N): " CONFIRM
if [[ ! "$CONFIRM" =~ ^[Yy]$ ]]; then
    echo "Initialization cancelled."
    exit 0
fi

echo ""
echo "Connecting to PostgreSQL as superuser to set up database..."

# Create database (and user if needed)
export PGPASSWORD="$DB_SUPERUSER_PASSWORD"
if [ "$CREATE_NEW_USER" = true ]; then
    echo "Creating new database user '$DB_USER' and database '$DB_NAME'..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_SUPERUSER" -d postgres <<EOF
-- Create the database user if it doesn't exist
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
    CREATE USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
  ELSE
    ALTER USER $DB_USER WITH PASSWORD '$DB_PASSWORD';
  END IF;
END
\$\$;

-- Create the database if it doesn't exist
SELECT 'CREATE DATABASE $DB_NAME WITH OWNER $DB_USER'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec

-- Grant all privileges on database to the user
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
EOF
    if [ $? -ne 0 ]; then
        echo "ERROR: Failed to create database or user. Please check your credentials and try again."
        exit 1
    fi
    echo "Database and user created successfully!"
else
    echo "Setting up database '$DB_NAME' for existing user '$DB_USER'..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_SUPERUSER" -d postgres <<EOF
-- Check if the user exists
DO \$\$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = '$DB_USER') THEN
    RAISE EXCEPTION 'User $DB_USER does not exist. Please create the user first or choose to create a new user.';
  END IF;
END
\$\$;

-- Create the database if it doesn't exist
SELECT 'CREATE DATABASE $DB_NAME WITH OWNER $DB_USER'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$DB_NAME')\gexec

-- Grant all privileges on database to the user
GRANT ALL PRIVILEGES ON DATABASE $DB_NAME TO $DB_USER;
EOF
    if [ $? -ne 0 ]; then
        echo "ERROR: Failed to set up database for existing user. Please check your credentials and try again."
        exit 1
    fi
    echo "Database set up successfully for existing user '$DB_USER'!"
fi
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
cat > "$CONFIG_FILE" <<EOF
storage:
  root_dir: "root"

web:
  port: 8080

database:
  host: "$DB_HOST"
  port: $DB_PORT
  user: "$DB_USER"
  password: "$DB_PASSWORD"
  database: "$DB_NAME"
EOF

echo "Configuration file generated successfully!"
echo ""

echo "=== Initialization Complete ==="
echo "1. Database '$DB_NAME' created and initialized"
if [ "$CREATE_NEW_USER" = true ]; then
    echo "2. Database user '$DB_USER' created with necessary privileges"
else
    echo "2. Database configured for existing user '$DB_USER'"
fi
echo "3. Configuration saved to '$CONFIG_FILE'"
echo ""
echo "To run the application:"
echo "  ./bin/file-hub"
echo ""
echo "Note: Make sure to set proper permissions on your config file if it contains sensitive information"
echo ""