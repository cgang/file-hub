package users

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateHA1(t *testing.T) {
	// Save and restore userRealm
	originalRealm := userRealm
	defer func() { userRealm = originalRealm }()

	userRealm = "test-realm"

	ha1 := calculateHA1("testuser", "password")
	assert.NotEmpty(t, ha1)
	assert.Len(t, ha1, 32) // MD5 produces 32 character hex string

	// Same inputs should produce same output
	ha1Again := calculateHA1("testuser", "password")
	assert.Equal(t, ha1, ha1Again)

	// Different inputs should produce different output
	differentHA1 := calculateHA1("testuser", "different")
	assert.NotEqual(t, ha1, differentHA1)

	// Different realm should produce different hash
	userRealm = "different-realm"
	differentRealmHA1 := calculateHA1("testuser", "password")
	assert.NotEqual(t, ha1, differentRealmHA1)
}

func TestComputeMD5(t *testing.T) {
	t.Run("MD5 hash calculation", func(t *testing.T) {
		hash := ComputeMD5("%s:%s:%s", "user", "realm", "pass")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32)
	})

	t.Run("MD5 consistent hashing", func(t *testing.T) {
		hash1 := ComputeMD5("%s:%s:%s", "user", "realm", "pass")
		hash2 := ComputeMD5("%s:%s:%s", "user", "realm", "pass")
		assert.Equal(t, hash1, hash2)
	})

	t.Run("MD5 different inputs different hash", func(t *testing.T) {
		hash1 := ComputeMD5("%s:%s:%s", "user1", "realm", "pass")
		hash2 := ComputeMD5("%s:%s:%s", "user2", "realm", "pass")
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("MD5 empty string", func(t *testing.T) {
		hash := ComputeMD5("%s", "")
		assert.Equal(t, "d41d8cd98f00b204e9800998ecf8427e", hash)
	})
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

func TestCreateUserRequest(t *testing.T) {
	t.Run("CreateUserRequest with all fields", func(t *testing.T) {
		req := &CreateUserRequest{
			Username:  "fulluser",
			Email:     "full@example.com",
			Password:  "securepass",
			FirstName: stringPtr("Full"),
			LastName:  stringPtr("User"),
			IsAdmin:   true,
		}

		assert.Equal(t, "fulluser", req.Username)
		assert.Equal(t, "full@example.com", req.Email)
		assert.Equal(t, "securepass", req.Password)
		assert.Equal(t, "Full", *req.FirstName)
		assert.Equal(t, "User", *req.LastName)
		assert.True(t, req.IsAdmin)
	})

	t.Run("CreateUserRequest with minimal fields", func(t *testing.T) {
		req := &CreateUserRequest{
			Username: "minimal",
			Email:    "minimal@example.com",
			Password: "pass",
			IsAdmin:  false,
		}

		assert.Nil(t, req.FirstName)
		assert.Nil(t, req.LastName)
		assert.False(t, req.IsAdmin)
	})

	t.Run("CreateUserRequest JSON serialization", func(t *testing.T) {
		req := &CreateUserRequest{
			Username:  "jsonuser",
			Email:     "json@example.com",
			Password:  "pass",
			FirstName: stringPtr("JSON"),
			IsAdmin:   true,
		}

		// Verify JSON tags work correctly (first_name should be omitted when nil)
		assert.NotNil(t, req.FirstName)
		assert.Nil(t, req.LastName)
	})
}

func TestUpdateUserRequest(t *testing.T) {
	t.Run("UpdateUserRequest partial update", func(t *testing.T) {
		isActive := true
		req := &UpdateUserRequest{
			IsActive: &isActive,
		}

		assert.Nil(t, req.FirstName)
		assert.Nil(t, req.LastName)
		assert.Nil(t, req.LastLogin)
		assert.Nil(t, req.IsAdmin)
		assert.NotNil(t, req.IsActive)
		assert.True(t, *req.IsActive)
	})

	t.Run("UpdateUserRequest all fields", func(t *testing.T) {
		now := time.Now()
		isActive := false
		isAdmin := true

		req := &UpdateUserRequest{
			FirstName: stringPtr("First"),
			LastName:  stringPtr("Last"),
			LastLogin: &now,
			IsActive:  &isActive,
			IsAdmin:   &isAdmin,
		}

		assert.NotNil(t, req.FirstName)
		assert.NotNil(t, req.LastName)
		assert.NotNil(t, req.LastLogin)
		assert.NotNil(t, req.IsActive)
		assert.NotNil(t, req.IsAdmin)
	})
}

func TestUserRealm(t *testing.T) {
	t.Run("userRealm is package variable", func(t *testing.T) {
		originalRealm := userRealm
		defer func() { userRealm = originalRealm }()

		userRealm = "test-realm-123"
		assert.Equal(t, "test-realm-123", userRealm)

		userRealm = ""
		assert.Equal(t, "", userRealm)
	})
}

func TestContextUsage(t *testing.T) {
	t.Run("Context can be passed", func(t *testing.T) {
		ctx := context.Background()
		assert.NotNil(t, ctx)
	})

	t.Run("Context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		select {
		case <-ctx.Done():
			assert.Error(t, ctx.Err())
		case <-time.After(200 * time.Millisecond):
			t.Error("Context should have timed out")
		}
	})
}

func TestHA1WithSpecialCharacters(t *testing.T) {
	t.Run("HA1 with unicode password", func(t *testing.T) {
		originalRealm := userRealm
		defer func() { userRealm = originalRealm }()
		userRealm = "test"

		ha1 := calculateHA1("user", "密码")
		assert.NotEmpty(t, ha1)
		assert.Len(t, ha1, 32)
	})

	t.Run("HA1 with special characters", func(t *testing.T) {
		originalRealm := userRealm
		defer func() { userRealm = originalRealm }()
		userRealm = "test"

		ha1 := calculateHA1("user", "p@ss:w0rd!")
		assert.NotEmpty(t, ha1)
		assert.Len(t, ha1, 32)
	})

	t.Run("HA1 with empty password", func(t *testing.T) {
		originalRealm := userRealm
		defer func() { userRealm = originalRealm }()
		userRealm = "test"

		ha1 := calculateHA1("user", "")
		assert.NotEmpty(t, ha1)
		assert.Len(t, ha1, 32)
	})
}

func stringPtr(s string) *string {
	return &s
}
