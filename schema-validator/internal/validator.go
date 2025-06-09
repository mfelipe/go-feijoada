package internal

import (
	"encoding/json"
	"errors"

	"github.com/mfelipe/go-feijoada/schema-validator/config"

	"github.com/kaptinlin/jsonschema"
)

// validator is a struct that holds the schemas and provides methods for validation.
type validator struct {
	compiler *jsonschema.Compiler
}

// TODO: Make it singleflight
func (v *validator) Validate(schemaURI string, obj any) (*jsonschema.EvaluationResult, error) {
	schema, err := v.compiler.GetSchema(schemaURI)
	if err != nil {
		return nil, err
	}
	if schema == nil {
		return nil, errors.New("schema not found")
	}
	result := schema.Validate(obj)
	if result == nil {
		return nil, errors.New("validation result is nil")
	}

	return result, nil
}

func (v *validator) AddSchema(uri string, schema json.RawMessage) error {
	_, err := v.compiler.Compile(schema, uri)
	return err
}

func New(config config.Config) *validator {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultBaseURI = config.DefaultBaseURI
	return &validator{compiler: compiler}
}
