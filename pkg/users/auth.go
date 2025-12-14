package users

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
)

func ComputeMD5(format string, values ...any) string {
	sum := md5.Sum(fmt.Appendf(nil, format, values...))
	return hex.EncodeToString(sum[:])
}

// calculateHA1 calculates the HA1 value for digest authentication
func calculateHA1(username, password string) string {
	return ComputeMD5("%s:%s:%s", username, userRealm, password)
}

// Authenticate validates a user's credentials for basic authentication
func Authenticate(ctx context.Context, username, password string) (*model.User, error) {
	// Get user by username
	user, err := db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Calculate HA1 hash from provided credentials
	providedHA1 := calculateHA1(username, password)

	// Compare hashes using constant time comparison
	if user.HA1 != providedHA1 {
		return nil, errors.New("invalid credentials")
	}

	updateLastLogin(context.Background(), user)

	return user, nil
}

func updateLastLogin(ctx context.Context, user *model.User) {
	// Update last login time
	now := time.Now()
	updateReq := &UpdateUserRequest{
		LastLogin: &now,
	}

	err := Update(ctx, user.ID, updateReq)
	if err != nil {
		log.Printf("Failed to update last login time: %s", err)
	}
}

// ValidateDigest validates a user's credentials for digest authentication
func ValidateDigest(ctx context.Context, username, uri, nonce, nc, cnonce, qop, response, method string) (*model.User, error) {
	// Get user by username
	user, err := db.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Calculate HA2
	ha2 := ComputeMD5("%s:%s", method, uri)

	// Calculate the expected response using the stored HA1
	expectedResponse := ComputeMD5("%s:%s:%s:%s:%s:%s", user.HA1, nonce, nc, cnonce, qop, ha2)
	// Compare responses using constant time comparison
	if response != expectedResponse {
		return nil, errors.New("invalid credentials")
	}

	updateLastLogin(context.Background(), user)

	return user, nil
}
