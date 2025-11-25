package db

import (
	"time"

	"github.com/go-pg/pg/v10"
)

// ErrNoRows is returned when a query returns no rows
var ErrNoRows = pg.ErrNoRows

// User represents a user in the system
type User struct {
	tableName   struct{} `pg:"users"`
	ID          int        `json:"id" pg:"id,pk"`
	Username    string     `json:"username" pg:"username,unique"`
	Email       string     `json:"email" pg:"email,unique"`
	Password    string     `json:"-" pg:"password_hash"` // Don't expose password hash in JSON
	Salt        string     `json:"-" pg:"salt"`          // Don't expose salt in JSON
	FirstName   *string    `json:"first_name,omitempty" pg:"first_name"`
	LastName    *string    `json:"last_name,omitempty" pg:"last_name"`
	CreatedAt   time.Time  `json:"created_at" pg:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" pg:"updated_at"`
	LastLogin   *time.Time `json:"last_login,omitempty" pg:"last_login"`
	IsActive    bool       `json:"is_active" pg:"is_active"`
	IsAdmin     bool       `json:"is_admin" pg:"is_admin"`
}

// File represents a file in the system
type File struct {
	tableName struct{}   `pg:"files"`
	ID        int        `json:"id" pg:"id,pk"`
	UserID    int        `json:"user_id" pg:"user_id,notnull"`
	Path      string     `json:"path" pg:"path,notnull"`
	MimeType  *string    `json:"mime_type,omitempty" pg:"mime_type"`
	Size      int64      `json:"size" pg:"size"`
	Checksum  *string    `json:"checksum,omitempty" pg:"checksum"`
	CreatedAt time.Time  `json:"created_at" pg:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" pg:"updated_at"`
	IsDeleted bool       `json:"is_deleted" pg:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" pg:"deleted_at"`
}

// UserQuota represents storage quota for a user
type UserQuota struct {
	tableName       struct{}  `pg:"user_quota"`
	ID              int       `json:"id" pg:"id,pk"`
	UserID          int       `json:"user_id" pg:"user_id,unique,notnull"`
	TotalQuotaBytes int64     `json:"total_quota_bytes" pg:"total_quota_bytes"`
	UsedBytes       int64     `json:"used_bytes" pg:"used_bytes"`
	UpdatedAt       time.Time `json:"updated_at" pg:"updated_at"`
}