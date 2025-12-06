package auth

import (
	"net/http"
	"strings"

	"github.com/cgang/file-hub/pkg/model"
	"github.com/cgang/file-hub/pkg/web/session"
	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "filehub_session"
)

var (
	nonceStore   = NewNonceStore()
	SessionStore *session.Store
	userRealm    string
)

func Init(store *session.Store, realm string) {
	SessionStore = store
	userRealm = realm
}

// Authenticate handles authentication with support for sessions
func Authenticate(c *gin.Context) {
	// First, check if there's a valid session cookie
	sessionID, err := c.Cookie(SessionCookieName)
	if err == nil && SessionStore != nil {
		if sess, ok := SessionStore.Get(sessionID); ok {
			// Valid session found, set user in context and continue
			c.Set("user", sess.User)
			c.Next()
			return
		}
	}

	// No valid session, check for Authorization header
	authStr := c.GetHeader("Authorization")
	if authStr == "" {
		// Create a digest challenge
		challenge, err := createDigestChallenge(userRealm)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create auth challenge")
			c.Abort()
			return
		}

		// Also support basic auth
		c.Header("WWW-Authenticate", `Basic realm="`+userRealm+`"`)
		c.Header("WWW-Authenticate", generateWWWAuthenticateHeader(challenge))
		c.String(http.StatusUnauthorized, "No authorization provided")
		c.Abort()
		return
	}

	kind, creds, ok := strings.Cut(authStr, " ")
	if !ok {
		c.String(http.StatusBadRequest, "Invalid authorization header")
		c.Abort()
		return
	}

	switch kind {
	case "Basic":
		handleBasicAuth(c, creds, userRealm)
	case "Digest":
		handleDigestAuth(c, authStr, nonceStore, userRealm)
	default:
		c.String(http.StatusBadRequest, "Unsupported authorization method")
		c.Abort()
		return
	}
}

// GetAuthenticatedUser retrieves the authenticated user from the context
func GetAuthenticatedUser(c *gin.Context) (*model.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	u, ok := user.(*model.User)
	return u, ok
}

// CreateSession creates a new session for the user and sets a cookie
func CreateSession(c *gin.Context, user *model.User) error {
	if SessionStore == nil {
		return nil // Sessions not enabled
	}

	session, err := SessionStore.Create(user)
	if err != nil {
		return err
	}

	// Set cookie with session ID
	c.SetCookie(SessionCookieName, session.ID, 24*3600, "/", "", false, true)
	return nil
}

// DestroySession destroys the current session
func DestroySession(c *gin.Context) {
	if SessionStore == nil {
		return // Sessions not enabled
	}

	sessionID, err := c.Cookie(SessionCookieName)
	if err == nil {
		SessionStore.Destroy(sessionID)
	}

	// Clear the session cookie
	c.SetCookie(SessionCookieName, "", -1, "/", "", false, true)
}
