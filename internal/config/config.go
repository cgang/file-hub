package config

import (
	"gopkg.in/yaml.v3"
	"os"
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

// LoadConfig loads configuration from a YAML file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
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
