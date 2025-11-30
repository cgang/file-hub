package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/uptrace/bun"
)

// UserModel represents a user in the system
type UserModel struct {
	bun.BaseModel `bun:"table:users"`
	*model.User
}

func wrapUser(mu *model.User) *UserModel {
	return &UserModel{User: mu}
}

func newUserModel(id int) *UserModel {
	return &UserModel{User: &model.User{ID: id}}
}

// CreateUser creates a new user in the database
func CreateUser(ctx context.Context, user *model.User) error {
	// Set creation timestamp
	user.CreatedAt = time.Now()
	user.UpdatedAt = user.CreatedAt

	_, err := db.NewInsert().Model(wrapUser(user)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Initialize user quota
	quota := &model.UserQuota{
		UserID:          user.ID,
		TotalQuotaBytes: 10737418240, // 10GB default
		UsedBytes:       0,
		UpdatedAt:       time.Now(),
	}

	_, err = db.NewInsert().Model(wrapQuota(quota)).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to initialize user quota: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(ctx context.Context, id int) (*model.User, error) {
	user := newUserModel(id)
	err := db.NewSelect().
		Model(user).
		Where("id = ? AND is_active = ?", id, true).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.User, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user := &UserModel{}
	err := db.NewSelect().
		Model(user).
		Where("username = ? AND is_active = ?", username, true).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.User, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &UserModel{}
	err := db.NewSelect().
		Model(user).
		Where("email = ? AND is_active = ?", email, true).
		Scan(ctx)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.User, nil
}

func CountUsers(ctx context.Context) (int, error) {
	count, err := db.NewSelect().Model((*UserModel)(nil)).Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
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
func UpdateUser(ctx context.Context, id int, update *UserUpdate) error {
	user := newUserModel(id)

	// Get the existing user first
	err := db.NewSelect().Model(user).Where("id = ?", id).Scan(ctx)
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

	result, err := db.NewUpdate().Model(user).Where("id = ?", id).Exec(ctx)
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
func DeleteUser(ctx context.Context, id int) error {
	user := &model.User{ID: id, IsActive: false, UpdatedAt: time.Now()}
	result, err := db.NewUpdate().Model(wrapUser(user)).Where("id = ?", id).Exec(ctx)
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
func UpdateUserHA1(ctx context.Context, id int, ha1 string) error {
	user := &model.User{ID: id, HA1: ha1, UpdatedAt: time.Now()}
	result, err := db.NewUpdate().Model(wrapUser(user)).Where("id = ?", id).Exec(ctx)
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
