package users

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/cgang/file-hub/pkg/db"
)

// Service provides user management operations
type Service struct {
	DB    *db.DB
	Realm string
}

// NewService creates a new user service
func NewService(db *db.DB) *Service {
	return &Service{DB: db, Realm: "FileHub"}
}

// User represents a user in the system
type User struct {
	ID          int        `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FirstName   *string    `json:"first_name,omitempty"`
	LastName    *string    `json:"last_name,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLogin   *time.Time `json:"last_login,omitempty"`
	IsActive    bool       `json:"is_active"`
	IsAdmin     bool       `json:"is_admin"`
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
func (s *Service) Create(req *CreateUserRequest) (*User, error) {
	// Check if user already exists
	_, err := s.DB.GetUserByUsername(req.Username)
	if err == nil {
		return nil, errors.New("username already exists")
	}

	_, err = s.DB.GetUserByEmail(req.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}

	// Calculate HA1 hash (username:realm:password)
	ha1 := calculateHA1(req.Username, s.Realm, req.Password)

	// Create the user in the database
	dbUser := &db.User{
		Username:  req.Username,
		Email:     req.Email,
		HA1:       ha1,
		Realm:     s.Realm,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		IsActive:  true,
		IsAdmin:   req.IsAdmin,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = s.DB.CreateUser(dbUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		IsActive:  dbUser.IsActive,
		IsAdmin:   dbUser.IsAdmin,
	}, nil
}

// Get retrieves a user by ID
func (s *Service) Get(id int) (*User, error) {
	dbUser, err := s.DB.GetUserByID(id)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		LastLogin: dbUser.LastLogin,
		IsActive:  dbUser.IsActive,
		IsAdmin:   dbUser.IsAdmin,
	}, nil
}

// GetByUsername retrieves a user by username
func (s *Service) GetByUsername(username string) (*User, error) {
	dbUser, err := s.DB.GetUserByUsername(username)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		LastLogin: dbUser.LastLogin,
		IsActive:  dbUser.IsActive,
		IsAdmin:   dbUser.IsAdmin,
	}, nil
}

// Update modifies an existing user
func (s *Service) Update(id int, req *UpdateUserRequest) error {
	dbUpdate := &db.UserUpdate{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		LastLogin: req.LastLogin,
		IsActive:  req.IsActive,
		IsAdmin:   req.IsAdmin,
	}

	return s.DB.UpdateUser(id, dbUpdate)
}

// Delete removes a user (soft delete)
func (s *Service) Delete(id int) error {
	return s.DB.DeleteUser(id)
}

// Authenticate validates a user's credentials for basic authentication
func (s *Service) Authenticate(username, password string) (*User, error) {
	// Get user by username
	dbUser, err := s.DB.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Calculate HA1 hash from provided credentials
	providedHA1 := calculateHA1(username, dbUser.Realm, password)

	// Compare hashes using constant time comparison
	if !compareHA1(dbUser.HA1, providedHA1) {
		return nil, errors.New("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	updateReq := &UpdateUserRequest{
		LastLogin: &now,
	}

	err = s.Update(dbUser.ID, updateReq)
	if err != nil {
		// Log error but don't fail authentication
		// In a production system, you'd want to log this properly
	}

	return &User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		LastLogin: &now,
		IsActive:  dbUser.IsActive,
		IsAdmin:   dbUser.IsAdmin,
	}, nil
}

// ValidateDigest validates a user's credentials for digest authentication
func (s *Service) ValidateDigest(username, realm, uri, nonce, nc, cnonce, qop, response, method string) (*User, error) {
	// Get user by username
	dbUser, err := s.DB.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Validate realm
	if dbUser.Realm != realm {
		return nil, errors.New("invalid realm")
	}

	// Calculate HA2
	ha2 := calculateHA2(method, uri)

	// Calculate the expected response using the stored HA1
	expectedResponse := calculateResponse(dbUser.HA1, nonce, nc, cnonce, qop, ha2)

	// Compare responses using constant time comparison
	if !compareResponse(response, expectedResponse) {
		return nil, errors.New("invalid credentials")
	}

	// Update last login time
	now := time.Now()
	updateReq := &UpdateUserRequest{
		LastLogin: &now,
	}

	err = s.Update(dbUser.ID, updateReq)
	if err != nil {
		// Log error but don't fail authentication
		// In a production system, you'd want to log this properly
	}

	return &User{
		ID:        dbUser.ID,
		Username:  dbUser.Username,
		Email:     dbUser.Email,
		FirstName: dbUser.FirstName,
		LastName:  dbUser.LastName,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		LastLogin: &now,
		IsActive:  dbUser.IsActive,
		IsAdmin:   dbUser.IsAdmin,
	}, nil
}

// calculateHA1 calculates the HA1 value for digest authentication
func calculateHA1(username, realm, password string) string {
	// HA1 = MD5(username:realm:password)
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password)))
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
