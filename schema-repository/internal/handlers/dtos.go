package handlers

import (
	"encoding/json"

	"github.com/mfelipe/go-feijoada/schema-repository/internal/models"
)

// SchemaBody defines the request body for creating a new schema.
type SchemaBody struct {
	Schema json.RawMessage `json:"schema" binding:"required,json_schema"`
}

// SchemaRequestURI defines the response body for retrieving or creating a schema.
type SchemaRequestURI struct {
	Name    string        `json:"name" uri:"name" binding:"required"`
	Version models.Semver `json:"version" uri:"version" binding:"required"`
}

// SchemaResponseBody defines the response body for retrieving or creating a schema.
type SchemaResponseBody struct {
	Schema json.RawMessage `json:"schema"`
}

// ErrorResponse defines the structure for error messages.
type ErrorResponse struct {
	Error string `json:"error"`
}
