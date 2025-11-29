package stor

import (
	"context"
	"io"
	"time"

	"github.com/cgang/file-hub/pkg/model"
)

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

func ForUser(ctx context.Context, user *model.User) (Storage, error) {
	// For simplicity, we use filesystem storage for all users in this example
	return newFsStorage(user), nil
}
