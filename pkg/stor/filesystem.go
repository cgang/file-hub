package stor

import (
	"context"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// fsStorage implements Storage based on the local filesystem
type fsStorage struct {
	rootDir string
}

// getFullPath combines the user's home directory with the relative path
func (s *fsStorage) getFullPath(repo, name string) string {
	return path.Join(s.rootDir, repo, path.Clean(name))
}

func (s *fsStorage) PutFile(ctx context.Context, repo, name string, data io.Reader) error {
	fullPath := s.getFullPath(repo, name)

	dir := path.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, data)
	return err
}

func (s *fsStorage) DeleteFile(ctx context.Context, repo, name string) error {
	fullPath := s.getFullPath(repo, name)
	return os.Remove(fullPath)
}

func (s *fsStorage) OpenFile(ctx context.Context, repo, name string) (io.ReadCloser, error) {
	fullPath := s.getFullPath(repo, name)
	return os.Open(fullPath)
}

func (s *fsStorage) CopyFile(ctx context.Context, repo, srcName, destName string) error {
	srcPath := s.getFullPath(repo, srcName)

	input, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer input.Close()

	return s.PutFile(ctx, repo, destName, input)
}

func (s *fsStorage) Scan(ctx context.Context, repo string, visit func(*FileMeta) error) error {
	rootDir := s.getFullPath(repo, "")

	return filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error occurs while walk to %s: %s", path, err)
			return filepath.SkipDir
		}

		meta := &FileMeta{
			Name:  d.Name(),
			Path:  strings.TrimPrefix(path, rootDir),
			IsDir: d.IsDir(),
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		meta.LastModified = info.ModTime()
		if !d.IsDir() {
			meta.Size = info.Size()
		}

		return visit(meta)
	})
}

func (s *fsStorage) GetContentType(ctx context.Context, repo, name string) (string, error) {
	return getContentType(path.Ext(name)), nil
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
