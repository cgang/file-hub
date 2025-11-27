package db

import (
	"database/sql"
	"time"

	"github.com/uptrace/bun"
)

// ErrNoRows is returned when a query returns no rows
var ErrNoRows = sql.ErrNoRows

// User represents a user in the system
type User struct {
	bun.BaseModel `bun:"table:users"`

	ID          int        `json:"id" bun:"id,pk,autoincrement"`
	Username    string     `json:"username" bun:"username,unique,notnull"`
	Email       string     `json:"email" bun:"email,unique,notnull"`
	HA1         string     `json:"-" bun:"ha1_hash,notnull"` // Don't expose HA1 hash in JSON
	Realm       string     `json:"-" bun:"realm,notnull"`    // Don't expose realm in JSON
	FirstName   *string    `json:"first_name,omitempty" bun:"first_name"`
	LastName    *string    `json:"last_name,omitempty" bun:"last_name"`
	CreatedAt   time.Time  `json:"created_at" bun:"created_at,notnull"`
	UpdatedAt   time.Time  `json:"updated_at" bun:"updated_at,notnull"`
	LastLogin   *time.Time `json:"last_login,omitempty" bun:"last_login"`
	IsActive    bool       `json:"is_active" bun:"is_active,notnull"`
	IsAdmin     bool       `json:"is_admin" bun:"is_admin,notnull"`
}

// File represents a file in the system
type File struct {
	bun.BaseModel `bun:"table:files"`

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

// UserQuota represents storage quota for a user
type UserQuota struct {
	bun.BaseModel `bun:"table:user_quota"`

	ID              int       `json:"id" bun:"id,pk,autoincrement"`
	UserID          int       `json:"user_id" bun:"user_id,unique,notnull"`
	TotalQuotaBytes int64     `json:"total_quota_bytes" bun:"total_quota_bytes,notnull"`
	UsedBytes       int64     `json:"used_bytes" bun:"used_bytes,notnull"`
	UpdatedAt       time.Time `json:"updated_at" bun:"updated_at,notnull"`
}