package model

import "time"

type FileObject struct {
	ID        int        `json:"id" bun:"id,pk,autoincrement"`
	UserID    int        `json:"user_id" bun:"user_id,notnull"`
	Path      string     `json:"path" bun:"path,notnull"`
	MimeType  *string    `json:"mime_type,omitempty" bun:"mime_type"`
	Size      int64      `json:"size" bun:"size,notnull"`
	Checksum  *string    `json:"checksum,omitempty" bun:"checksum"`
	CreatedAt time.Time  `json:"created_at" bun:"created_at,notnull"`
	UpdatedAt time.Time  `json:"updated_at" bun:"updated_at,notnull"`
	IsDeleted bool       `json:"is_deleted" bun:"is_deleted,notnull"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" bun:"deleted_at"`
}
