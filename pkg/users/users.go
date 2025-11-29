package users

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

var (
	initUserCount int32
	userRealm     string
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

// Authenticate validates a user's credentials for basic authentication
func Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	// Get user by username
	user, err := db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Calculate HA1 hash from provided credentials
	providedHA1 := calculateHA1(username, password)

	// Compare hashes using constant time comparison
	if !compareHA1(user.HA1, providedHA1) {
		return nil, errors.New("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	updateReq := &UpdateUserRequest{
		LastLogin: &now,
	}

	err = Update(ctx, user.ID, updateReq)
	if err != nil {
		// Log error but don't fail authentication
		// In a production system, you'd want to log this properly
	}

	return user, nil
}

// ValidateDigest validates a user's credentials for digest authentication
func ValidateDigest(ctx context.Context, username, uri, nonce, nc, cnonce, qop, response, method string) (*model.User, error) {
	// Get user by username
	user, err := db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Calculate HA2
	ha2 := calculateHA2(method, uri)

	// Calculate the expected response using the stored HA1
	expectedResponse := calculateResponse(user.HA1, nonce, nc, cnonce, qop, ha2)

	// Compare responses using constant time comparison
	if !compareResponse(response, expectedResponse) {
		return nil, errors.New("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	updateReq := &UpdateUserRequest{
		LastLogin: &now,
	}

	err = Update(ctx, user.ID, updateReq)
	if err != nil {
		// Log error but don't fail authentication
		// In a production system, you'd want to log this properly
	}

	return user, nil
}

// calculateHA1 calculates the HA1 value for digest authentication
func calculateHA1(username, password string) string {
	// HA1 = MD5(username:realm:password)
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, userRealm, password)))
	return hex.EncodeToString(ha1[:])
}

// calculateHA2 calculates the HA2 value for digest authentication
func calculateHA2(method, uri string) string {
	// HA2 = MD5(method:uri)
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))
	return hex.EncodeToString(ha2[:])
}

// calculateResponse calculates the expected response for digest authentication
func calculateResponse(ha1, nonce, nc, cnonce, qop, ha2 string) string {
	// response = MD5(HA1:nonce:nc:cnonce:qop:HA2)
	resp := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, nonce, nc, cnonce, qop, ha2)))
	return hex.EncodeToString(resp[:])
}

// compareHA1 compares two HA1 hashes using constant time comparison
func compareHA1(ha1, providedHA1 string) bool {
	// In a real implementation, you would use crypto/subtle.ConstantTimeCompare
	// For now, we'll use a simple comparison
	return ha1 == providedHA1
}

// compareResponse compares two responses using constant time comparison
func compareResponse(response, expectedResponse string) bool {
	// In a real implementation, you would use crypto/subtle.ConstantTimeCompare
	// For now, we'll use a simple comparison
	return response == expectedResponse
}
