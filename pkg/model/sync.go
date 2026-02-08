package model

import "time"

type ChangeLog struct {
	ID        int       `bun:"id,pk,autoincrement"`
	RepoID    int       `bun:"repo_id,notnull"`
	Operation string    `bun:"operation,notnull"`
	Path      string    `bun:"path,notnull"`
	OldPath   *string   `bun:"old_path"`
	UserID    int       `bun:"user_id,notnull"`
	Version   string    `bun:"version,notnull"`
	Timestamp time.Time `bun:"timestamp,notnull"`
}

type RepositoryVersion struct {
	ID             int       `bun:"id,pk,autoincrement"`
	RepoID         int       `bun:"repo_id,unique,notnull"`
	CurrentVersion string    `bun:"current_version,notnull"`
	VersionVector  string    `bun:"version_vector,notnull"`
	UpdatedAt      time.Time `bun:"updated_at,notnull"`
}

type UploadSession struct {
	ID             int       `bun:"id,pk,autoincrement"`
	UploadID       string    `bun:"upload_id,unique,notnull"`
	RepoID         int       `bun:"repo_id,notnull"`
	Path           string    `bun:"path,notnull"`
	TotalSize      int64     `bun:"total_size,notnull"`
	UserID         int       `bun:"user_id,notnull"`
	ChunksUploaded int       `bun:"chunks_uploaded,default:0"`
	TotalChunks    int       `bun:"total_chunks,notnull"`
	CreatedAt      time.Time `bun:"created_at,notnull"`
	ExpiresAt      time.Time `bun:"expires_at,notnull"`
	Status         string    `bun:"status,default:'active'"`
}

type UploadChunk struct {
	ID         int       `bun:"id,pk,autoincrement"`
	UploadID   string    `bun:"upload_id,notnull"`
	ChunkIndex int       `bun:"chunk_index,notnull"`
	Offset     int64     `bun:"offset,notnull"`
	Size       int64     `bun:"size,notnull"`
	Checksum   *string   `bun:"checksum"`
	UploadedAt time.Time `bun:"uploaded_at,notnull"`
}
