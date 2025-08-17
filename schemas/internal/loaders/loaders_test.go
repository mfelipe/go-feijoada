package loaders

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCachedFileLoader_Load(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "schema-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create test schema files
	userSchema := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"name": {"type": "string"},
			"email": {"type": "string", "format": "email"}
		},
		"required": ["id", "name"]
	}`

	productSchema := `{
		"type": "object",
		"properties": {
			"id": {"type": "string"},
			"title": {"type": "string"},
			"price": {"type": "number", "minimum": 0}
		},
		"required": ["id", "title", "price"]
	}`

	// Write test files
	userFile := filepath.Join(tempDir, "user.json")
	productFile := filepath.Join(tempDir, "product.json")

	err = os.WriteFile(userFile, []byte(userSchema), 0644)
	require.NoError(t, err)

	err = os.WriteFile(productFile, []byte(productSchema), 0644)
	require.NoError(t, err)

	loader := NewCachedFileLoader()

	t.Run("Load existing file", func(t *testing.T) {
		schema, err := loader.Load(userFile, "")
		assert.NoError(t, err)
		assert.NotNil(t, schema)
	})

	t.Run("Load cached file", func(t *testing.T) {
		// Load again to test caching
		schema, err := loader.Load(userFile, "")
		assert.NoError(t, err)
		assert.NotNil(t, schema)
	})

	t.Run("Load different file", func(t *testing.T) {
		schema, err := loader.Load(productFile, "")
		assert.NoError(t, err)
		assert.NotNil(t, schema)
	})

	t.Run("Load non-existent file", func(t *testing.T) {
		_, err := loader.Load(filepath.Join(tempDir, "nonexistent.json"), "")
		assert.Error(t, err)
	})

	t.Run("Load empty path", func(t *testing.T) {
		_, err := loader.Load("", "")
		assert.Error(t, err)
	})
}

func TestNewCachedFileLoader(t *testing.T) {
	loader := NewCachedFileLoader()
	assert.NotNil(t, loader)
	assert.NotNil(t, loader.loader)
	assert.NotNil(t, loader.cache)
}
