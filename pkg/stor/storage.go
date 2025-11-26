package stor

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cgang/file-hub/pkg/users"
)

// File represents a file or directory in the storage system
type File struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsDir        bool      `json:"is_dir"`
	Size         int64     `json:"size"`
	LastModified time.Time `json:"last_modified"`
	ContentType  string    `json:"content_type"`
}

// Storage defines an interface for file operations
type Storage interface {
	// File operations
	CreateFile(ctx context.Context, user *users.User, path string) (*os.File, error)
	DeleteFile(ctx context.Context, user *users.User, path string) error
	CreateDir(ctx context.Context, user *users.User, path string) error
	CopyFile(ctx context.Context, user *users.User, src, dst string) error
	MoveFile(ctx context.Context, user *users.User, src, dst string) error
	GetFileInfo(ctx context.Context, user *users.User, path string) (*File, error)
	OpenFile(ctx context.Context, user *users.User, path string) (io.ReadCloser, error)
	ListDir(ctx context.Context, user *users.User, path string) ([]*File, error)
	WriteToFile(ctx context.Context, user *users.User, path string, content io.Reader) error
}

func NewStorage(users *users.Service) Storage {
	return &OsStorage{users: users}
}

// OsStorage implements Storage using standard OS operations
type OsStorage struct {
	users *users.Service
}

// getFullPath combines the user's home directory with the relative path
func (s *OsStorage) getFullPath(user *users.User, path string) string {
	// Clean the path to prevent directory traversal attacks
	cleanPath := filepath.Clean(path)
	
	// Join with user's home directory
	fullPath := filepath.Join(user.HomeDir, cleanPath)
	
	// Ensure the path is still within the user's home directory
	// This prevents directory traversal attacks
	if !strings.HasPrefix(fullPath, user.HomeDir+string(filepath.Separator)) && fullPath != user.HomeDir {
		// If the path is outside the user's home directory, default to the home directory
		return user.HomeDir
	}
	
	return fullPath
}

func (s *OsStorage) CreateFile(ctx context.Context, user *users.User, path string) (*os.File, error) {
	fullPath := s.getFullPath(user, path)
	return os.Create(fullPath)
}

func (s *OsStorage) DeleteFile(ctx context.Context, user *users.User, path string) error {
	fullPath := s.getFullPath(user, path)
	return os.RemoveAll(fullPath)
}

func (s *OsStorage) CreateDir(ctx context.Context, user *users.User, path string) error {
	fullPath := s.getFullPath(user, path)
	return os.MkdirAll(fullPath, 0755)
}

func (s *OsStorage) CopyFile(ctx context.Context, user *users.User, src, dst string) error {
	// Handle directory or file copy
	srcFullPath := s.getFullPath(user, src)
	dstFullPath := s.getFullPath(user, dst)

	srcInfo, err := os.Stat(srcFullPath)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(ctx, srcFullPath, dstFullPath)
	}
	return copyFile(ctx, srcFullPath, dstFullPath)
}

func (s *OsStorage) MoveFile(ctx context.Context, user *users.User, src, dst string) error {
	srcFullPath := s.getFullPath(user, src)
	dstFullPath := s.getFullPath(user, dst)
	return os.Rename(srcFullPath, dstFullPath)
}

func (s *OsStorage) GetFileInfo(ctx context.Context, user *users.User, path string) (*File, error) {
	fullPath := s.getFullPath(user, path)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	file := &File{
		Name:         fileInfo.Name(),
		Path:         path,
		IsDir:        fileInfo.IsDir(),
		Size:         fileInfo.Size(),
		LastModified: fileInfo.ModTime(),
		ContentType:  "application/octet-stream",
	}

	if !file.IsDir {
		if ext := filepath.Ext(path); ext != "" {
			contentType := getContentType(ext)
			if contentType != "" {
				file.ContentType = contentType
			}
		}
	}

	return file, nil
}

func (s *OsStorage) ListDir(ctx context.Context, user *users.User, path string) ([]*File, error) {
	fullPath := s.getFullPath(user, path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	files := make([]*File, 0, len(entries))
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		file, err := s.GetFileInfo(ctx, user, entryPath)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func (s *OsStorage) OpenFile(ctx context.Context, user *users.User, path string) (io.ReadCloser, error) {
	fullPath := s.getFullPath(user, path)
	return os.Open(fullPath)
}

func (s *OsStorage) WriteToFile(ctx context.Context, user *users.User, path string, content io.Reader) error {
	// Create directory if needed
	fullPath := s.getFullPath(user, path)
	dirPath := filepath.Dir(fullPath)
	if err := s.CreateDir(ctx, user, dirPath); err != nil {
		return err
	}

	// Open file for writing
	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Copy content
	_, err = io.Copy(file, content)
	return err
}

// Helper functions
func copyFile(ctx context.Context, src, dst string) error {
	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func copyDir(ctx context.Context, src, dst string) error {
	// Create destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read directory contents
	files, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	// Copy each file/dir
	for _, file := range files {
		srcPath := filepath.Join(src, file.Name())
		dstPath := filepath.Join(dst, file.Name())

		if file.IsDir() {
			if err := copyDir(ctx, srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(ctx, srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func getContentType(ext string) string {
	// Simplified content type mapping - in real implementation use mime package
	switch strings.ToLower(ext) {
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	default:
		return "application/octet-stream"
	}
}