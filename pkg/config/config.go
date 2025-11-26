package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// WebConfig holds the WebDAV server configuration
type WebConfig struct {
	Port int `yaml:"port"`
}

// DatabaseConfig holds the database configuration
// It only supports URI format
type DatabaseConfig struct {
	URI string `yaml:"uri"`
}

// Config represents the main application configuration
type Config struct {
	Web      WebConfig      `yaml:"web"`
	Database DatabaseConfig `yaml:"database"`
}

// getConfigDirs returns a list of directories to search for config files
// It includes directories from CONFIG_PATH environment variable (split by OS-specific separator)
// and the current working directory
func getConfigDirs() ([]string, error) {
	var searchPaths []string

	cp := os.Getenv("CONFIG_PATH")
	if cp == "" {
		return []string{""}, nil // empty string represents current directory
	}

	// Split CONFIG_PATH by the OS-specific list separator (colon on Unix, semicolon on Windows)
	configPaths := filepath.SplitList(cp)
	for _, configDir := range configPaths {
		// Trim whitespace from directory path
		configDir = strings.TrimSpace(configDir)
		if configDir != "" {
			searchPaths = append(searchPaths, configDir)
		}
	}

	return searchPaths, nil
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	// Get the list of directories to search for config files
	searchPaths, err := getConfigDirs()
	if err != nil {
		return nil, err
	}
	// Search for config.yaml in each directory from getConfigDirs
	for _, searchDir := range searchPaths {
		configFile := filepath.Join(searchDir, "config.yaml")
		var config Config
		if err = loadYamlFile(configFile, &config); err == nil {
			return &config, nil
		} else if os.IsNotExist(err) {
			continue // Try next directory
		} else {
			return nil, fmt.Errorf("error reading config file '%s': %w", configFile, err)
		}
	}

	return nil, fmt.Errorf("no config.yaml file found in CONFIG_PATH directories")
}

func loadYamlFile(filename string, out interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	defer file.Close()

	decoder := yaml.NewDecoder(file)
	return decoder.Decode(out)
}

// SaveConfig saves configuration to a YAML file
func (c *Config) SaveConfig(filename string) error {
	data, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// GetDefaultConfig returns a configuration with default values
func GetDefaultConfig() *Config {
	return &Config{
		Web: WebConfig{
			Port: 8080,
		},
		Database: DatabaseConfig{
			URI: "postgresql://filehub:filehub@localhost:5432/filehub",
		},
	}
}
