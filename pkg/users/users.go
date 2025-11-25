package users

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/scrypt"

	"github.com/cgang/file-hub/pkg/db"
)

// Service provides user management operations
type Service struct {
	DB *db.DB
}

// NewService creates a new user service
func NewService(db *db.DB) *Service {
	return &Service{DB: db}
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

	// Hash the password
	salt, err := generateSalt()
	if err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	passwordHash, err := hashPassword(req.Password, salt)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create the user in the database
	dbUser := &db.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  passwordHash,
		Salt:      salt,
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

// Authenticate validates a user's credentials
func (s *Service) Authenticate(username, password string) (*User, error) {
	// Get user by username
	dbUser, err := s.DB.GetUserByUsername(username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Hash the provided password with the user's salt
	passwordHash, err := hashPassword(password, dbUser.Salt)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Compare hashes
	if subtle.ConstantTimeCompare([]byte(passwordHash), []byte(dbUser.Password)) != 1 {
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

// generateSalt creates a random salt for password hashing
func generateSalt() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// hashPassword hashes a password with the provided salt using scrypt
func hashPassword(password, salt string) (string, error) {
	// Using recommended scrypt parameters
	hash, err := scrypt.Key([]byte(password), []byte(salt), 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
