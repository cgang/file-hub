package config

import (
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
