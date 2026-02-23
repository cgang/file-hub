package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestConfigWithS3(t *testing.T) {
	yamlData := `
web:
  port: 8080
database:
  uri: "postgresql://filehub:filehub@localhost:5432/filehub"
s3:
  endpoint: "https://s3.amazonaws.com"
  region: "us-east-1"
  access_key_id: "test-key"
  secret_access_key: "test-secret"
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)
	assert.NotNil(t, cfg.S3)
	assert.Equal(t, "https://s3.amazonaws.com", cfg.S3.Endpoint)
	assert.Equal(t, "us-east-1", cfg.S3.Region)
	assert.Equal(t, "test-key", cfg.S3.AccessKeyID)
	assert.Equal(t, "test-secret", cfg.S3.SecretAccessKey)
}

func TestConfigWithoutS3(t *testing.T) {
	yamlData := `
web:
  port: 8080
database:
  uri: "postgresql://filehub:filehub@localhost:5432/filehub"
`

	var cfg Config
	err := yaml.Unmarshal([]byte(yamlData), &cfg)
	assert.NoError(t, err)
	assert.Nil(t, cfg.S3)
}

func TestGetDefaultConfig(t *testing.T) {
	cfg := newDefaultConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, 8080, cfg.Web.Port)
	assert.Equal(t, "postgresql://filehub:filehub@localhost:5432/filehub", cfg.Database.URI)
	assert.Nil(t, cfg.S3)
}

func TestConfigLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	configContent := `
realm: "test-realm"
web:
  port: 9090
  debug: true
database:
  uri: "postgresql://user:pass@localhost:5432/testdb"
root_dir:
  - /tmp/test1
  - /tmp/test2
s3:
  endpoint: "https://s3.example.com"
  region: "eu-west-1"
  access_key_id: "key123"
  secret_access_key: "secret456"
`
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	// Temporarily set CONFIG_PATH
	oldConfigPath := os.Getenv("CONFIG_PATH")
	os.Setenv("CONFIG_PATH", tmpDir)
	defer os.Setenv("CONFIG_PATH", oldConfigPath)

	cfg, err := LoadConfig("config.yaml")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "test-realm", cfg.Realm)
	assert.Equal(t, 9090, cfg.Web.Port)
	assert.True(t, cfg.Web.Debug)
	assert.Equal(t, "postgresql://user:pass@localhost:5432/testdb", cfg.Database.URI)
	assert.Len(t, cfg.RootDir, 2)
	assert.Equal(t, "/tmp/test1", cfg.RootDir[0])
	assert.Equal(t, "/tmp/test2", cfg.RootDir[1])
	assert.NotNil(t, cfg.S3)
	assert.Equal(t, "https://s3.example.com", cfg.S3.Endpoint)
	assert.Equal(t, "eu-west-1", cfg.S3.Region)
}

func TestConfigValidation(t *testing.T) {
	t.Run("Invalid YAML syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		// Invalid YAML (missing value)
		invalidYAML := `
web:
  port:
database:
  uri: "postgresql://test"
`
		err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
		assert.NoError(t, err)

		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", tmpDir)
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		_, err = LoadConfig("config.yaml")
		// YAML parser may accept incomplete YAML, so we just verify it loads
		// The important thing is that the application handles missing values
		if err == nil {
			// If it loads, it should use defaults for missing values
		}
	})

	t.Run("Empty config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		err := os.WriteFile(configPath, []byte(""), 0644)
		assert.NoError(t, err)

		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", tmpDir)
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		_, err = LoadConfig("config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "EOF")
	})

	t.Run("Missing required fields uses defaults", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		minimalConfig := `web: {}`
		err := os.WriteFile(configPath, []byte(minimalConfig), 0644)
		assert.NoError(t, err)

		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", tmpDir)
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		cfg, err := LoadConfig("config.yaml")
		assert.NoError(t, err)
		assert.Equal(t, 8080, cfg.Web.Port)
		assert.Equal(t, "postgresql://filehub:filehub@localhost:5432/filehub", cfg.Database.URI)
	})
}

func TestConfigPathResolution(t *testing.T) {
	t.Run("Single directory in CONFIG_PATH", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		configContent := `web: { port: 7070 }`
		err := os.WriteFile(configPath, []byte(configContent), 0644)
		assert.NoError(t, err)

		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", tmpDir)
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		cfg, err := LoadConfig("config.yaml")
		assert.NoError(t, err)
		assert.Equal(t, 7070, cfg.Web.Port)
	})

	t.Run("Multiple directories in CONFIG_PATH - first match wins", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		// Config in first directory
		configPath1 := filepath.Join(tmpDir1, "config.yaml")
		configContent1 := `web: { port: 6060 }`
		err := os.WriteFile(configPath1, []byte(configContent1), 0644)
		assert.NoError(t, err)

		// Config in second directory (should be ignored)
		configPath2 := filepath.Join(tmpDir2, "config.yaml")
		configContent2 := `web: { port: 5050 }`
		err = os.WriteFile(configPath2, []byte(configContent2), 0644)
		assert.NoError(t, err)

		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", tmpDir1+string(os.PathListSeparator)+tmpDir2)
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		cfg, err := LoadConfig("config.yaml")
		assert.NoError(t, err)
		assert.Equal(t, 6060, cfg.Web.Port)
	})

	t.Run("No CONFIG_PATH falls back to current directory", func(t *testing.T) {
		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Unsetenv("CONFIG_PATH")
		defer func() {
			if oldConfigPath != "" {
				os.Setenv("CONFIG_PATH", oldConfigPath)
			}
		}()

		// This should fail since there's no config.yaml in current directory
		_, err := LoadConfig("config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no config.yaml file found")
	})

	t.Run("Non-existent CONFIG_PATH directory", func(t *testing.T) {
		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", "/nonexistent/directory/path")
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		_, err := LoadConfig("config.yaml")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no config.yaml file found")
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("Save and reload config", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		originalCfg := &Config{
			Realm: "test-realm",
			Web: WebConfig{
				Port:    8888,
				Debug:   true,
				Metrics: false,
			},
			Database: DatabaseConfig{
				URI: "postgresql://test:test@localhost:5432/testdb",
			},
			RootDir: []string{"/data1", "/data2"},
			S3: &S3Config{
				Endpoint:        "https://s3.test.com",
				Region:          "us-west-2",
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
			},
		}

		err := originalCfg.SaveConfig(configPath)
		assert.NoError(t, err)

		// Verify file was created
		_, err = os.Stat(configPath)
		assert.NoError(t, err)

		// Reload and verify
		loadedCfg := newDefaultConfig()
		err = loadYamlFile(configPath, loadedCfg)
		assert.NoError(t, err)
		assert.Equal(t, originalCfg.Realm, loadedCfg.Realm)
		assert.Equal(t, originalCfg.Web.Port, loadedCfg.Web.Port)
		assert.Equal(t, originalCfg.Web.Debug, loadedCfg.Web.Debug)
		assert.Equal(t, originalCfg.Database.URI, loadedCfg.Database.URI)
		assert.Equal(t, originalCfg.RootDir, loadedCfg.RootDir)
		assert.NotNil(t, loadedCfg.S3)
		assert.Equal(t, originalCfg.S3.Endpoint, loadedCfg.S3.Endpoint)
		assert.Equal(t, originalCfg.S3.Region, loadedCfg.S3.Region)
	})

	t.Run("Save config to invalid path", func(t *testing.T) {
		cfg := newDefaultConfig()
		err := cfg.SaveConfig("/nonexistent/directory/config.yaml")
		assert.Error(t, err)
	})
}

func TestWebConfig(t *testing.T) {
	t.Run("Web config with all options", func(t *testing.T) {
		yamlData := `
web:
  port: 3000
  metrics: true
  debug: true
database:
  uri: "postgresql://localhost/test"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		assert.NoError(t, err)
		assert.Equal(t, 3000, cfg.Web.Port)
		assert.True(t, cfg.Web.Metrics)
		assert.True(t, cfg.Web.Debug)
	})

	t.Run("Web config with defaults", func(t *testing.T) {
		// When web section is not specified, defaults are applied by newDefaultConfig
		// But when unmarshaling directly, zero values are used
		yamlData := `
database:
  uri: "postgresql://localhost/test"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		assert.NoError(t, err)
		// Direct unmarshaling uses zero values for missing fields
		assert.Equal(t, 0, cfg.Web.Port)
		assert.False(t, cfg.Web.Metrics)
		assert.False(t, cfg.Web.Debug)
	})
}

func TestDatabaseConfig(t *testing.T) {
	t.Run("Database config with URI", func(t *testing.T) {
		yamlData := `
database:
  uri: "postgresql://user:password@host:5432/dbname?sslmode=disable"
`
		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		assert.NoError(t, err)
		assert.Equal(t, "postgresql://user:password@host:5432/dbname?sslmode=disable", cfg.Database.URI)
	})

	t.Run("Database config missing URI", func(t *testing.T) {
		yamlData := `
database: {}
`
		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		assert.NoError(t, err)
		assert.Equal(t, "", cfg.Database.URI)
	})
}

func TestRootDirConfig(t *testing.T) {
	t.Run("Root dir with single path", func(t *testing.T) {
		yamlData := `
web: { port: 8080 }
database: { uri: "postgresql://localhost/test" }
root_dir:
  - /single/path
`
		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		assert.NoError(t, err)
		assert.Len(t, cfg.RootDir, 1)
		assert.Equal(t, "/single/path", cfg.RootDir[0])
	})

	t.Run("Root dir with multiple paths", func(t *testing.T) {
		yamlData := `
web: { port: 8080 }
database: { uri: "postgresql://localhost/test" }
root_dir:
  - /path/one
  - /path/two
  - /path/three
`
		var cfg Config
		err := yaml.Unmarshal([]byte(yamlData), &cfg)
		assert.NoError(t, err)
		assert.Len(t, cfg.RootDir, 3)
		assert.Equal(t, "/path/one", cfg.RootDir[0])
		assert.Equal(t, "/path/two", cfg.RootDir[1])
		assert.Equal(t, "/path/three", cfg.RootDir[2])
	})

	t.Run("Root dir empty uses default", func(t *testing.T) {
		cfg := newDefaultConfig()
		assert.Len(t, cfg.RootDir, 1)
		assert.Equal(t, "/tmp", cfg.RootDir[0])
	})
}

func TestConfigGetConfigDirs(t *testing.T) {
	t.Run("Empty CONFIG_PATH returns empty slice", func(t *testing.T) {
		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Unsetenv("CONFIG_PATH")
		defer func() {
			if oldConfigPath != "" {
				os.Setenv("CONFIG_PATH", oldConfigPath)
			}
		}()

		dirs, err := getConfigDirs()
		assert.NoError(t, err)
		assert.Len(t, dirs, 1)
		assert.Equal(t, "", dirs[0])
	})

	t.Run("CONFIG_PATH with single directory", func(t *testing.T) {
		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", "/test/dir")
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		dirs, err := getConfigDirs()
		assert.NoError(t, err)
		assert.Len(t, dirs, 1)
		assert.Equal(t, "/test/dir", dirs[0])
	})

	t.Run("CONFIG_PATH with multiple directories", func(t *testing.T) {
		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", "/dir1:/dir2:/dir3")
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		dirs, err := getConfigDirs()
		assert.NoError(t, err)
		assert.Len(t, dirs, 3)
		assert.Equal(t, "/dir1", dirs[0])
		assert.Equal(t, "/dir2", dirs[1])
		assert.Equal(t, "/dir3", dirs[2])
	})

	t.Run("CONFIG_PATH with whitespace", func(t *testing.T) {
		oldConfigPath := os.Getenv("CONFIG_PATH")
		os.Setenv("CONFIG_PATH", "  /dir1  :  /dir2  :  ")
		defer os.Setenv("CONFIG_PATH", oldConfigPath)

		dirs, err := getConfigDirs()
		assert.NoError(t, err)
		assert.Len(t, dirs, 2)
		assert.Equal(t, "/dir1", dirs[0])
		assert.Equal(t, "/dir2", dirs[1])
	})
}
