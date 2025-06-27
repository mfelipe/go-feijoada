package repository

import (
	"context"
	"errors"

	"github.com/redis/go-redis/v9"

	"github.com/mfelipe/go-feijoada/schema-repository/internal/clients"
)

type redisClient struct {
	client clients.Redis
}

func (r *redisClient) Set(ctx context.Context, key string, value string) error {
	return r.client.Set(ctx, key, value, 0).Err()
}

func (r *redisClient) Del(ctx context.Context, keys ...string) error {
	val, err := r.client.Del(ctx, keys...).Result()

	if val == 0 && err == nil {
		return errors.New(ErrorKeyNotFound)
	}

	return err
}

func (r *redisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()

	if val == "" && errors.Is(err, redis.Nil) {
		return val, errors.New(ErrorKeyNotFound)
	}

	return val, err
}
