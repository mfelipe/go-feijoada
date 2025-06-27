package schemavalidator

import (
	"encoding/json"

	"github.com/kaptinlin/jsonschema"

	"github.com/mfelipe/go-feijoada/schema-validator/config"
	"github.com/mfelipe/go-feijoada/schema-validator/internal"
)

// SchemaValidator defines an interface for validating JSON data against a schema.
type SchemaValidator interface {
	Validate(schemaURI string, obj any) (*jsonschema.EvaluationResult, error)
	AddSchema(uri string, schema json.RawMessage) error
}

func New(config config.Config) SchemaValidator {
	return internal.New(config)
}
