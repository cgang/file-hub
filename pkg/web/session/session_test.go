package session

import (
	"testing"
	"time"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestSessionStore(t *testing.T) {
	// Create a new session store
	store := NewStore()

	// Create a test user
	user := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Test creating a session
	session, err := store.Create(user)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, user, session.User)

	// Test retrieving a session
	retrievedSession, ok := store.Get(session.ID)
	assert.True(t, ok)
	assert.Equal(t, session, retrievedSession)

	// Test extending a session
	time.Sleep(100 * time.Millisecond) // Small delay
	ok = store.Extend(session.ID)
	assert.True(t, ok)

	// Test destroying a session
	store.Destroy(session.ID)
	_, ok = store.Get(session.ID)
	assert.False(t, ok)
}

func TestSessionExpiration(t *testing.T) {
	// Create a new session store
	store := NewStore()

	// Create a test user
	user := &model.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Create a session with a short expiry time for testing
	sessionID, err := generateSessionID()
	assert.NoError(t, err)

	session := &Session{
		ID:        sessionID,
		User:      user,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}

	store.sessions[sessionID] = session

	// Try to get the expired session
	_, ok := store.Get(sessionID)
	assert.False(t, ok)

	// Verify the session was removed
	_, exists := store.sessions[sessionID]
	assert.False(t, exists)
}
