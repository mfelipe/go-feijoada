package internal

import (
	"encoding/json"
	"errors"

	"github.com/mfelipe/go-feijoada/schema-validator/config"

	"github.com/kaptinlin/jsonschema"
	"github.com/rs/zerolog"
)

// validator is a struct that holds the schemas and provides methods for validation.
type validator struct {
	compiler *jsonschema.Compiler
	logger   *zerolog.Logger
}

func (v *validator) Validate(schemaURI string, obj any) error {
	schema, err := v.compiler.GetSchema(schemaURI)
	if err != nil {
		return err
	}
	// if schema == nil {
	// 	return errors.New("schema not found")
	// }
	result := schema.Validate(obj)
	// if result == nil {
	// 	return errors.New("validation result is nil")
	// }
	if !result.IsValid() {
		resultList := result.ToList()
		// if resultList == nil {
		// 	return errors.New("validation result list is nil")
		// }
		var err error
		for valErr := range resultList.Errors {
			errors.Join(err, errors.New(valErr))
		}
		return err
	}

	return nil
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
