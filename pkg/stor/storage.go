package stor

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/cgang/file-hub/pkg/config"
	"github.com/cgang/file-hub/pkg/model"
)

// Global config variable
var (
	globalConfig *config.Config
	configOnce   sync.Once
)

// InitConfig initializes the global config
func InitConfig(cfg *config.Config) {
	configOnce.Do(func() {
		globalConfig = cfg
	})
}

// FileObject represents a file or directory in the storage system
type FileObject struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsDir        bool      `json:"is_dir"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ContentType  string    `json:"content_type"`
}

type Permission int

const (
	PermissionUnknown Permission = iota
	PermissionRead
	PermissionWrite
	PermissionDelete
)

// Storage defines an interface for file operations
type Storage interface {
	CheckPermission(ctx context.Context, path string, user *model.User, perm Permission) error
	// File operations
	/// CreateFile(ctx context.Context, path string) (*FileObject, error)
	DeleteFile(ctx context.Context, path string) error
	CreateDir(ctx context.Context, path string) error
	CopyFile(ctx context.Context, src, dst string) error
	MoveFile(ctx context.Context, src, dst string) error
	GetFileInfo(ctx context.Context, path string) (*FileObject, error)
	OpenFile(ctx context.Context, path string) (io.ReadCloser, error)
	ListDir(ctx context.Context, path string) ([]*FileObject, error)
	WriteToFile(ctx context.Context, path string, content io.Reader) error
}

// ForUser creates a Storage instance for the given user based on their HomeDir
func ForUser(ctx context.Context, user *model.User) (Storage, error) {
	u, err := url.Parse(user.HomeDir)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "s3":
		s3Cfg := globalConfig.S3
		if s3Cfg == nil {
			return nil, errors.New("S3 not configured")
		}
		// Create S3 storage with the parsed bucket and prefix
		bucket := u.Host
		prefix := strings.TrimPrefix(u.Path, "/")
		return newS3Storage(user, globalConfig.S3, bucket, prefix), nil
	case "file", "":
		// Fall through to filesystem storage
		// Fall back to filesystem storage
		return newFsStorage(user, u.Path), nil
	default:
		return nil, fmt.Errorf("unsupported storage scheme: %s", u.Scheme)
	}
}

type storage struct {
	user *model.User
}

func (s *storage) CheckPermission(ctx context.Context, path string, user *model.User, perm Permission) error {
	if s.user.ID == user.ID {
		return nil // Owner has all permissions
	}

	// FIXME add real permission checks based on ACLs
	// For simplicity, we assume all permissions are granted in this example
	// In a real implementation, check the user's permissions for the given path
	return nil
}
