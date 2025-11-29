-- File Hub MVP Database Schema
-- Simplified PostgreSQL schema for user management and basic file metadata

-- Users table for authentication and user information
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    ha1_hash VARCHAR(255) NOT NULL,  -- Store HA1 hash for digest auth (username:realm:password)
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    home_dir VARCHAR(255) NOT NULL,  -- Home directory path for the user
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE
);

-- File metadata table to track files on the filesystem
CREATE TABLE files (
    id SERIAL PRIMARY KEY,                           -- Sequential ID for better performance
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    path TEXT NOT NULL,                             -- Full path including filename relative to WebDAV root
    mime_type VARCHAR(255),
    size BIGINT,                                    -- File size in bytes
    checksum VARCHAR(64),                           -- SHA-256 hash of file content
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_deleted BOOLEAN DEFAULT FALSE,                -- For soft delete
    deleted_at TIMESTAMP WITH TIME ZONE              -- Timestamp when marked as deleted
);

-- Create indexes for files table
CREATE INDEX idx_files_user_id ON files (user_id);
CREATE INDEX idx_files_path ON files (path);

-- Quota management for users
CREATE TABLE user_quota (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    total_quota_bytes BIGINT NOT NULL DEFAULT 10737418240, -- 10GB default
    used_bytes BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Trigger to update user quota when files are added/removed
CREATE OR REPLACE FUNCTION update_user_quota()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        -- When a file is added, increase used quota
        UPDATE user_quota
        SET used_bytes = used_bytes + COALESCE(NEW.size, 0),
            updated_at = CURRENT_TIMESTAMP
        WHERE user_id = NEW.user_id;
        RETURN NEW;
    ELSIF (TG_OP = 'UPDATE') THEN
        -- When a file is updated (size changed), adjust used quota
        UPDATE user_quota
        SET used_bytes = used_bytes - COALESCE(OLD.size, 0) + COALESCE(NEW.size, 0),
            updated_at = CURRENT_TIMESTAMP
        WHERE user_id = NEW.user_id;
        RETURN NEW;
    ELSIF (TG_OP = 'DELETE') THEN
        -- When a file is removed, decrease used quota
        UPDATE user_quota
        SET used_bytes = used_bytes - COALESCE(OLD.size, 0),
            updated_at = CURRENT_TIMESTAMP
        WHERE user_id = OLD.user_id;
        RETURN OLD;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER user_quota_trigger
    AFTER INSERT OR UPDATE OR DELETE ON files
    FOR EACH ROW EXECUTE FUNCTION update_user_quota();

-- Comments for documentation
COMMENT ON TABLE users IS 'User accounts and authentication information';
COMMENT ON TABLE files IS 'Metadata for files stored on the filesystem (directories implicit in path)';
COMMENT ON TABLE user_quota IS 'Storage quota management for users';

-- Index comments
COMMENT ON INDEX idx_files_user_id IS 'Index on user_id column for faster file lookups by user';
COMMENT ON INDEX idx_files_path IS 'Index on path column for faster file lookups by path';

-- Relations documentation
/*
Relationships summary:

users table is the central identity table
  - files table references users via user_id (many-to-one)
  - user_quota table references users via user_id (one-to-one)

files table stores metadata about files
  - directories are implicit in the file path structure
*/