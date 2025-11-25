package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/cgang/file-hub/pkg/users"
)

// UserService is the user service instance used for authentication
var UserService *users.Service

// NonceStore stores nonces for digest authentication
var nonceStore = NewNonceStore()

// Realm is the authentication realm
const Realm = "FileHub"

// SetUserService sets the user service instance for authentication
func SetUserService(service *users.Service) {
	UserService = service
}

func Authenticate(c *gin.Context) {
	authStr := c.GetHeader("Authorization")
	if authStr == "" {
		// Create a digest challenge
		challenge, err := createDigestChallenge(Realm)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create auth challenge")
			c.Abort()
			return
		}

		// Also support basic auth
		c.Header("WWW-Authenticate", `Basic realm="`+Realm+`"`)
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
		handleBasicAuth(c, creds, UserService, Realm)
	case "Digest":
		handleDigestAuth(c, authStr, UserService, nonceStore, Realm)
	default:
		c.String(http.StatusBadRequest, "Unsupported authorization method")
		c.Abort()
		return
	}
}

// GetAuthenticatedUser retrieves the authenticated user from the context
func GetAuthenticatedUser(c *gin.Context) (*users.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	u, ok := user.(*users.User)
	return u, ok
}
