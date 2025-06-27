package streambuffer

import (
	"context"

	"github.com/mfelipe/go-feijoada/stream-buffer/config"
	"github.com/mfelipe/go-feijoada/stream-buffer/internal/redis"
	"github.com/mfelipe/go-feijoada/stream-buffer/internal/valkey"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
)

type Stream interface {
	Add(ctx context.Context, message models.Message) error
	ReadGroup(ctx context.Context) (map[string]models.Message, error)
	Ack(ctx context.Context, ids ...string) error
	Delete(ctx context.Context, ids ...string) error
}

// New creates a new Stream client based on the loaded configuration, For Redis or Valkey
func New(config config.Config) Stream {
	var s Stream
	if config.Redis != nil {
		s = redis.New(*config.Redis, config.Stream)
	}
	if config.Valkey != nil {
		s = valkey.New(*config.Valkey, config.Stream)
	}

	return s
}
