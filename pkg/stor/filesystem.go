package stor

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cgang/file-hub/pkg/model"
)

// fsStorage implements Storage based on the local filesystem
type fsStorage struct {
	storage
	rootDir string
}

func newFsStorage(user *model.User, rootDir string) *fsStorage {
	return &fsStorage{
		storage: storage{user},
		rootDir: rootDir,
	}
}

// getFullPath combines the user's home directory with the relative path
func (s *fsStorage) getFullPath(path string) string {
	cleanPath := filepath.Clean(path)
	fullPath := filepath.Join(s.rootDir, cleanPath)

	// Ensure the path is still within the user's home directory
	// This prevents directory traversal attacks
	if !strings.HasPrefix(fullPath, s.rootDir+string(filepath.Separator)) && fullPath != s.rootDir {
		// If the path is outside the user's home directory, default to the home directory
		return s.rootDir
	}

	return fullPath
}

func (s *fsStorage) CreateFile(ctx context.Context, path string) (*os.File, error) {
	fullPath := s.getFullPath(path)
	return os.Create(fullPath)
}

func (s *fsStorage) DeleteFile(ctx context.Context, path string) error {
	fullPath := s.getFullPath(path)
	return os.RemoveAll(fullPath)
}

func (s *fsStorage) CreateDir(ctx context.Context, path string) error {
	fullPath := s.getFullPath(path)
	return os.MkdirAll(fullPath, 0755)
}

func (s *fsStorage) CopyFile(ctx context.Context, src, dst string) error {
	// Handle directory or file copy
	srcFullPath := s.getFullPath(src)
	dstFullPath := s.getFullPath(dst)

	srcInfo, err := os.Stat(srcFullPath)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(ctx, srcFullPath, dstFullPath)
	}
	return copyFile(ctx, srcFullPath, dstFullPath)
}

func (s *fsStorage) MoveFile(ctx context.Context, src, dst string) error {
	srcFullPath := s.getFullPath(src)
	dstFullPath := s.getFullPath(dst)
	return os.Rename(srcFullPath, dstFullPath)
}

func (s *fsStorage) GetFileInfo(ctx context.Context, path string) (*FileObject, error) {
	fullPath := s.getFullPath(path)
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}

	file := &FileObject{
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

func (s *fsStorage) ListDir(ctx context.Context, path string) ([]*FileObject, error) {
	fullPath := s.getFullPath(path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	files := make([]*FileObject, 0, len(entries))
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		file, err := s.GetFileInfo(ctx, entryPath)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func (s *fsStorage) OpenFile(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := s.getFullPath(path)
	return os.Open(fullPath)
}

func (s *fsStorage) WriteToFile(ctx context.Context, path string, content io.Reader) error {
	// Create directory if needed
	fullPath := s.getFullPath(path)
	dirPath := filepath.Dir(fullPath)
	if err := s.CreateDir(ctx, dirPath); err != nil {
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
