package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/cgang/file-hub/pkg/users"
)

// UserService is the user service instance used for authentication
var UserService *users.Service

// SetUserService sets the user service instance for authentication
func SetUserService(service *users.Service) {
	UserService = service
}

func Authenticate(c *gin.Context) {
	authStr := c.GetHeader("Authorization")
	if authStr == "" {
		c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
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
		username, password, ok := parseBasicAuth(creds)
		if !ok {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.String(http.StatusUnauthorized, "Invalid authorization format")
			c.Abort()
			return
		}

		if UserService == nil {
			c.String(http.StatusInternalServerError, "User service not initialized")
			c.Abort()
			return
		}

		// Authenticate the user using our user service
		user, err := UserService.Authenticate(username, password)
		if err != nil {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.String(http.StatusUnauthorized, "Invalid username or password")
			c.Abort()
			return
		}

		// Store the authenticated user in the context
		c.Set("user", user)
	default:
		c.String(http.StatusBadRequest, "Unsupported authorization method")
		c.Abort()
		return
	}

	c.Next()
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
