package auth

import (
	"encoding/base64"
	"strings"
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
