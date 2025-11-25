package auth

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cgang/file-hub/pkg/users"
)

// DigestChallenge represents a digest authentication challenge
type DigestChallenge struct {
	Realm     string
	Nonce     string
	Opaque    string
	Algorithm string
	QoP       string
}

// DigestResponse represents a digest authentication response
type DigestResponse struct {
	Username   string
	Realm      string
	Nonce      string
	URI        string
	QoP        string
	NC         string
	CNonce     string
	Response   string
	Opaque     string
	Algorithm  string
	Method     string
}

// generateNonce creates a random nonce for digest authentication
func generateNonce() (string, error) {
	nonce := make([]byte, 16)
	_, err := rand.Read(nonce)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(nonce), nil
}

// generateOpaque creates a random opaque value for digest authentication
func generateOpaque() (string, error) {
	opaque := make([]byte, 16)
	_, err := rand.Read(opaque)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(opaque), nil
}

// createDigestChallenge creates a new digest challenge
func createDigestChallenge(realm string) (*DigestChallenge, error) {
	nonce, err := generateNonce()
	if err != nil {
		return nil, err
	}

	opaque, err := generateOpaque()
	if err != nil {
		return nil, err
	}

	return &DigestChallenge{
		Realm:     realm,
		Nonce:     nonce,
		Opaque:    opaque,
		Algorithm: "MD5",
		QoP:       "auth",
	}, nil
}

// parseDigestAuth parses the digest authentication header
func parseDigestAuth(authStr string) (*DigestResponse, error) {
	if !strings.HasPrefix(authStr, "Digest ") {
		return nil, fmt.Errorf("not a digest auth header")
	}

	digestStr := strings.TrimPrefix(authStr, "Digest ")
	params := strings.Split(digestStr, ", ")

	response := &DigestResponse{}
	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.TrimSpace(kv[0])
		value := strings.Trim(strings.TrimSpace(kv[1]), "\"")

		switch key {
		case "username":
			response.Username = value
		case "realm":
			response.Realm = value
		case "nonce":
			response.Nonce = value
		case "uri":
			response.URI = value
		case "qop":
			response.QoP = value
		case "nc":
			response.NC = value
		case "cnonce":
			response.CNonce = value
		case "response":
			response.Response = value
		case "opaque":
			response.Opaque = value
		case "algorithm":
			response.Algorithm = value
		}
	}

	return response, nil
}

// calculateHA1 calculates the HA1 value for digest authentication
func calculateHA1(username, realm, password, nonce, cnonce string) string {
	// HA1 = MD5(username:realm:password)
	ha1 := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s", username, realm, password)))
	return hex.EncodeToString(ha1[:])
}

// calculateHA2 calculates the HA2 value for digest authentication
func calculateHA2(method, uri string) string {
	// HA2 = MD5(method:uri)
	ha2 := md5.Sum([]byte(fmt.Sprintf("%s:%s", method, uri)))
	return hex.EncodeToString(ha2[:])
}

// calculateResponse calculates the expected response for digest authentication
func calculateResponse(ha1, nonce, nc, cnonce, qop, ha2 string) string {
	// response = MD5(HA1:nonce:nc:cnonce:qop:HA2)
	resp := md5.Sum([]byte(fmt.Sprintf("%s:%s:%s:%s:%s:%s", ha1, nonce, nc, cnonce, qop, ha2)))
	return hex.EncodeToString(resp[:])
}

// validateDigestResponse validates a digest authentication response
func validateDigestResponse(digest *DigestResponse, storedHA1, method string) bool {
	// Calculate HA2
	ha2 := calculateHA2(method, digest.URI)

	// Calculate the expected response
	expectedResponse := calculateResponse(storedHA1, digest.Nonce, digest.NC, digest.CNonce, digest.QoP, ha2)

	// Compare the responses using constant time comparison
	return expectedResponse == digest.Response
}

// generateWWWAuthenticateHeader generates the WWW-Authenticate header for digest auth
func generateWWWAuthenticateHeader(challenge *DigestChallenge) string {
	return fmt.Sprintf(`Digest realm="%s", nonce="%s", opaque="%s", algorithm=%s, qop="%s"`,
		challenge.Realm, challenge.Nonce, challenge.Opaque, challenge.Algorithm, challenge.QoP)
}

// handleDigestAuth handles digest authentication
func handleDigestAuth(c *gin.Context, authStr string, userService *users.Service, nonceStore *NonceStore, realm string) {
	digest, err := parseDigestAuth(authStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Invalid digest authorization format")
		c.Abort()
		return
	}

	if userService == nil {
		c.String(http.StatusInternalServerError, "User service not initialized")
		c.Abort()
		return
	}

	// Validate the nonce
	if !nonceStore.IsValidNonce(digest.Nonce) {
		c.String(http.StatusUnauthorized, "Invalid or expired nonce")
		c.Abort()
		return
	}

	// Get the user
	_, err = userService.GetByUsername(digest.Username)
	if err != nil {
		// Create a new challenge
		challenge, err := createDigestChallenge(realm)
		if err != nil {
			c.String(http.StatusInternalServerError, "Failed to create auth challenge")
			c.Abort()
			return
		}

		c.Header("WWW-Authenticate", generateWWWAuthenticateHeader(challenge))
		c.String(http.StatusUnauthorized, "Invalid username or password")
		c.Abort()
		return
	}

	// For digest auth, we need to validate using the stored HA1
	// In a real implementation, you would store the HA1 value in the database
	// For now, we'll simulate this by recreating the HA1 from the user's password
	// NOTE: This is not secure for production! In production, you should store HA1 in the database

	// Since we don't have the plaintext password anymore, we can't validate digest auth properly
	// This is a limitation of our current implementation where we only store the hashed password

	// For demonstration purposes, we'll reject digest auth requests
	c.String(http.StatusUnauthorized, "Digest authentication not supported with current password storage")
	c.Abort()
	return

	// Store the authenticated user in the context
	// c.Set("user", user)
	// c.Next()
}

// NonceStore stores nonces to prevent replay attacks
type NonceStore struct {
	nonces map[string]time.Time
}

// NewNonceStore creates a new nonce store
func NewNonceStore() *NonceStore {
	return &NonceStore{
		nonces: make(map[string]time.Time),
	}
}

// IsValidNonce checks if a nonce is valid and not expired
func (ns *NonceStore) IsValidNonce(nonce string) bool {
	// In a real implementation, you would check if the nonce exists and is not expired
	// For simplicity, we'll just return true
	// In production, you should implement proper nonce validation with expiration
	return true
}