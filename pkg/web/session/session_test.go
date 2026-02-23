package session

import (
	"sync"
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

func TestGenerateSessionID(t *testing.T) {
	t.Run("Session ID format", func(t *testing.T) {
		id, err := generateSessionID()
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.Len(t, id, 32) // 16 bytes = 32 hex characters
	})

	t.Run("Session IDs are unique", func(t *testing.T) {
		ids := make(map[string]bool)
		for i := 0; i < 100; i++ {
			id, err := generateSessionID()
			assert.NoError(t, err)
			assert.False(t, ids[id], "Duplicate session ID generated")
			ids[id] = true
		}
	})

	t.Run("Session ID is hexadecimal", func(t *testing.T) {
		id, err := generateSessionID()
		assert.NoError(t, err)

		// Verify all characters are valid hex
		for _, c := range id {
			assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
		}
	})
}

func TestSessionCreation(t *testing.T) {
	t.Run("Session has correct expiry time", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, err := store.Create(user)
		assert.NoError(t, err)

		// Expiry should be approximately 24 hours from now
		expectedExpiry := time.Now().Add(24 * time.Hour)
		assert.InDelta(t, expectedExpiry.Unix(), session.ExpiresAt.Unix(), 1.0)
	})

	t.Run("Session has creation time", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, err := store.Create(user)
		assert.NoError(t, err)

		assert.True(t, session.CreatedAt.Before(session.ExpiresAt))
		assert.InDelta(t, time.Now().Unix(), session.CreatedAt.Unix(), 1.0)
	})

	t.Run("Multiple sessions for same user", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session1, err := store.Create(user)
		assert.NoError(t, err)

		session2, err := store.Create(user)
		assert.NoError(t, err)

		// Sessions should have different IDs
		assert.NotEqual(t, session1.ID, session2.ID)

		// Both sessions should be retrievable
		retrieved1, ok1 := store.Get(session1.ID)
		retrieved2, ok2 := store.Get(session2.ID)
		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.Equal(t, session1.ID, retrieved1.ID)
		assert.Equal(t, session2.ID, retrieved2.ID)
	})
}

func TestSessionGet(t *testing.T) {
	t.Run("Get non-existent session", func(t *testing.T) {
		store := NewStore()
		session, ok := store.Get("non-existent-id")
		assert.False(t, ok)
		assert.Nil(t, session)
	})

	t.Run("Get removes expired sessions", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, _ := store.Create(user)

		// Manually expire the session
		store.sessions[session.ID].ExpiresAt = time.Now().Add(-1 * time.Hour)

		// Get should return false and remove the session
		retrieved, ok := store.Get(session.ID)
		assert.False(t, ok)
		assert.Nil(t, retrieved)

		// Verify it's removed from the map
		_, exists := store.sessions[session.ID]
		assert.False(t, exists)
	})
}

func TestSessionDestroy(t *testing.T) {
	t.Run("Destroy existing session", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, _ := store.Create(user)

		// Verify session exists
		_, ok := store.Get(session.ID)
		assert.True(t, ok)

		// Destroy the session
		store.Destroy(session.ID)

		// Verify session is gone
		_, ok = store.Get(session.ID)
		assert.False(t, ok)
	})

	t.Run("Destroy non-existent session", func(t *testing.T) {
		store := NewStore()
		// Should not panic
		assert.NotPanics(t, func() {
			store.Destroy("non-existent")
		})
	})
}

func TestSessionExtend(t *testing.T) {
	t.Run("Extend existing session", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, _ := store.Create(user)
		originalExpiry := session.ExpiresAt

		// Wait a bit then extend
		time.Sleep(50 * time.Millisecond)
		ok := store.Extend(session.ID)
		assert.True(t, ok)

		// Get the extended session
		extended, exists := store.Get(session.ID)
		assert.True(t, exists)
		assert.True(t, extended.ExpiresAt.After(originalExpiry))
	})

	t.Run("Extend non-existent session", func(t *testing.T) {
		store := NewStore()
		ok := store.Extend("non-existent")
		assert.False(t, ok)
	})

	t.Run("Extend resets expiry to 24 hours", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, _ := store.Create(user)

		// Manually set a near expiry time
		store.sessions[session.ID].ExpiresAt = time.Now().Add(1 * time.Minute)

		// Extend
		store.Extend(session.ID)

		// Should now expire in ~24 hours
		expectedExpiry := time.Now().Add(24 * time.Hour)
		extended, _ := store.Get(session.ID)
		assert.InDelta(t, expectedExpiry.Unix(), extended.ExpiresAt.Unix(), 1.0)
	})
}

func TestSessionConcurrentAccess(t *testing.T) {
	t.Run("Concurrent session creation", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		var wg sync.WaitGroup
		sessions := make(chan *Session, 10)

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				session, err := store.Create(user)
				if err == nil {
					sessions <- session
				}
			}()
		}

		wg.Wait()
		close(sessions)

		// All sessions should have unique IDs
		ids := make(map[string]bool)
		for session := range sessions {
			assert.False(t, ids[session.ID], "Duplicate session ID")
			ids[session.ID] = true
		}
	})

	t.Run("Concurrent session access and destroy", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, _ := store.Create(user)

		var wg sync.WaitGroup

		// Concurrent reads
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				store.Get(session.ID)
			}()
		}

		// One destroy
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Destroy(session.ID)
		}()

		wg.Wait()

		// Session should be destroyed
		_, ok := store.Get(session.ID)
		assert.False(t, ok)
	})
}

func TestSessionWithDifferentUsers(t *testing.T) {
	t.Run("Sessions for different users", func(t *testing.T) {
		store := NewStore()

		user1 := &model.User{ID: 1, Username: "user1"}
		user2 := &model.User{ID: 2, Username: "user2"}
		user3 := &model.User{ID: 3, Username: "user3"}

		session1, _ := store.Create(user1)
		session2, _ := store.Create(user2)
		session3, _ := store.Create(user3)

		// All sessions should be retrievable
		r1, ok1 := store.Get(session1.ID)
		r2, ok2 := store.Get(session2.ID)
		r3, ok3 := store.Get(session3.ID)

		assert.True(t, ok1)
		assert.True(t, ok2)
		assert.True(t, ok3)

		assert.Equal(t, user1.ID, r1.User.ID)
		assert.Equal(t, user2.ID, r2.User.ID)
		assert.Equal(t, user3.ID, r3.User.ID)
	})
}

func TestSessionProperties(t *testing.T) {
	t.Run("Session ID length", func(t *testing.T) {
		store := NewStore()
		user := &model.User{ID: 1, Username: "testuser"}

		session, _ := store.Create(user)
		assert.Len(t, session.ID, 32)
	})

	t.Run("Session stores user reference", func(t *testing.T) {
		store := NewStore()
		user := &model.User{
			ID:        42,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: stringPtr("Test"),
			LastName:  stringPtr("User"),
			IsAdmin:   true,
		}

		session, _ := store.Create(user)
		assert.Equal(t, user.ID, session.User.ID)
		assert.Equal(t, user.Username, session.User.Username)
		assert.Equal(t, user.IsAdmin, session.User.IsAdmin)
	})
}

func stringPtr(s string) *string {
	return &s
}
