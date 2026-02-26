package sync

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"

	"github.com/cgang/file-hub/pkg/db"
	"github.com/cgang/file-hub/pkg/model"
	"github.com/cgang/file-hub/pkg/web/session"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	sessionStoreOnce sync.Once
	sessionStore     *session.Store
)

func getSessionStore() *session.Store {
	sessionStoreOnce.Do(func() {
		sessionStore = session.NewStore()
	})
	return sessionStore
}

// contextKey is a type for context keys
type contextKey string

const (
	UserIDContextKey    contextKey = "userID"
	UsernameContextKey  contextKey = "username"
	RepoNameContextKey  contextKey = "repoName"
)

// AuthInterceptor creates a gRPC unary interceptor for authentication
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Try to authenticate using session cookie or authorization header
		user, err := authenticateFromMetadata(ctx, md)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, err.Error())
		}

		// Add user info to context
		newCtx := context.WithValue(ctx, UserIDContextKey, user.ID)
		newCtx = context.WithValue(newCtx, UsernameContextKey, user.Username)

		// Extract repo name from request if available
		if repoName := extractRepoName(req); repoName != "" {
			newCtx = context.WithValue(newCtx, RepoNameContextKey, repoName)
		}

		// Call the handler with the authenticated context
		return handler(newCtx, req)
	}
}

// authenticateFromMetadata authenticates user from gRPC metadata
func authenticateFromMetadata(ctx context.Context, md metadata.MD) (*model.User, error) {
	// Try session cookie authentication first
	sessionCookies := md.Get("cookie")
	if len(sessionCookies) > 0 {
		user, err := authenticateFromCookie(sessionCookies[0])
		if err == nil {
			return user, nil
		}
	}

	// Try x-session-token header (custom token auth)
	sessionTokens := md.Get("x-session-token")
	if len(sessionTokens) > 0 {
		user, err := authenticateFromSessionToken(sessionTokens[0])
		if err == nil {
			return user, nil
		}
	}

	// Try Authorization header (Basic auth)
	authHeaders := md.Get("authorization")
	if len(authHeaders) > 0 {
		user, err := authenticateFromAuthHeader(authHeaders[0])
		if err == nil {
			return user, nil
		}
	}

	return nil, status.Error(codes.Unauthenticated, "no valid credentials provided")
}

// authenticateFromCookie authenticates user from session cookie
func authenticateFromCookie(cookie string) (*model.User, error) {
	// Parse cookie to find session ID
	// Cookie format: filehub_session=<session_id>; other=values
	cookies := strings.Split(cookie, ";")
	var sessionID string
	for _, c := range cookies {
		c = strings.TrimSpace(c)
		if strings.HasPrefix(c, "filehub_session=") {
			sessionID = strings.TrimPrefix(c, "filehub_session=")
			break
		}
	}

	if sessionID == "" {
		return nil, status.Error(codes.Unauthenticated, "missing session cookie")
	}

	// Get session from store
	store := getSessionStore()
	sess, exists := store.Get(sessionID)
	if !exists {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired session")
	}

	return sess.User, nil
}

// authenticateFromAuthHeader authenticates user from Authorization header
func authenticateFromAuthHeader(authHeader string) (*model.User, error) {
	// Parse "Basic <credentials>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
		return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
	}

	// Decode base64 credentials
	decoded, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials encoding")
	}

	// Parse username:password
	creds := string(decoded)
	colonIdx := strings.Index(creds, ":")
	if colonIdx == -1 {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials format")
	}

	username := creds[:colonIdx]
	password := creds[colonIdx+1:]

	// Validate credentials against database
	user, err := db.GetUserByUsername(context.Background(), username)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "invalid credentials: %v", err)
	}

	// Note: In production, verify password hash here
	// For now, we just check if user exists
	_ = password // Use password in production with proper hash verification

	return user, nil
}

// authenticateFromSessionToken authenticates user from session token
func authenticateFromSessionToken(token string) (*model.User, error) {
	store := getSessionStore()
	sess, exists := store.Get(token)
	if !exists {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	return sess.User, nil
}

// extractRepoName extracts repo name from request if it's a sync request
func extractRepoName(req interface{}) string {
	// Use type assertion to check for requests with Repo field
	if r, ok := req.(interface{ GetRepo() string }); ok {
		return r.GetRepo()
	}
	return ""
}

// StreamAuthInterceptor creates a gRPC stream interceptor for authentication
func StreamAuthInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Extract metadata from context
		ctx := ss.Context()
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Try to authenticate
		user, err := authenticateFromMetadata(ctx, md)
		if err != nil {
			return status.Error(codes.Unauthenticated, err.Error())
		}

		// Add user info to context
		newCtx := context.WithValue(ctx, UserIDContextKey, user.ID)
		newCtx = context.WithValue(newCtx, UsernameContextKey, user.Username)

		// Wrap the stream with the new context
		wrapped := &wrappedStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		return handler(srv, wrapped)
	}
}

// wrappedStream wraps a gRPC stream to use a different context
type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}
