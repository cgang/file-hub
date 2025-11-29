package stor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cgang/file-hub/pkg/model"
)

// mockUserService creates a mock user service for testing
func mockUserService() interface{} {
	return nil // We won't actually use the service in these tests
}

// mockUser creates a mock user for testing
func mockUser(homeDir string) *model.User {
	return &model.User{
		ID:        1,
		Username:  "testuser",
		Email:     "test@example.com",
		FirstName: stringPtr("Test"),
		LastName:  stringPtr("User"),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		IsActive:  true,
		IsAdmin:   false,
		HomeDir:   homeDir,
	}
}

func stringPtr(s string) *string {
	return &s
}

func setupTest(t *testing.T) (*OsStorage, string) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create storage with mock user
	user := mockUser(tempDir)
	storage := newFsStorage(user)

	return storage, tempDir
}

func TestOsStorage_CreateFile(t *testing.T) {
	storage, _ := setupTest(t)
	filePath := "test.txt"

	// Create a file
	file, err := storage.CreateFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	defer file.Close()

	// Write some content
	content := "test content"
	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
}

func TestOsStorage_GetFileInfo(t *testing.T) {
	storage, _ := setupTest(t)
	filePath := "test_info.txt"

	// Create a file
	file, err := storage.CreateFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	defer file.Close()

	// Write some content
	content := "test info content"
	_, err = file.WriteString(content)
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	// Get file info
	info, err := storage.GetFileInfo(context.Background(), filePath)
	if err != nil {
		t.Fatalf("GetFileInfo failed: %v", err)
	}

	if info.Name != "test_info.txt" {
		t.Errorf("Expected file name 'test_info.txt', got '%s'", info.Name)
	}

	if info.Size != int64(len(content)) {
		t.Errorf("Expected file size %d, got %d", len(content), info.Size)
	}
}

func TestOsStorage_CreateDir(t *testing.T) {
	storage, _ := setupTest(t)
	dirPath := "test_dir"

	// Create a directory
	err := storage.CreateDir(context.Background(), dirPath)
	if err != nil {
		t.Fatalf("CreateDir failed: %v", err)
	}
}

func TestOsStorage_CopyFile(t *testing.T) {
	storage, tempDir := setupTest(t)
	srcPath := "source.txt"
	dstPath := "destination.txt"

	// Create source file with content
	srcFile, err := storage.CreateFile(context.Background(), srcPath)
	if err != nil {
		t.Fatalf("CreateFile for source failed: %v", err)
	}
	content := "copy test content"
	_, err = srcFile.WriteString(content)
	if err != nil {
		t.Fatalf("WriteString to source failed: %v", err)
	}
	srcFile.Close()

	// Copy file
	err = storage.CopyFile(context.Background(), srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify content
	dstFullPath := filepath.Join(tempDir, dstPath)
	dstContent, err := os.ReadFile(dstFullPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(dstContent) != content {
		t.Errorf("Expected content '%s', got '%s'", content, string(dstContent))
	}
}

func TestOsStorage_MoveFile(t *testing.T) {
	storage, tempDir := setupTest(t)
	srcPath := "move_source.txt"
	dstPath := "move_destination.txt"

	// Create source file with content
	srcFile, err := storage.CreateFile(context.Background(), srcPath)
	if err != nil {
		t.Fatalf("CreateFile for source failed: %v", err)
	}
	content := "move test content"
	_, err = srcFile.WriteString(content)
	if err != nil {
		t.Fatalf("WriteString to source failed: %v", err)
	}
	srcFile.Close()

	// Move file
	err = storage.MoveFile(context.Background(), srcPath, dstPath)
	if err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	// Verify content
	dstFullPath := filepath.Join(tempDir, dstPath)
	dstContent, err := os.ReadFile(dstFullPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(dstContent) != content {
		t.Errorf("Expected content '%s', got '%s'", content, string(dstContent))
	}
}

func TestOsStorage_DeleteFile(t *testing.T) {
	storage, _ := setupTest(t)
	filePath := "delete_test.txt"

	// Create file
	file, err := storage.CreateFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	_, err = file.WriteString("delete test content")
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	file.Close()

	// Delete file
	err = storage.DeleteFile(context.Background(), filePath)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}
}
