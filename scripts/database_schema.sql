-- File Hub Database Schema
-- PostgreSQL schema for user management, repositories, file metadata, shares and quota management

-- Users table for authentication and user information
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    ha1_hash VARCHAR(255) NOT NULL,  -- Store HA1 hash for digest auth (username:realm:password)
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE
);

-- Repositories table representing file repositories owned by users
CREATE TABLE repositories (
    id SERIAL PRIMARY KEY,
    owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    root_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- File metadata table to track files and directories in repositories
CREATE TABLE files (
    id SERIAL PRIMARY KEY,
    parent_id INTEGER,  -- NULL for root level files/directories
    owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_id INTEGER NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,  -- File or directory name
    path TEXT NOT NULL,          -- Full path including filename relative to repository root
    mime_type VARCHAR(255),
    size BIGINT NOT NULL DEFAULT 0,  -- File size in bytes
    checksum VARCHAR(64),            -- SHA-256 hash of file content
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_dir BOOLEAN NOT NULL DEFAULT FALSE,  -- True for directories, false for files
    deleted BOOLEAN NOT NULL DEFAULT FALSE   -- Soft delete flag
);

-- Shares table for sharing repository paths with other users
CREATE TABLE shares (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    path TEXT NOT NULL  -- Path within the repository being shared
);

-- Quota management for users
CREATE TABLE user_quota (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    total_quota_bytes BIGINT NOT NULL DEFAULT 10737418240, -- 10GB default
    used_bytes BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for better query performance
CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_users_email ON users (email);
CREATE INDEX idx_repositories_owner_id ON repositories (owner_id);
CREATE INDEX idx_repositories_name ON repositories (name);
CREATE INDEX idx_files_owner_id ON files (owner_id);
CREATE INDEX idx_files_repo_id ON files (repo_id);
CREATE INDEX idx_files_path ON files (path);
CREATE INDEX idx_files_parent_id ON files (parent_id);
CREATE INDEX idx_shares_user_id ON shares (user_id);
CREATE INDEX idx_shares_repo_id ON shares (repo_id);
CREATE INDEX idx_user_quota_user_id ON user_quota (user_id);

-- Comments for documentation
COMMENT ON TABLE users IS 'User accounts and authentication information';
COMMENT ON TABLE repositories IS 'File repositories owned by users';
COMMENT ON TABLE files IS 'Metadata for files and directories stored in repositories';
COMMENT ON TABLE shares IS 'Shared access to repository paths for specific users';
COMMENT ON TABLE user_quota IS 'Storage quota management for users';

-- Relations documentation
/*
Relationships summary:

users table is the central identity table
  - repositories table references users via owner_id (many-to-one)
  - files table references users via owner_id (many-to-one)
  - shares table references users via owner_id and user_id (many-to-many)
  - user_quota table references users via user_id (one-to-one)

repositories table
  - files table references repositories via repo_id (many-to-one)
  - shares table references repositories via repo_id (many-to-one)

files table stores metadata about files and directories
  - parent_id references other files for hierarchical structure
*/