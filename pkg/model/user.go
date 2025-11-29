package model

import "time"

// User represents a user in the system
type User struct {
	ID        int        `json:"id" bun:"id,pk,autoincrement"`
	Username  string     `json:"username" bun:"username,unique,notnull"`
	Email     string     `json:"email" bun:"email,unique,notnull"`
	HA1       string     `json:"-" bun:"ha1_hash,notnull"` // Don't expose HA1 hash in JSON
	FirstName *string    `json:"first_name,omitempty" bun:"first_name"`
	LastName  *string    `json:"last_name,omitempty" bun:"last_name"`
	CreatedAt time.Time  `json:"created_at" bun:"created_at,notnull"`
	UpdatedAt time.Time  `json:"updated_at" bun:"updated_at,notnull"`
	LastLogin *time.Time `json:"last_login,omitempty" bun:"last_login"`
	IsActive  bool       `json:"is_active" bun:"is_active,notnull"`
	IsAdmin   bool       `json:"is_admin" bun:"is_admin,notnull"`
	HomeDir   string     `json:"home_dir" bun:"home_dir,notnull"`
}

type UserQuota struct {
	ID              int       `json:"id" bun:"id,pk,autoincrement"`
	UserID          int       `json:"user_id" bun:"user_id,unique,notnull"`
	TotalQuotaBytes int64     `json:"total_quota_bytes" bun:"total_quota_bytes,notnull"`
	UsedBytes       int64     `json:"used_bytes" bun:"used_bytes,notnull"`
	UpdatedAt       time.Time `json:"updated_at" bun:"updated_at,notnull"`
}
