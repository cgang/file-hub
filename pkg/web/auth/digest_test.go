package auth

import (
	"crypto/md5"
	"fmt"
	"testing"
)

func TestGenerateNonce(t *testing.T) {
	nonce1, err := generateNonce()
	if err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	nonce2, err := generateNonce()
	if err != nil {
		t.Fatalf("Failed to generate nonce: %v", err)
	}

	if nonce1 == nonce2 {
		t.Error("Generated nonces should be different")
	}

	if len(nonce1) == 0 {
		t.Error("Generated nonce should not be empty")
	}
}

func TestGenerateOpaque(t *testing.T) {
	opaque1, err := generateOpaque()
	if err != nil {
		t.Fatalf("Failed to generate opaque: %v", err)
	}

	opaque2, err := generateOpaque()
	if err != nil {
		t.Fatalf("Failed to generate opaque: %v", err)
	}

	if opaque1 == opaque2 {
		t.Error("Generated opaques should be different")
	}

	if len(opaque1) == 0 {
		t.Error("Generated opaque should not be empty")
	}
}

func TestCreateDigestChallenge(t *testing.T) {
	challenge, err := createDigestChallenge("test")
	if err != nil {
		t.Fatalf("Failed to create digest challenge: %v", err)
	}

	if challenge.Realm != "test" {
		t.Errorf("Expected realm 'test', got '%s'", challenge.Realm)
	}

	if len(challenge.Nonce) == 0 {
		t.Error("Nonce should not be empty")
	}

	if len(challenge.Opaque) == 0 {
		t.Error("Opaque should not be empty")
	}

	if challenge.Algorithm != "MD5" {
		t.Errorf("Expected algorithm 'MD5', got '%s'", challenge.Algorithm)
	}

	if challenge.QoP != "auth" {
		t.Errorf("Expected qop 'auth', got '%s'", challenge.QoP)
	}
}

func TestParseDigestAuth(t *testing.T) {
	authStr := `Digest username="testuser", realm="test", nonce="abc123", uri="/test", qop="auth", nc=00000001, cnonce="def456", response="responsehash", opaque="opaquevalue", algorithm="MD5"`

	digest, err := parseDigestAuth(authStr)
	if err != nil {
		t.Fatalf("Failed to parse digest auth: %v", err)
	}

	if digest.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", digest.Username)
	}

	if digest.Realm != "test" {
		t.Errorf("Expected realm 'test', got '%s'", digest.Realm)
	}

	if digest.Nonce != "abc123" {
		t.Errorf("Expected nonce 'abc123', got '%s'", digest.Nonce)
	}

	if digest.URI != "/test" {
		t.Errorf("Expected URI '/test', got '%s'", digest.URI)
	}

	if digest.QoP != "auth" {
		t.Errorf("Expected qop 'auth', got '%s'", digest.QoP)
	}

	if digest.NC != "00000001" {
		t.Errorf("Expected nc '00000001', got '%s'", digest.NC)
	}

	if digest.CNonce != "def456" {
		t.Errorf("Expected cnonce 'def456', got '%s'", digest.CNonce)
	}

	if digest.Response != "responsehash" {
		t.Errorf("Expected response 'responsehash', got '%s'", digest.Response)
	}

	if digest.Opaque != "opaquevalue" {
		t.Errorf("Expected opaque 'opaquevalue', got '%s'", digest.Opaque)
	}

	if digest.Algorithm != "MD5" {
		t.Errorf("Expected algorithm 'MD5', got '%s'", digest.Algorithm)
	}
}

func TestCalculateHA1(t *testing.T) {
	ha1 := calculateHA1("testuser", "test", "password")
	// Calculate the expected value
	expected := fmt.Sprintf("%x", md5.Sum([]byte("testuser:test:password")))

	if ha1 != expected {
		t.Errorf("Expected HA1 '%s', got '%s'", expected, ha1)
	}
}

func TestCalculateHA2(t *testing.T) {
	ha2 := calculateHA2("GET", "/test")
	// Calculate the expected value
	expected := fmt.Sprintf("%x", md5.Sum([]byte("GET:/test")))

	if ha2 != expected {
		t.Errorf("Expected HA2 '%s', got '%s'", expected, ha2)
	}
}