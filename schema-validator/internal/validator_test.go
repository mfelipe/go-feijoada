package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mfelipe/go-feijoada/schema-validator/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidator_Validate(t *testing.T) {
	cfg := config.Config{
		DefaultBaseURI: "http://localhost:8080",
	}

	tests := []struct {
		name        string
		schema      string
		data        interface{}
		expectValid bool
		expectError bool
	}{
		{
			name: "Valid data against schema",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				},
				"required": ["name"]
			}`,
			data: map[string]interface{}{
				"name": "John",
				"age":  30,
			},
			expectValid: true,
			expectError: false,
		},
		{
			name: "Invalid data - missing required field",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				},
				"required": ["name"]
			}`,
			data: map[string]interface{}{
				"age": 30,
			},
			expectValid: false,
			expectError: false,
		},
		{
			name: "Invalid data - wrong type",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "number"}
				}
			}`,
			data: map[string]interface{}{
				"name": "John",
				"age":  "thirty",
			},
			expectValid: false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := New(cfg)

			// Add schema to validator
			schemaBytes := json.RawMessage(tt.schema)
			err := validator.AddSchema("http://test-schema", schemaBytes)
			require.NoError(t, err)

			// Validate data
			result, err := validator.Validate("http://test-schema", tt.data)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectValid, result.IsValid())
			}
		})
	}
}

func TestValidator_AddSchema(t *testing.T) {
	cfg := config.Config{
		DefaultBaseURI: "http://localhost:8080",
	}
	validator := New(cfg)

	t.Run("Valid schema", func(t *testing.T) {
		schema := json.RawMessage(`{"type": "string"}`)
		err := validator.AddSchema("http://test", schema)
		assert.NoError(t, err)
	})

	t.Run("Invalid schema", func(t *testing.T) {
		schema := json.RawMessage(`{invalid}`)
		err := validator.AddSchema("http://test", schema)
		assert.Error(t, err)
	})
}

func TestValidator_WithHTTPLoader(t *testing.T) {
	// Create a test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		schema := `{"type": "object", "properties": {"id": {"type": "string"}}}`
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(schema))
	}))
	defer server.Close()

	cfg := config.Config{
		DefaultBaseURI: server.URL,
	}
	validator := New(cfg)

	// Test validation with HTTP schema reference
	schemaWithRef := `{
		"type": "object",
		"properties": {
			"user": {"$ref": "` + server.URL + `/user-schema"}
		}
	}`

	schemaBytes := json.RawMessage(schemaWithRef)
	err := validator.AddSchema("http://test-with-ref", schemaBytes)
	assert.NoError(t, err)

	// Test validation
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"id": "123",
		},
	}

	result, err := validator.Validate("http://test-with-ref", data)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestNew(t *testing.T) {
	cfg := config.Config{
		DefaultBaseURI: "http://localhost:8080",
	}

	validator := New(cfg)
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.compiler)
}
