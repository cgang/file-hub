package model

import "time"

// A Repository represents a file repository owned by a user.
// Each user may own multiple repositories.
// A repository with user name as its name is considered the user's home repository.
// The home repository is created upon user registration.
type Repository struct {
	ID        int       `json:"id" bun:"id,pk,autoincrement"`
	OwnerID   int       `json:"owner_id" bun:"owner_id,notnull"`
	Name      string    `json:"name" bun:"name,notnull"`
	Root      string    `json:"root" bun:"root,notnull"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,notnull"`
}

// A Share represents a shared access to a repository for a specific user.
// It contains the necessary information to identify the share and the associated user.
type Share struct {
	ID      int    `json:"id" bun:"id,pk,autoincrement"`
	RepoID  int    `json:"repo_id" bun:"repo_id,notnull"`
	OwnerID int    `json:"owner_id" bun:"owner_id,notnull"`
	UserID  int    `json:"user_id" bun:"user_id,notnull"`
	Path    string `json:"path" bun:"path,notnull"`
}

// FileObject represents a file stored in a repository.
// It contains metadata about the file such as its path, size, and MIME type.
type FileObject struct {
	ID        int       `json:"id" bun:"id,pk,autoincrement"`
	ParentID  int       `json:"parent_id,omitempty" bun:"parent_id"`
	OwnerID   int       `json:"owner_id" bun:"owner_id,notnull"`
	RepoID    int       `json:"repo_id" bun:"repo_id,notnull"`
	Name      string    `json:"name" bun:"name,notnull"`
	Path      string    `json:"path" bun:"path,notnull"`
	MimeType  *string   `json:"mime_type,omitempty" bun:"mime_type"`
	Size      int64     `json:"size" bun:"size,notnull"`
	Checksum  *string   `json:"checksum,omitempty" bun:"checksum"`
	CreatedAt time.Time `json:"created_at" bun:"created_at,notnull"`
	UpdatedAt time.Time `json:"updated_at" bun:"updated_at,notnull"`
	IsDir     bool      `json:"is_dir" bun:"is_dir"`
}

func (o *FileObject) ContentType() string {
	if o.IsDir {
		return "httpd/unix-directory"
	}

	if mt := o.MimeType; mt != nil {
		return *mt
	} else {
		return "application/octet-stream"
	}
}
