package repository

import (
	"context"

	"github.com/mfelipe/go-feijoada/schema-repository/config"
	"github.com/mfelipe/go-feijoada/schema-repository/internal/clients"
)

type Repository interface {
	Set(ctx context.Context, key string, value string) error
	Del(ctx context.Context, keys ...string) error
	Get(ctx context.Context, key string) (string, error)
}

// NewRepository creates a new Redis or Valkey implementation of Repository interface
// TODO: Options - WithClient
func NewRepository(cfg config.Repository) Repository {
	var r Repository

	if cfg.Redis != nil {
		c := clients.NewRedisClient(*cfg.Redis)
		r = &redisClient{
			client: c,
		}
	} else if cfg.Valkey != nil {
		c := clients.NewValkeyClient(*cfg.Valkey)
		r = &valkeyClient{
			client: c,
		}
	} else {
		panic("no repository configured")
	}

	return r
}
