CREATE TABLE change_log (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    operation VARCHAR(20) NOT NULL CHECK (operation IN ('create', 'modify', 'delete', 'move', 'copy')),
    path TEXT NOT NULL,
    old_path TEXT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    version VARCHAR(64) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE repository_versions (
    id SERIAL PRIMARY KEY,
    repo_id INTEGER NOT NULL UNIQUE REFERENCES repositories(id),
    current_version VARCHAR(64) NOT NULL,
    version_vector JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE upload_sessions (
    id SERIAL PRIMARY KEY,
    upload_id VARCHAR(64) UNIQUE NOT NULL,
    repo_id INTEGER NOT NULL REFERENCES repositories(id),
    path TEXT NOT NULL,
    total_size BIGINT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES users(id),
    chunks_uploaded INTEGER DEFAULT 0,
    total_chunks INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP + INTERVAL '1 day',
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'completed', 'cancelled'))
);

CREATE TABLE upload_chunks (
    id SERIAL PRIMARY KEY,
    upload_id VARCHAR(64) NOT NULL REFERENCES upload_sessions(upload_id) ON DELETE CASCADE,
    chunk_index INTEGER NOT NULL,
    offset BIGINT NOT NULL,
    size BIGINT NOT NULL,
    checksum VARCHAR(64),
    uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(upload_id, chunk_index)
);

CREATE INDEX idx_change_log_repo_id ON change_log(repo_id);
CREATE INDEX idx_change_log_path ON change_log(path);
CREATE INDEX idx_change_log_timestamp ON change_log(timestamp DESC);
CREATE INDEX idx_change_log_version ON change_log(version);
CREATE INDEX idx_change_log_repo_version ON change_log(repo_id, version);

CREATE INDEX idx_upload_sessions_upload_id ON upload_sessions(upload_id);
CREATE INDEX idx_upload_sessions_repo_id ON upload_sessions(repo_id);
CREATE INDEX idx_upload_sessions_user_id ON upload_sessions(user_id);
CREATE INDEX idx_upload_sessions_expires_at ON upload_sessions(expires_at);

CREATE INDEX idx_upload_chunks_upload_id ON upload_chunks(upload_id);

COMMENT ON TABLE change_log IS 'Tracks all file operations for sync protocol';
COMMENT ON TABLE repository_versions IS 'Stores version state for each repository';
COMMENT ON TABLE upload_sessions IS 'Chunked upload session management';
COMMENT ON TABLE upload_chunks IS 'Individual chunk data for resumable uploads';
