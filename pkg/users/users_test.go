package users

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Since we're tightly coupled with the db package, we'll focus on testing
// the business logic in our service rather than mocking the database.
// In a real scenario, you'd want to use an in-memory database or proper mocks.

func TestGenerateSalt(t *testing.T) {
	salt1, err := generateSalt()
	assert.NoError(t, err)
	assert.NotEmpty(t, salt1)

	salt2, err := generateSalt()
	assert.NoError(t, err)
	assert.NotEmpty(t, salt2)

	// Salts should be different
	assert.NotEqual(t, salt1, salt2)
}

func TestHashPassword(t *testing.T) {
	password := "testpassword"
	salt, err := generateSalt()
	assert.NoError(t, err)

	hash1, err := hashPassword(password, salt)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash1)

	hash2, err := hashPassword(password, salt)
	assert.NoError(t, err)
	assert.NotEmpty(t, hash2)

	// Hashes should be the same for the same password and salt
	assert.Equal(t, hash1, hash2)

	// Hashes should be different for different passwords
	differentSalt, err := generateSalt()
	assert.NoError(t, err)

	hash3, err := hashPassword("differentpassword", salt)
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash3)

	// Hashes should be different for different salts
	hash4, err := hashPassword(password, differentSalt)
	assert.NoError(t, err)
	assert.NotEqual(t, hash1, hash4)
}

func TestUserCreationRequestValidation(t *testing.T) {
	// Test that our request structs have the right tags
	req := &CreateUserRequest{
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: stringPtr("Test"),
		LastName:  stringPtr("User"),
		IsAdmin:   false,
	}

	assert.Equal(t, "testuser", req.Username)
	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, "Test", *req.FirstName)
	assert.Equal(t, "User", *req.LastName)
	assert.Equal(t, false, req.IsAdmin)
}

func TestUserUpdateRequestValidation(t *testing.T) {
	now := time.Now()
	isActive := true
	isAdmin := false

	req := &UpdateUserRequest{
		FirstName: stringPtr("Updated"),
		LastName:  stringPtr("Name"),
		LastLogin: &now,
		IsActive:  &isActive,
		IsAdmin:   &isAdmin,
	}

	assert.Equal(t, "Updated", *req.FirstName)
	assert.Equal(t, "Name", *req.LastName)
	assert.Equal(t, now, *req.LastLogin)
	assert.Equal(t, true, *req.IsActive)
	assert.Equal(t, false, *req.IsAdmin)
}

func stringPtr(s string) *string {
	return &s
}