package stor

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
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
	CreateFile(path string) (*os.File, error)
	DeleteFile(path string) error
	CreateDir(path string) error
	CopyFile(src, dst string) error
	MoveFile(src, dst string) error
	GetFileInfo(path string) (*File, error)
	OpenFile(path string) (io.ReadCloser, error)
	ListDir(path string) ([]*File, error)
	WriteToFile(path string, content io.Reader) error
}

func NewStorage(rootDir string) Storage {
	return &OsStorage{rootDir: rootDir}
}

// OsStorage implements Storage using standard OS operations
type OsStorage struct {
	rootDir string
}

func (s *OsStorage) CreateFile(path string) (*os.File, error) {
	fullPath := filepath.Join(s.rootDir, path)
	return os.Create(fullPath)
}

func (s *OsStorage) DeleteFile(path string) error {
	fullPath := filepath.Join(s.rootDir, path)
	return os.RemoveAll(fullPath)
}

func (s *OsStorage) CreateDir(path string) error {
	fullPath := filepath.Join(s.rootDir, path)
	return os.MkdirAll(fullPath, 0755)
}

func (s *OsStorage) CopyFile(src, dst string) error {
	// Handle directory or file copy
	srcFullPath := filepath.Join(s.rootDir, src)
	dstFullPath := filepath.Join(s.rootDir, dst)

	srcInfo, err := os.Stat(srcFullPath)
	if err != nil {
		return err
	}

	if srcInfo.IsDir() {
		return copyDir(srcFullPath, dstFullPath)
	}
	return copyFile(srcFullPath, dstFullPath)
}

func (s *OsStorage) MoveFile(src, dst string) error {
	srcFullPath := filepath.Join(s.rootDir, src)
	dstFullPath := filepath.Join(s.rootDir, dst)
	return os.Rename(srcFullPath, dstFullPath)
}

func (s *OsStorage) GetFileInfo(path string) (*File, error) {
	fullPath := filepath.Join(s.rootDir, path)
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

func (s *OsStorage) ListDir(path string) ([]*File, error) {
	fullPath := filepath.Join(s.rootDir, path)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	files := make([]*File, 0, len(entries))
	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		file, err := s.GetFileInfo(entryPath)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	return files, nil
}

func (s *OsStorage) OpenFile(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.rootDir, path)
	return os.Open(fullPath)
}

func (s *OsStorage) WriteToFile(path string, content io.Reader) error {
	// Create directory if needed
	fullPath := filepath.Join(s.rootDir, path)
	dirPath := filepath.Dir(fullPath)
	if err := s.CreateDir(dirPath); err != nil {
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
func copyFile(src, dst string) error {
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

func copyDir(src, dst string) error {
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
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
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
