package stor

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileMeta(t *testing.T) {
	t.Run("newFileMeta creates file metadata", func(t *testing.T) {
		now := time.Now()
		meta := newFileMeta("/docs/readme.txt", now)

		assert.Equal(t, "readme.txt", meta.Name)
		assert.Equal(t, "/docs/readme.txt", meta.Path)
		assert.False(t, meta.IsDir)
		assert.Equal(t, int64(0), meta.Size)
		assert.Equal(t, now, meta.ModTime)
	})

	t.Run("newDirMeta creates directory metadata", func(t *testing.T) {
		now := time.Now()
		meta := newDirMeta("/docs/reports", now)

		assert.Equal(t, "reports", meta.Name)
		assert.Equal(t, "/docs/reports", meta.Path)
		assert.True(t, meta.IsDir)
		assert.Equal(t, int64(0), meta.Size)
		assert.Equal(t, now, meta.ModTime)
	})

	t.Run("toObject converts to FileObject", func(t *testing.T) {
		now := time.Now()
		meta := &FileMeta{
			Name:    "test.txt",
			Path:    "/test.txt",
			IsDir:   false,
			Size:    1024,
			ModTime: now,
		}

		obj := meta.toObject(10, 42, 5)

		assert.Equal(t, 10, obj.RepoID)
		assert.Equal(t, 42, obj.OwnerID)
		assert.Equal(t, 5, obj.ParentID)
		assert.Equal(t, "test.txt", obj.Name)
		assert.Equal(t, "/test.txt", obj.Path)
		assert.Equal(t, int64(1024), obj.Size)
		assert.Equal(t, now, obj.ModTime)
		assert.False(t, obj.IsDir)
	})
}

func TestGetContentType(t *testing.T) {
	t.Run("Known file extensions", func(t *testing.T) {
		tests := []struct {
			ext      string
			expected string
		}{
			{".txt", "text/plain"},
			{".TXT", "text/plain"},
			{".html", "text/html"},
			{".htm", "text/html"},
			{".css", "text/css"},
			{".js", "application/javascript"},
			{".json", "application/json"},
			{".png", "image/png"},
			{".jpg", "image/jpeg"},
			{".jpeg", "image/jpeg"},
			{".gif", "image/gif"},
		}

		for _, test := range tests {
			t.Run(test.ext, func(t *testing.T) {
				result := getContentType(test.ext)
				assert.Equal(t, test.expected, result)
			})
		}
	})

	t.Run("Unknown file extension", func(t *testing.T) {
		result := getContentType(".unknown")
		assert.Equal(t, "application/octet-stream", result)
	})

	t.Run("Empty extension", func(t *testing.T) {
		result := getContentType("")
		assert.Equal(t, "application/octet-stream", result)
	})
}

func TestFsStorage(t *testing.T) {
	t.Run("getFullPath combines root and path", func(t *testing.T) {
		storage := &fsStorage{rootDir: "/data"}

		assert.Equal(t, "/data/repo/file.txt", storage.getFullPath("repo", "file.txt"))
		assert.Equal(t, "/data/repo/docs/readme.txt", storage.getFullPath("repo", "/docs/readme.txt"))
	})

	t.Run("getFullPath cleans path", func(t *testing.T) {
		storage := &fsStorage{rootDir: "/data"}

		// Path traversal should be cleaned
		fullPath := storage.getFullPath("repo", "../etc/passwd")
		assert.NotContains(t, fullPath, "..")
	})

	t.Run("getFullPath handles root path", func(t *testing.T) {
		storage := &fsStorage{rootDir: "/data"}

		assert.Equal(t, "/data/repo", storage.getFullPath("repo", "/"))
		assert.Equal(t, "/data/repo", storage.getFullPath("repo", ""))
	})
}

func TestIsConfiguredRoot(t *testing.T) {
	t.Run("Configured root returns true", func(t *testing.T) {
		originalRoots := rootDirs
		defer func() { rootDirs = originalRoots }()

		rootDirs = []string{"/data1", "/data2", "/data3"}

		assert.True(t, isConfiguredRoot("/data1"))
		assert.True(t, isConfiguredRoot("/data2"))
		assert.True(t, isConfiguredRoot("/data3"))
	})

	t.Run("Unconfigured root returns false", func(t *testing.T) {
		originalRoots := rootDirs
		defer func() { rootDirs = originalRoots }()

		rootDirs = []string{"/data1", "/data2"}

		assert.False(t, isConfiguredRoot("/data3"))
		assert.False(t, isConfiguredRoot("/other"))
	})
}

func TestValidRoot(t *testing.T) {
	t.Run("Existing directory returns true", func(t *testing.T) {
		originalRoots := rootDirs
		defer func() { rootDirs = originalRoots }()

		tmpDir := t.TempDir()
		rootDirs = []string{tmpDir}

		assert.True(t, ValidRoot(tmpDir))
	})

	t.Run("Non-existent directory gets created", func(t *testing.T) {
		originalRoots := rootDirs
		defer func() { rootDirs = originalRoots }()

		baseDir := t.TempDir()
		newDir := filepath.Join(baseDir, "new-root")
		rootDirs = []string{baseDir, newDir}

		assert.True(t, ValidRoot(newDir))

		// Verify directory was created
		_, err := os.Stat(newDir)
		assert.NoError(t, err)
	})

	t.Run("Unconfigured root returns false", func(t *testing.T) {
		originalRoots := rootDirs
		defer func() { rootDirs = originalRoots }()

		rootDirs = []string{"/data"}

		assert.False(t, ValidRoot("/other"))
	})

	t.Run("Path is cleaned", func(t *testing.T) {
		originalRoots := rootDirs
		defer func() { rootDirs = originalRoots }()

		tmpDir := t.TempDir()
		rootDirs = []string{tmpDir}

		// Path with trailing slash should be cleaned
		assert.True(t, ValidRoot(tmpDir+"/"))
	})
}

func TestS3KeyGeneration(t *testing.T) {
	t.Run("getS3Key generates consistent keys", func(t *testing.T) {
		storage := &s3Storage{bucket: "my-bucket"}

		key1 := storage.getS3Key("repo1", "/docs/file.txt")
		key2 := storage.getS3Key("repo1", "/docs/file.txt")

		assert.Equal(t, key1, key2, "Same path should generate same key")
	})

	t.Run("getS3Key generates different keys for different paths", func(t *testing.T) {
		storage := &s3Storage{bucket: "my-bucket"}

		key1 := storage.getS3Key("repo1", "/docs/file1.txt")
		key2 := storage.getS3Key("repo1", "/docs/file2.txt")

		assert.NotEqual(t, key1, key2, "Different paths should generate different keys")
	})

	t.Run("getS3Key uses hash prefix", func(t *testing.T) {
		storage := &s3Storage{bucket: "my-bucket"}

		key := storage.getS3Key("repo", "/path/file.txt")

		// Key should start with a hash prefix (2 characters = 1 byte from SHA256)
		parts := strings.Split(key, "/")
		assert.Greater(t, len(parts[0]), 0, "Key should have a prefix")
		assert.LessOrEqual(t, len(parts[0]), 4, "Prefix should be short")
	})

	t.Run("getS3Key cleans path", func(t *testing.T) {
		storage := &s3Storage{bucket: "my-bucket"}

		key1 := storage.getS3Key("repo", "/docs/../file.txt")
		key2 := storage.getS3Key("repo", "/file.txt")

		assert.Equal(t, key1, key2, "Cleaned paths should generate same key")
	})

	t.Run("getS3Key handles root path", func(t *testing.T) {
		storage := &s3Storage{bucket: "my-bucket"}

		key := storage.getS3Key("repo", "/")
		assert.Contains(t, key, "repo")
	})
}

func TestHashPrefix(t *testing.T) {
	t.Run("hashPrefix generates consistent prefixes", func(t *testing.T) {
		prefix1 := hashPrefix("test-string")
		prefix2 := hashPrefix("test-string")

		assert.Equal(t, prefix1, prefix2, "Same input should produce same prefix")
	})

	t.Run("hashPrefix generates different prefixes for different inputs", func(t *testing.T) {
		prefix1 := hashPrefix("string1")
		prefix2 := hashPrefix("string2")

		assert.NotEqual(t, prefix1, prefix2, "Different inputs should produce different prefixes")
	})

	t.Run("hashPrefix returns 4 character hex string", func(t *testing.T) {
		prefix := hashPrefix("test")
		assert.Len(t, prefix, 4, "Prefix should be 4 hex characters (2 bytes)")
	})

	t.Run("hashPrefix with empty string", func(t *testing.T) {
		prefix := hashPrefix("")
		assert.NotEmpty(t, prefix, "Empty string should still produce a prefix")
		assert.Len(t, prefix, 4)
	})
}

func TestStorageInterface(t *testing.T) {
	t.Run("fsStorage implements Storage interface", func(t *testing.T) {
		var _ Storage = (*fsStorage)(nil)
	})

	t.Run("s3Storage implements Storage interface", func(t *testing.T) {
		var _ Storage = (*s3Storage)(nil)
	})
}

func TestPathOperations(t *testing.T) {
	t.Run("path.Base extraction", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"/docs/file.txt", "file.txt"},
			{"/", "/"},          // path.Base returns "/" for root
			{"", "."},           // path.Base returns "." for empty
			{"file.txt", "file.txt"},
			{"/deep/nested/path/file.pdf", "file.pdf"},
		}

		for _, test := range tests {
			t.Run(test.input, func(t *testing.T) {
				result := path.Base(test.input)
				assert.Equal(t, test.expected, result)
			})
		}
	})

	t.Run("path.Clean normalization", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"/docs/../file.txt", "/file.txt"},
			{"/docs/./file.txt", "/docs/file.txt"},
			{"/docs//file.txt", "/docs/file.txt"},
			{"/", "/"},
			{"", "."},
		}

		for _, test := range tests {
			t.Run(test.input, func(t *testing.T) {
				result := path.Clean(test.input)
				assert.Equal(t, test.expected, result)
			})
		}
	})

	t.Run("path.Dir extraction", func(t *testing.T) {
		tests := []struct {
			input    string
			expected string
		}{
			{"/docs/file.txt", "/docs"},
			{"/file.txt", "/"},
			{"/", "/"},
			{"file.txt", "."},
		}

		for _, test := range tests {
			t.Run(test.input, func(t *testing.T) {
				result := path.Dir(test.input)
				assert.Equal(t, test.expected, result)
			})
		}
	})
}

func TestFileMetaContentType(t *testing.T) {
	t.Run("FileObject ContentType for file with mime type", func(t *testing.T) {
		mimeType := "text/plain"
		obj := &FileMeta{
			Name:  "test.txt",
			Path:  "/test.txt",
			IsDir: false,
		}

		// FileMeta doesn't have ContentType method, but converts to FileObject which does
		fileObj := obj.toObject(1, 1, 0)
		fileObj.MimeType = &mimeType

		assert.Equal(t, "text/plain", fileObj.ContentType())
	})

	t.Run("FileObject ContentType for directory", func(t *testing.T) {
		obj := &FileMeta{
			Name:  "docs",
			Path:  "/docs",
			IsDir: true,
		}

		fileObj := obj.toObject(1, 1, 0)
		assert.Equal(t, "httpd/unix-directory", fileObj.ContentType())
	})
}

func TestContextUsage(t *testing.T) {
	t.Run("Context cancellation propagation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		assert.Equal(t, context.Canceled, ctx.Err())
	})

	t.Run("Context with timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		select {
		case <-ctx.Done():
			assert.Error(t, ctx.Err())
		case <-time.After(200 * time.Millisecond):
			t.Error("Context should have timed out")
		}
	})
}

func TestErrorConditions(t *testing.T) {
	t.Run("IsNotFound checks for sql.ErrNoRows", func(t *testing.T) {
		// IsNotFound uses errors.Is() to check for sql.ErrNoRows
		// os.ErrNotExist is different from sql.ErrNoRows
		assert.False(t, IsNotFound(nil))
		assert.False(t, IsNotFound(os.ErrExist))
		assert.False(t, IsNotFound(os.ErrNotExist))
	})
}
