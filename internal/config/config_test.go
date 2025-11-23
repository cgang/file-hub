package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test config file named config.yaml in the temp directory
	testConfigContent := `storage:
  root_dir: "/test/storage/path"
webdav:
  port: "9090"
`

	err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), []byte(testConfigContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write test config file: %v", err)
	}

	t.Run("LoadConfigFromFile", func(t *testing.T) {
		// Temporarily clear CONFIG_PATH to ensure it doesn't interfere
		originalConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", "") // Clear CONFIG_PATH
		defer os.Setenv("CONFIG_PATH", originalConfigPath) // Restore original value

		// Change to the temp directory so that the config file is in the current working directory
		originalWd, err := os.Getwd()
		if err != nil {
			t.Fatalf("Failed to get working directory: %v", err)
		}
		defer os.Chdir(originalWd) // Restore original working directory

		err = os.Chdir(tempDir)
		if err != nil {
			t.Fatalf("Failed to change to temp directory: %v", err)
		}

		// Load the config file from the current directory
		config, err := LoadConfig("nonexistent.yaml") // Use a fallback that doesn't exist
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}

		if config.Storage.RootDir != "/test/storage/path" {
			t.Errorf("Expected root_dir '/test/storage/path', got '%s'", config.Storage.RootDir)
		}

		if config.WebDAV.Port != "9090" {
			t.Errorf("Expected port '9090', got '%s'", config.WebDAV.Port)
		}
	})

	t.Run("LoadConfigWithCONFIG_PATH", func(t *testing.T) {
		// Set CONFIG_PATH to the temp directory
		originalConfigPath := os.Getenv("CONFIG_PATH")
		defer os.Setenv("CONFIG_PATH", originalConfigPath) // Restore original value
		os.Setenv("CONFIG_PATH", tempDir)
		
		// Load config - it should find config.yaml in the temp directory
		config, err := LoadConfig("nonexistent.yaml") // Use a non-existent default file
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		
		if config.Storage.RootDir != "/test/storage/path" {
			t.Errorf("Expected root_dir '/test/storage/path', got '%s'", config.Storage.RootDir)
		}
		
		if config.WebDAV.Port != "9090" {
			t.Errorf("Expected port '9090', got '%s'", config.WebDAV.Port)
		}
	})

	t.Run("GetDefaultConfig", func(t *testing.T) {
		config := GetDefaultConfig()
		
		if config.Storage.RootDir != "./webdav_root" {
			t.Errorf("Expected default root_dir './webdav_root', got '%s'", config.Storage.RootDir)
		}
		
		if config.WebDAV.Port != "8080" {
			t.Errorf("Expected default port '8080', got '%s'", config.WebDAV.Port)
		}
	})

	t.Run("MultipleCONFIG_PATHs", func(t *testing.T) {
		// Create another directory with a different config file
		tempDir2 := t.TempDir()
		testConfigPath2 := filepath.Join(tempDir2, "config.yaml")
		testConfigContent2 := `storage:
  root_dir: "/another/storage/path"
webdav:
  port: "9091"
`
		err := os.WriteFile(testConfigPath2, []byte(testConfigContent2), 0644)
		if err != nil {
			t.Fatalf("Failed to write test config file: %v", err)
		}
		
		// Set CONFIG_PATH to both directories (first one should win)
		originalConfigPath := os.Getenv("CONFIG_PATH")
		defer os.Setenv("CONFIG_PATH", originalConfigPath)
		os.Setenv("CONFIG_PATH", tempDir+":"+tempDir2) // Use colon as separator (Unix)
		
		config, err := LoadConfig("nonexistent.yaml")
		if err != nil {
			t.Fatalf("LoadConfig failed: %v", err)
		}
		
		// Should use the config from the first directory in CONFIG_PATH
		if config.Storage.RootDir != "/test/storage/path" {
			t.Errorf("Expected root_dir '/test/storage/path', got '%s'", config.Storage.RootDir)
		}
		
		if config.WebDAV.Port != "9090" {
			t.Errorf("Expected port '9090', got '%s'", config.WebDAV.Port)
		}
	})
}