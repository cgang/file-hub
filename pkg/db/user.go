package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CreateUser creates a new user in the database
func (d *DB) CreateUser(user *User) error {
	// Set creation timestamp
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	_, err := d.NewInsert().Model(user).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Initialize user quota
	quota := &UserQuota{
		UserID:          user.ID,
		TotalQuotaBytes: 10737418240, // 10GB default
		UsedBytes:       0,
		UpdatedAt:       time.Now(),
	}

	_, err = d.NewInsert().Model(quota).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("failed to initialize user quota: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (d *DB) GetUserByID(id int) (*User, error) {
	user := &User{ID: id}
	err := d.NewSelect().
		Model(user).
		Where("id = ? AND is_active = ?", id, true).
		Scan(context.Background())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (d *DB) GetUserByUsername(username string) (*User, error) {
	user := &User{}
	err := d.NewSelect().
		Model(user).
		Where("username = ? AND is_active = ?", username, true).
		Scan(context.Background())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (d *DB) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := d.NewSelect().
		Model(user).
		Where("email = ? AND is_active = ?", email, true).
		Scan(context.Background())

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UserUpdate contains fields that can be updated for a user
type UserUpdate struct {
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	IsAdmin   *bool      `json:"is_admin,omitempty"`
}

// UpdateUser updates a user in the database
func (d *DB) UpdateUser(id int, update *UserUpdate) error {
	user := &User{ID: id}

	// Get the existing user first
	err := d.NewSelect().Model(user).Where("id = ?", id).Scan(context.Background())
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found")
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if they are provided
	if update.FirstName != nil {
		user.FirstName = update.FirstName
	}
	if update.LastName != nil {
		user.LastName = update.LastName
	}
	if update.LastLogin != nil {
		user.LastLogin = update.LastLogin
	}
	if update.IsActive != nil {
		user.IsActive = *update.IsActive
	}
	if update.IsAdmin != nil {
		user.IsAdmin = *update.IsAdmin
	}

	user.UpdatedAt = time.Now()

	result, err := d.NewUpdate().Model(user).Where("id = ?", id).Exec(context.Background())
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// DeleteUser marks a user as inactive (soft delete)
func (d *DB) DeleteUser(id int) error {
	user := &User{ID: id, IsActive: false, UpdatedAt: time.Now()}
	result, err := d.NewUpdate().Model(user).Where("id = ?", id).Exec(context.Background())

	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateUserHA1 updates a user's HA1 hash and realm
func (d *DB) UpdateUserHA1(id int, ha1, realm string) error {
	user := &User{ID: id, HA1: ha1, Realm: realm, UpdatedAt: time.Now()}
	result, err := d.NewUpdate().Model(user).Where("id = ?", id).Exec(context.Background())

	if err != nil {
		return fmt.Errorf("failed to update user HA1: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}