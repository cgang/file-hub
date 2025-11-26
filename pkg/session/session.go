package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/cgang/file-hub/pkg/users"
)

// Session represents a user session
type Session struct {
	ID        string
	User      *users.User
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Store manages sessions in memory
type Store struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewStore creates a new session store
func NewStore() *Store {
	store := &Store{
		sessions: make(map[string]*Session),
	}
	
	// Start a goroutine to clean up expired sessions periodically
	go store.cleanupExpiredSessions()
	
	return store
}

// generateSessionID creates a random session ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create creates a new session for a user
func (s *Store) Create(user *users.User) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	// Set session expiry to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)

	session := &Session{
		ID:        sessionID,
		User:      user,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	s.mu.Lock()
	s.sessions[sessionID] = session
	s.mu.Unlock()

	return session, nil
}

// Get retrieves a session by ID
func (s *Store) Get(sessionID string) (*Session, bool) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Remove expired session
		s.mu.Lock()
		delete(s.sessions, sessionID)
		s.mu.Unlock()
		return nil, false
	}

	return session, true
}

// Destroy removes a session
func (s *Store) Destroy(sessionID string) {
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
}

// Extend extends a session's expiry time
func (s *Store) Extend(sessionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, exists := s.sessions[sessionID]
	if !exists {
		return false
	}

	// Extend session expiry to 24 hours from now
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	return true
}

// cleanupExpiredSessions periodically removes expired sessions
func (s *Store) cleanupExpiredSessions() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, session := range s.sessions {
			if now.After(session.ExpiresAt) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}