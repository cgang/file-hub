package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// StorageConfig holds the storage configuration
type StorageConfig struct {
	RootDir string `yaml:"root_dir"`
}

// WebDAVConfig holds the WebDAV server configuration
type WebDAVConfig struct {
	Port string `yaml:"port"`
}

// Config represents the main application configuration
type Config struct {
	Storage StorageConfig `yaml:"storage"`
	WebDAV  WebDAVConfig  `yaml:"webdav"`
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
		if _, err := os.Stat(configFile); err == nil {
			// Found a config file, try to load it
			data, readErr := os.ReadFile(configFile)
			if readErr != nil {
				continue // Try next directory
			}

			var config Config
			if unmarshalErr := yaml.Unmarshal(data, &config); unmarshalErr != nil {
				continue // Try next directory
			}

			return &config, nil
		}
	}

	return nil, fmt.Errorf("no config.yaml file found in CONFIG_PATH directories or current working directory")
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
		Storage: StorageConfig{
			RootDir: "./webdav_root",
		},
		WebDAV: WebDAVConfig{
			Port: "8080",
		},
	}
}
