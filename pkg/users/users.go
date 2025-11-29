package users

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

var (
	userRealm string
)

// NewService creates a new user service
func Init(ctx context.Context, realm string) {
	userRealm = realm
}

func HasAnyUser(ctx context.Context) (bool, error) {
	count, err := db.CountUsers(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to count users: %w", err)
	}
	return count > 0, nil
}

// CreateUserRequest contains the information needed to create a user
type CreateUserRequest struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	IsAdmin   bool    `json:"is_admin"`
}

// UpdateUserRequest contains the information that can be updated for a user
type UpdateUserRequest struct {
	FirstName *string    `json:"first_name,omitempty"`
	LastName  *string    `json:"last_name,omitempty"`
	LastLogin *time.Time `json:"last_login,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	IsAdmin   *bool      `json:"is_admin,omitempty"`
}

// Create creates a new user with the provided details
func Create(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
	// Check if user already exists
	_, err := db.GetUserByUsername(ctx, req.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	_, err = db.GetUserByEmail(ctx, req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// Calculate HA1 hash (username:realm:password)
	ha1 := calculateHA1(req.Username, req.Password)

	// Create the user in the database
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		HA1:       ha1,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		IsAdmin:   req.IsAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = db.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

// Get retrieves a user by ID
func Get(ctx context.Context, id int) (*model.User, error) {
	return db.GetUserByID(ctx, id)
}

// GetByUsername retrieves a user by username
func GetByUsername(ctx context.Context, username string) (*model.User, error) {
	return db.GetUserByUsername(ctx, username)
}

// Update modifies an existing user
func Update(ctx context.Context, id int, req *UpdateUserRequest) error {
	dbUpdate := &db.UserUpdate{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		LastLogin: req.LastLogin,
		IsActive:  req.IsActive,
		IsAdmin:   req.IsAdmin,
	}

	return db.UpdateUser(ctx, id, dbUpdate)
}

// CreateFirstUser creates the first user with the provided details, bypassing duplicate checks
func CreateFirstUser(ctx context.Context, req *CreateUserRequest) (*model.User, error) {
	// Calculate HA1 hash (username:realm:password)
	ha1 := calculateHA1(req.Username, req.Password)

	// Create the user in the database
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		HA1:       ha1,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		IsAdmin:   req.IsAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := db.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}
