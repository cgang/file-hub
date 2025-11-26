package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cgang/file-hub/pkg/session"
	"github.com/cgang/file-hub/pkg/users"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSessionMiddleware(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Save original globals
	originalSessionStore := SessionStore
	defer func() {
		SessionStore = originalSessionStore
	}()

	// Create a session store
	sessionStore := session.NewStore()
	SessionStore = sessionStore

	// Create a test user
	user := &users.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Create a session
	session, _ := sessionStore.Create(user)

	// Create a test router with session middleware
	router := gin.New()
	router.Use(Authenticate)
	router.GET("/protected", func(c *gin.Context) {
		user, exists := GetAuthenticatedUser(c)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user": user.Username})
	})

	// Test accessing protected route with valid session
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: session.ID,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogoutHandler(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Save original globals
	originalSessionStore := SessionStore
	defer func() {
		SessionStore = originalSessionStore
	}()

	// Create a session store
	sessionStore := session.NewStore()
	SessionStore = sessionStore

	// Create a test user
	user := &users.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Create a session
	session, _ := sessionStore.Create(user)

	// Create a test router with logout handler
	router := gin.New()
	router.POST("/logout", LogoutHandler)

	// Test logout with valid session
	req, _ := http.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  SessionCookieName,
		Value: session.ID,
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify session was destroyed
	_, ok := sessionStore.Get(session.ID)
	assert.False(t, ok)
}