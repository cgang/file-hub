package stor

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTest(t *testing.T) (Storage, string) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	storage := NewStorage(tempDir)
	return storage, tempDir
}

func TestOsStorage_CreateFile(t *testing.T) {
	storage, _ := setupTest(t)
	filePath := "test.txt"

	// Create a file
	file, err := storage.CreateFile(filePath)
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

	// Check if file exists
	exists, err := storage.Exists(filePath)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("File should exist after creation")
	}
}

func TestOsStorage_GetFileInfo(t *testing.T) {
	storage, _ := setupTest(t)
	filePath := "test_info.txt"

	// Create a file
	file, err := storage.CreateFile(filePath)
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
	info, err := storage.GetFileInfo(filePath)
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
	err := storage.CreateDir(dirPath)
	if err != nil {
		t.Fatalf("CreateDir failed: %v", err)
	}

	// Check if directory exists
	exists, err := storage.Exists(dirPath)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Error("Directory should exist after creation")
	}
}

func TestOsStorage_CopyFile(t *testing.T) {
	storage, tempDir := setupTest(t)
	srcPath := "source.txt"
	dstPath := "destination.txt"

	// Create source file with content
	srcFile, err := storage.CreateFile(srcPath)
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
	err = storage.CopyFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file exists
	exists, err := storage.Exists(dstPath)
	if err != nil {
		t.Fatalf("Exists for destination failed: %v", err)
	}
	if !exists {
		t.Error("Destination file should exist after copy")
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
	srcFile, err := storage.CreateFile(srcPath)
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
	err = storage.MoveFile(srcPath, dstPath)
	if err != nil {
		t.Fatalf("MoveFile failed: %v", err)
	}

	// Verify source file no longer exists
	exists, err := storage.Exists(srcPath)
	if err != nil {
		t.Fatalf("Exists for source failed: %v", err)
	}
	if exists {
		t.Error("Source file should not exist after move")
	}

	// Verify destination file exists
	exists, err = storage.Exists(dstPath)
	if err != nil {
		t.Fatalf("Exists for destination failed: %v", err)
	}
	if !exists {
		t.Error("Destination file should exist after move")
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
	file, err := storage.CreateFile(filePath)
	if err != nil {
		t.Fatalf("CreateFile failed: %v", err)
	}
	_, err = file.WriteString("delete test content")
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}
	file.Close()

	// Verify file exists
	exists, err := storage.Exists(filePath)
	if err != nil {
		t.Fatalf("Exists before delete failed: %v", err)
	}
	if !exists {
		t.Error("File should exist before delete")
	}

	// Delete file
	err = storage.DeleteFile(filePath)
	if err != nil {
		t.Fatalf("DeleteFile failed: %v", err)
	}

	// Verify file no longer exists
	exists, err = storage.Exists(filePath)
	if err != nil {
		t.Fatalf("Exists after delete failed: %v", err)
	}
	if exists {
		t.Error("File should not exist after delete")
	}
}
