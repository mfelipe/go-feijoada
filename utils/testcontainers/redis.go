package testcontainers

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func StartRedis(ctx context.Context) *tcredis.RedisContainer {
	redisContainer, err := tcredis.Run(ctx,
		"redis:7",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		testcontainers.WithLogConsumers(NewStdoutLogConsumer("redis")),
	)
	validateContainerStart(ctx, redisContainer, err)
	return redisContainer
}
