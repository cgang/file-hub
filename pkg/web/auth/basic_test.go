package auth

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseBasicAuth(t *testing.T) {
	// Test valid credentials
	username, password, ok := parseBasicAuth(base64.StdEncoding.EncodeToString([]byte("user:pass")))
	assert.True(t, ok)
	assert.Equal(t, "user", username)
	assert.Equal(t, "pass", password)

	// Test credentials with colon in password
	username, password, ok = parseBasicAuth(base64.StdEncoding.EncodeToString([]byte("user:pass:word")))
	assert.True(t, ok)
	assert.Equal(t, "user", username)
	assert.Equal(t, "pass:word", password)

	// Test invalid base64
	username, password, ok = parseBasicAuth("invalid_base64")
	assert.False(t, ok)
	assert.Empty(t, username)
	assert.Empty(t, password)

	// Test missing colon
	username, password, ok = parseBasicAuth(base64.StdEncoding.EncodeToString([]byte("userpass")))
	assert.False(t, ok)
	assert.Empty(t, username)
	assert.Empty(t, password)

	// Test empty credentials
	username, password, ok = parseBasicAuth(base64.StdEncoding.EncodeToString([]byte(":")))
	assert.True(t, ok)
	assert.Empty(t, username)
	assert.Empty(t, password)
}