# File Hub Scripts

This directory contains various scripts for the File Hub project.

## Contents

- `database_schema.sql` - Database schema for the File Hub service
  - Contains the PostgreSQL schema for users, files, and quotas tables
  - Includes indexes, triggers, and functions for quota management
  - To apply: `psql -d filehub -f database_schema.sql`