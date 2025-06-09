package service

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/mfelipe/go-feijoada/schema-repository/config"
	"github.com/mfelipe/go-feijoada/schema-repository/internal/models"
	"github.com/mfelipe/go-feijoada/schema-repository/internal/repository"
)

// SchemaService provides methods to manage JSON schemas.
type SchemaService struct {
	cfg config.RepoData
	r   repository.Repository
}

// NewSchemaService creates a new instance of SchemaService.
func NewSchemaService(cfg config.RepoData, r repository.Repository) *SchemaService {
	if r == nil {
		panic(errors.New("repository not initialized"))
	}
	return &SchemaService{
		cfg: cfg,
		r:   r,
	}
}

// AddSchema adds a new schema or a new version of an existing schema.
func (s *SchemaService) AddSchema(ctx context.Context, name string, version models.Semver, schemaContent json.RawMessage) error {
	return s.r.Set(ctx, s.schemaKey(name, version), string(schemaContent))
}

// DeleteSchema removes a specific version of a schema
func (s *SchemaService) DeleteSchema(ctx context.Context, name string, version models.Semver) error {
	err := s.r.Del(ctx, s.schemaKey(name, version))

	if err != nil && err.Error() == repository.ErrorKeyNotFound {
		err = errors.New(ErrorSchemaNotFound)
	}

	return err
}

// GetSchema retrieves a specific version of a schema.
func (s *SchemaService) GetSchema(ctx context.Context, name string, version models.Semver) (json.RawMessage, error) {
	schema, err := s.r.Get(ctx, s.schemaKey(name, version))

	if err != nil && err.Error() == repository.ErrorKeyNotFound {
		return nil, errors.New(ErrorSchemaNotFound)
	}

	return safeToRawMessage(schema)
}

func (s *SchemaService) schemaKey(name string, version models.Semver) string {
	return strings.Join([]string{s.cfg.KeyPrefix, name, version.String()}, s.cfg.KeySeparator)
}

// safeToRawMessage avoid panics if the schema for some reason is not a valid JSON.
func safeToRawMessage(schema string) (rm json.RawMessage, e error) {
	defer func() {
		if r := recover(); r != nil {
			e = errors.New(ErrorInvalidJSONContent)
			switch err := r.(type) {
			case error:
				log.Err(err).Msgf("%s: %s", ErrorInvalidJSONContent, r)
			default:
				log.Error().Msgf("%s: %s", ErrorInvalidJSONContent, err)
			}
		}
	}()
	return json.RawMessage(schema), nil
}
