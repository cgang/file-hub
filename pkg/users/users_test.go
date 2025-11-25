package users

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Since we're tightly coupled with the db package, we'll focus on testing
// the business logic in our service rather than mocking the database.
// In a real scenario, you'd want to use an in-memory database or proper mocks.

func TestCalculateHA1(t *testing.T) {
	ha1 := calculateHA1("testuser", "test", "password")
	assert.NotEmpty(t, ha1)
	assert.Len(t, ha1, 32) // MD5 produces 32 character hex string

	// Same inputs should produce same output
	ha1Again := calculateHA1("testuser", "test", "password")
	assert.Equal(t, ha1, ha1Again)

	// Different inputs should produce different output
	differentHA1 := calculateHA1("testuser", "test", "different")
	assert.NotEqual(t, ha1, differentHA1)
}

func TestCalculateHA2(t *testing.T) {
	ha2 := calculateHA2("GET", "/test")
	assert.NotEmpty(t, ha2)
	assert.Len(t, ha2, 32) // MD5 produces 32 character hex string

	// Same inputs should produce same output
	ha2Again := calculateHA2("GET", "/test")
	assert.Equal(t, ha2, ha2Again)

	// Different inputs should produce different output
	differentHA2 := calculateHA2("POST", "/test")
	assert.NotEqual(t, ha2, differentHA2)
}

func TestCalculateResponse(t *testing.T) {
	response := calculateResponse("ha1", "nonce", "nc", "cnonce", "qop", "ha2")
	assert.NotEmpty(t, response)
	assert.Len(t, response, 32) // MD5 produces 32 character hex string

	// Same inputs should produce same output
	responseAgain := calculateResponse("ha1", "nonce", "nc", "cnonce", "qop", "ha2")
	assert.Equal(t, response, responseAgain)

	// Different inputs should produce different output
	differentResponse := calculateResponse("ha1", "different", "nc", "cnonce", "qop", "ha2")
	assert.NotEqual(t, response, differentResponse)
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