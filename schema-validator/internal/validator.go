package internal

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/kaptinlin/jsonschema"
	"golang.org/x/sync/singleflight"

	"github.com/mfelipe/go-feijoada/schema-validator/config"
	utilshttp "github.com/mfelipe/go-feijoada/utils/http"
)

// validator is a struct that holds the schemas and provides methods for validation.
type validator struct {
	compiler *jsonschema.Compiler
}

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

//goland:noinspection GoExportedFuncWithUnexportedType
func New(config config.Config) *validator {
	compiler := jsonschema.NewCompiler()
	compiler.DefaultBaseURI = config.DefaultBaseURI

	return &validator{
		compiler: compiler,
	}
}

// overrideHTTPLoader overwrites the default Compiler http client to get schemas
// The retriable client could be configured accordingly to each scenario, here is on the defaults
// Although this scenario would be extremely rare (besides initial load where the compiler may be empty), single flight
// calls for new schemas may prevent unnecessary high load for the schema repository service.
func (v *validator) overrideHTTPLoader() {
	var singleFlightGroup = singleflight.Group{}
	var httpClient = *retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = utilshttp.CustomPooledTransport()
	var stdHTTPClient = httpClient.StandardClient()

	sfHTTPLoader := func(url string) (io.ReadCloser, error) {
		did, err, _ := singleFlightGroup.Do(url, func() (interface{}, error) {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				return nil, err
			}

			resp, err := stdHTTPClient.Do(req)
			if err != nil {
				return nil, jsonschema.ErrFailedToFetch
			}

			if resp.StatusCode != http.StatusOK {
				err = resp.Body.Close()
				if err != nil {
					return nil, err
				}
				return nil, jsonschema.ErrInvalidHTTPStatusCode
			}

			return resp.Body, nil
		})

		if err != nil {
			return nil, err
		}

		return did.(io.ReadCloser), nil
	}

	v.compiler.RegisterLoader("http", sfHTTPLoader)
	v.compiler.RegisterLoader("https", sfHTTPLoader)
}
