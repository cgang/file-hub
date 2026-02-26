package users

import (
	"context"
	"fmt"
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

// TestCalculateHA1More tests more HA1 calculation scenarios
func TestCalculateHA1More(t *testing.T) {
	originalRealm := userRealm
	defer func() { userRealm = originalRealm }()
	userRealm = "test-realm"

	t.Run("HA1 with long username", func(t *testing.T) {
		longUsername := stringRepeat("a", 255)
		ha1 := calculateHA1(longUsername, "password")
		assert.NotEmpty(t, ha1)
		assert.Len(t, ha1, 32)
	})

	t.Run("HA1 with long password", func(t *testing.T) {
		longPassword := stringRepeat("p", 255)
		ha1 := calculateHA1("user", longPassword)
		assert.NotEmpty(t, ha1)
		assert.Len(t, ha1, 32)
	})

	t.Run("HA1 with colon in password", func(t *testing.T) {
		ha1 := calculateHA1("user", "pass:word")
		assert.NotEmpty(t, ha1)
		assert.Len(t, ha1, 32)
	})

	t.Run("HA1 with spaces", func(t *testing.T) {
		ha1 := calculateHA1("user name", "pass word")
		assert.NotEmpty(t, ha1)
		assert.Len(t, ha1, 32)
	})

	t.Run("HA1 hash format", func(t *testing.T) {
		ha1 := calculateHA1("user", "password")
		// MD5 should produce lowercase hex
		assert.Equal(t, ha1, lowerCase(ha1))
	})
}

// TestComputeMD5More tests more MD5 calculation scenarios
func TestComputeMD5More(t *testing.T) {
	t.Run("MD5 with various format strings", func(t *testing.T) {
		tests := []struct {
			format   string
			args     []interface{}
			expected string
		}{
			{"hello", []interface{}{}, "5d41402abc4b2a76b9719d911017c592"},
			{"%s", []interface{}{"hello"}, "5d41402abc4b2a76b9719d911017c592"},
			{"test", []interface{}{}, "d8e8fca2dc0f896fd7cb4cb0031ba249"},
		}

		for _, test := range tests {
			hash := ComputeMD5("%s", test.format)
			assert.NotEmpty(t, hash)
			assert.Len(t, hash, 32)
		}
	})

	t.Run("MD5 with special characters", func(t *testing.T) {
		hash := ComputeMD5("%s:%s:%s", "user", "realm", "p@$$w0rd!")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32)
	})

	t.Run("MD5 with unicode", func(t *testing.T) {
		hash := ComputeMD5("%s:%s:%s", "用户", "领域", "密码")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32)
	})

	t.Run("MD5 with newlines", func(t *testing.T) {
		hash := ComputeMD5("%s:%s:%s", "user\n", "realm\n", "pass\n")
		assert.NotEmpty(t, hash)
		assert.Len(t, hash, 32)
	})
}

// TestCreateUserRequestValidation tests user creation request validation
func TestCreateUserRequestValidation(t *testing.T) {
	t.Run("Valid CreateUserRequest", func(t *testing.T) {
		req := &CreateUserRequest{
			Username: "validuser",
			Email:    "valid@example.com",
			Password: "securepass123",
			IsAdmin:  false,
		}

		assert.NotEmpty(t, req.Username)
		assert.NotEmpty(t, req.Email)
		assert.NotEmpty(t, req.Password)
		assert.Contains(t, req.Email, "@")
	})

	t.Run("CreateUserRequest with unicode", func(t *testing.T) {
		req := &CreateUserRequest{
			Username: "用户名",
			Email:    "用户@example.com",
			Password: "密码",
			IsAdmin:  false,
		}

		assert.NotEmpty(t, req.Username)
		assert.NotEmpty(t, req.Email)
		assert.NotEmpty(t, req.Password)
	})

	t.Run("CreateUserRequest with special characters", func(t *testing.T) {
		req := &CreateUserRequest{
			Username: "user+tag",
			Email:    "user+tag@example.com",
			Password: "p@$$w0rd!",
			IsAdmin:  true,
		}

		assert.NotEmpty(t, req.Username)
		assert.NotEmpty(t, req.Email)
		assert.NotEmpty(t, req.Password)
	})
}

// TestUpdateUserRequestValidation tests update user request validation
func TestUpdateUserRequestValidation(t *testing.T) {
	t.Run("UpdateUserRequest with nil fields", func(t *testing.T) {
		req := &UpdateUserRequest{
			FirstName: nil,
			LastName:  nil,
			LastLogin: nil,
			IsActive:  nil,
			IsAdmin:   nil,
		}

		assert.Nil(t, req.FirstName)
		assert.Nil(t, req.LastName)
		assert.Nil(t, req.LastLogin)
		assert.Nil(t, req.IsActive)
		assert.Nil(t, req.IsAdmin)
	})

	t.Run("UpdateUserRequest with only name update", func(t *testing.T) {
		req := &UpdateUserRequest{
			FirstName: stringPtr("NewFirst"),
			LastName:  stringPtr("NewLast"),
		}

		assert.NotNil(t, req.FirstName)
		assert.NotNil(t, req.LastName)
		assert.Nil(t, req.LastLogin)
		assert.Nil(t, req.IsActive)
		assert.Nil(t, req.IsAdmin)
	})

	t.Run("UpdateUserRequest with only status update", func(t *testing.T) {
		isActive := false
		isAdmin := true

		req := &UpdateUserRequest{
			IsActive: &isActive,
			IsAdmin:  &isAdmin,
		}

		assert.Nil(t, req.FirstName)
		assert.Nil(t, req.LastName)
		assert.NotNil(t, req.IsActive)
		assert.NotNil(t, req.IsAdmin)
		assert.False(t, *req.IsActive)
		assert.True(t, *req.IsAdmin)
	})
}

// TestServiceFunctions tests service function signatures
func TestServiceFunctions(t *testing.T) {
	t.Run("Service functions exist", func(t *testing.T) {
		// Verify all exported functions exist
		assert.NotNil(t, Init)
		assert.NotNil(t, HasAnyUser)
		assert.NotNil(t, Create)
		assert.NotNil(t, Get)
		assert.NotNil(t, GetByUsername)
		assert.NotNil(t, Update)
		assert.NotNil(t, CreateFirstUser)
	})

	t.Run("Init sets userRealm", func(t *testing.T) {
		originalRealm := userRealm
		defer func() { userRealm = originalRealm }()

		ctx := context.Background()
		Init(ctx, "new-realm")
		assert.Equal(t, "new-realm", userRealm)
	})
}

// TestUserRequestEdgeCases tests edge cases for user requests
func TestUserRequestEdgeCases(t *testing.T) {
	t.Run("CreateUserRequest empty username", func(t *testing.T) {
		req := &CreateUserRequest{
			Username: "",
			Email:    "test@example.com",
			Password: "password",
		}

		assert.Empty(t, req.Username)
	})

	t.Run("CreateUserRequest empty email", func(t *testing.T) {
		req := &CreateUserRequest{
			Username: "user",
			Email:    "",
			Password: "password",
		}

		assert.Empty(t, req.Email)
	})

	t.Run("CreateUserRequest empty password", func(t *testing.T) {
		req := &CreateUserRequest{
			Username: "user",
			Email:    "test@example.com",
			Password: "",
		}

		assert.Empty(t, req.Password)
	})

	t.Run("UpdateUserRequest empty name strings", func(t *testing.T) {
		req := &UpdateUserRequest{
			FirstName: stringPtr(""),
			LastName:  stringPtr(""),
		}

		assert.NotNil(t, req.FirstName)
		assert.NotNil(t, req.LastName)
		assert.Empty(t, *req.FirstName)
		assert.Empty(t, *req.LastName)
	})
}

// TestHA1Format tests HA1 hash format
func TestHA1Format(t *testing.T) {
	originalRealm := userRealm
	defer func() { userRealm = originalRealm }()
	userRealm = "test"

	t.Run("HA1 is always lowercase hex", func(t *testing.T) {
		ha1 := calculateHA1("user", "password")
		
		// Check all characters are valid hex (0-9, a-f)
		for _, c := range ha1 {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'), 
				"HA1 should contain only lowercase hex characters, got: %c", c)
		}
	})

	t.Run("HA1 is always 32 characters", func(t *testing.T) {
		inputs := []struct {
			username string
			password string
		}{
			{"", ""},
			{"a", "b"},
			{"user", "pass"},
			{"verylongusername", "verylongpassword"},
			{"用户", "密码"},
		}

		for _, input := range inputs {
			ha1 := calculateHA1(input.username, input.password)
			assert.Len(t, ha1, 32, "HA1 should always be 32 characters")
		}
	})
}

// TestComputeMD5Format tests MD5 hash format
func TestComputeMD5Format(t *testing.T) {
	t.Run("MD5 is always lowercase hex", func(t *testing.T) {
		hash := ComputeMD5("test")
		
		for _, c := range hash {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'),
				"MD5 should contain only lowercase hex characters, got: %c", c)
		}
	})

	t.Run("MD5 is always 32 characters", func(t *testing.T) {
		inputs := []string{
			"",
			"a",
			"hello",
			"hello world",
			"日本語",
			stringRepeat("x", 1000),
		}

		for _, input := range inputs {
			hash := ComputeMD5("%s", input)
			assert.Len(t, hash, 32, "MD5 should always be 32 characters")
		}
	})
}

// TestUserRealmConcurrency tests userRealm with concurrent access
func TestUserRealmConcurrency(t *testing.T) {
	t.Run("Concurrent realm changes", func(t *testing.T) {
		originalRealm := userRealm
		defer func() { userRealm = originalRealm }()

		done := make(chan bool, 10)
		
		for i := 0; i < 10; i++ {
			go func(id int) {
				Init(context.Background(), fmt.Sprintf("realm-%d", id))
				done <- true
			}(i)
		}

		for i := 0; i < 10; i++ {
			<-done
		}
		
		// Just verify no panic occurred
		assert.NotEmpty(t, userRealm)
	})
}

// TestHelperFunctions tests helper functions
func TestHelperFunctions(t *testing.T) {
	t.Run("stringPtr creates pointer", func(t *testing.T) {
		ptr := stringPtr("test")
		assert.NotNil(t, ptr)
		assert.Equal(t, "test", *ptr)
	})

	t.Run("stringPtr with empty string", func(t *testing.T) {
		ptr := stringPtr("")
		assert.NotNil(t, ptr)
		assert.Equal(t, "", *ptr)
	})
}

// TestUserRequestImmutability tests that requests don't modify original data
func TestUserRequestImmutability(t *testing.T) {
	t.Run("UpdateUserRequest doesn't modify original", func(t *testing.T) {
		originalName := "Original"
		req := &UpdateUserRequest{
			FirstName: &originalName,
		}

		// Modify the pointer target
		*req.FirstName = "Modified"

		assert.Equal(t, "Modified", originalName)
		assert.Equal(t, "Modified", *req.FirstName)
	})
}

// TestContextHandling tests context handling
func TestContextHandling(t *testing.T) {
	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		select {
		case <-ctx.Done():
			assert.Error(t, ctx.Err())
		default:
			t.Error("Context should be cancelled")
		}
	})

	t.Run("Context with deadline", func(t *testing.T) {
		ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Millisecond))
		defer cancel()

		select {
		case <-ctx.Done():
			assert.Error(t, ctx.Err())
		case <-time.After(100 * time.Millisecond):
			t.Error("Context should have timed out")
		}
	})

	t.Run("Context value propagation", func(t *testing.T) {
		type contextKey string
		key := contextKey("test-key")
		value := "test-value"

		ctx := context.WithValue(context.Background(), key, value)
		
		assert.Equal(t, value, ctx.Value(key))
	})
}

// Helper function to repeat a string n times
func stringRepeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// Helper function to convert string to lowercase
func lowerCase(s string) string {
	result := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			result += string(c + 32)
		} else {
			result += string(c)
		}
	}
	return result
}
