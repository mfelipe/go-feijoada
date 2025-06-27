package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestConfig struct {
	Server struct {
		Port     int    `koanf:"port"`
		Host     string `koanf:"host"`
		Timeout  int    `koanf:"timeout"`
		Database struct {
			Host string `koanf:"host"`
			Port int    `koanf:"port"`
		} `koanf:"database"`
	} `koanf:"server"`
}

func TestLoad(t *testing.T) {
	prefix := "CFG"
	// Create a temporary YAML file for testing
	yamlContent := `
cfg:
  server:
    port: 8080
    host: ws.example.com
    timeout: 30
    database:
      host: db.example.com
      port: 5432
`
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configPath, []byte(yamlContent), 0644)
	assert.NoError(t, err)

	t.Run("Load from YAML file", func(t *testing.T) {
		var cfg TestConfig
		Load[TestConfig](prefix, configPath, &cfg)

		// Verify YAML values were loaded correctly
		assert.Equal(t, 8080, cfg.Server.Port)
		assert.Equal(t, "ws.example.com", cfg.Server.Host)
		assert.Equal(t, 30, cfg.Server.Timeout)
		assert.Equal(t, "db.example.com", cfg.Server.Database.Host)
		assert.Equal(t, 5432, cfg.Server.Database.Port)
	})

	t.Run("Override with environment variables", func(t *testing.T) {
		// Set environment variables that should override YAML values
		t.Setenv(prefix+"_SERVER_TIMEOUT", "60")
		t.Setenv(prefix+"_SERVER_DATABASE_PORT", "3254")

		var cfg TestConfig
		Load[TestConfig](prefix, configPath, &cfg)

		// Verify environment variables override YAML values
		assert.Equal(t, "db.example.com", cfg.Server.Database.Host) // Not overridden
		assert.Equal(t, 3254, cfg.Server.Database.Port)
		assert.Equal(t, 8080, cfg.Server.Port)             // Not overridden
		assert.Equal(t, "ws.example.com", cfg.Server.Host) // Not overridden
		assert.Equal(t, 60, cfg.Server.Timeout)
	})

	t.Run("Invalid YAML file path", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected Load to panic with invalid file path")
			}
		}()
		var cfg TestConfig
		Load[TestConfig](prefix, "nonexistent.yaml", &cfg)
	})

	t.Run("Invalid environment variable type", func(t *testing.T) {
		t.Setenv(prefix+"_SERVER_DATABASE_PORT", "invalid")

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected Load to panic with invalid environment variable type")
			}
		}()
		var cfg TestConfig
		Load[TestConfig](prefix, configPath, &cfg)
	})
}
