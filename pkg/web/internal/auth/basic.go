package auth

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/cgang/file-hub/pkg/users"
	"github.com/gin-gonic/gin"
)

// parseBasicAuth parses the basic auth credentials from the encoded string
func parseBasicAuth(encoded string) (username, password string, ok bool) {
	// Decode the base64 encoded credentials
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", "", false
	}

	// Split the credentials on the first colon
	cs := string(decoded)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return "", "", false
	}

	// Return the username and password
	return cs[:s], cs[s+1:], true
}

// handleBasicAuth handles basic authentication
func handleBasicAuth(c *gin.Context, creds string, realm string) {
	username, password, ok := parseBasicAuth(creds)
	if !ok {
		c.Header("WWW-Authenticate", `Basic realm="`+realm+`"`)
		c.String(http.StatusUnauthorized, "Invalid authorization format")
		c.Abort()
		return
	}

	// Authenticate the user using our user service
	user, err := users.Authenticate(c, username, password)
	if err != nil {
		c.Header("WWW-Authenticate", `Basic realm="`+realm+`"`)
		c.String(http.StatusUnauthorized, "Invalid username or password")
		c.Abort()
		return
	}

	// Store the authenticated user in the context
	c.Set("user", user)
	c.Next()
}
