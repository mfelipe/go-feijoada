package clients

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/mfelipe/go-feijoada/schema-repository/config"
)

type Redis interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

func NewRedisClient(cfg config.RepoServer) Redis {
	if cfg.IsCluster {
		c := redis.NewClusterClient(&redis.ClusterOptions{

			Addrs:      []string{cfg.Address},
			ClientName: cfg.ClientName,
			NewClient: func(opt *redis.Options) *redis.Client {
				opt.CredentialsProvider = func() (username string, password string) {
					return cfg.Username, cfg.Password
				}
				return redis.NewClient(opt)
			},
		})

		return c
	} else {
		c := redis.NewClient(&redis.Options{
			Addr: cfg.Address,
			CredentialsProvider: func() (username string, password string) {
				return cfg.Username, cfg.Password
			},
		})

		return c
	}
}
