package internal_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
	"github.com/valkey-io/valkey-go"

	streambuffer "github.com/mfelipe/go-feijoada/stream-buffer"
	"github.com/mfelipe/go-feijoada/stream-buffer/config"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
)

type testCase struct {
	name   string
	config config.Config
}

// StdoutLogConsumer is a LogConsumer that prints the log to stdout
type StdoutLogConsumer struct {
	prefix string
}

// Accept prints the log to stdout
func (lc *StdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Printf("[%s] %s", lc.prefix, string(l.Content))
}

func startRedis(ctx context.Context) *tcredis.RedisContainer {
	redisContainer, err := tcredis.Run(ctx,
		"redis:7",
		tcredis.WithSnapshotting(10, 1),
		tcredis.WithLogLevel(tcredis.LogLevelVerbose),
		tcredis.WithConfigFile(filepath.Join("testdata", "redis7.conf")),
		testcontainers.WithLogConsumers(&StdoutLogConsumer{"redis"}),
	)
	return validateContainerStart(ctx, redisContainer, err)
}

func startValkey(ctx context.Context) *tcvalkey.ValkeyContainer {
	valkeyContainer, err := tcvalkey.Run(ctx,
		"valkey/valkey:7.2.5",
		tcvalkey.WithSnapshotting(10, 1),
		tcvalkey.WithLogLevel(tcvalkey.LogLevelVerbose),
		tcvalkey.WithConfigFile(filepath.Join("testdata", "valkey7.conf")),
		testcontainers.WithLogConsumers(&StdoutLogConsumer{"valkey"}),
	)
	return validateContainerStart(ctx, valkeyContainer, err)
}

func validateContainerStart[T testcontainers.Container](ctx context.Context, container T, runErr error) T {
	var c T
	if runErr != nil {
		log.Printf("failed to start container: %s", runErr)
		return c
	}

	state, err := container.State(ctx)
	if err != nil {
		log.Printf("failed to get container state: %s", err)
		return c
	}

	log.Printf("%T status: %s", container, state.Status)

	return container
}

func getHostPort(ctx context.Context, t *testing.T, container testcontainers.Container) (string, nat.Port) {
	host, err := container.Host(ctx)
	if err != nil {
		require.NoError(t, err)
		return "", ""
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		require.NoError(t, err)
		return "", ""
	}

	return host, port
}

func notNilError(err error) bool {
	return !(err == nil || errors.Is(err, redis.Nil) || errors.Is(err, valkey.Nil))
}

func TestStreamIntegration(t *testing.T) {
	ctx := context.Background()

	redisC := startRedis(ctx)
	require.NotNil(t, redisC)
	defer redisC.Terminate(ctx)
	redisHost, redisPort := getHostPort(ctx, t, redisC)

	valkeyC := startValkey(ctx)
	require.NotNil(t, valkeyC)
	defer valkeyC.Terminate(ctx)
	valkeyHost, valkeyPort := getHostPort(ctx, t, valkeyC)

	streamCfg := config.Stream{
		Name:      "test-stream",
		Group:     "test-group",
		Consumer:  "test-consumer",
		ReadCount: 100,
		Block:     time.Second,
	}

	testCases := []testCase{
		{
			name: "Redis Backend",
			config: config.Config{
				Redis: &config.Server{
					IsCluster:  false,
					Address:    fmt.Sprintf("%s:%d", redisHost, redisPort.Int()),
					ClientName: "test-client",
				},
				Stream: streamCfg,
			},
		},
		{
			name: "Valkey Backend",
			config: config.Config{
				Valkey: &config.Server{
					IsCluster:  false,
					Address:    fmt.Sprintf("valkey://%s:%d", valkeyHost, valkeyPort.Int()),
					ClientName: "test-client",
				},
				Stream: streamCfg,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			stream := streambuffer.New(tc.config)

			// Test Add and ReadGroup
			msg := models.Message{
				SchemaURI: "test-schema",
				Data:      json.RawMessage(`{"test":"data"}`),
			}

			err := stream.Add(ctx, msg)
			if notNilError(err) {
				require.NoError(t, err)
			}

			messages, err := stream.ReadGroup(ctx)
			if notNilError(err) {
				require.NoError(t, err)
			}
			require.Len(t, messages, 1)

			var msgID string
			for id, m := range messages {
				msgID = id
				assert.Equal(t, msg.SchemaURI, m.SchemaURI)
				assert.JSONEq(t, string(msg.Data), string(m.Data))
			}

			// Test Ack
			err = stream.Ack(ctx, msgID)
			if notNilError(err) {
				require.NoError(t, err)
			}

			// Test Delete
			err = stream.Delete(ctx, msgID)
			if notNilError(err) {
				require.NoError(t, err)
			}

			// Verify message is gone
			messages, err = stream.ReadGroup(ctx)
			if notNilError(err) {
				require.NoError(t, err)
			}
			assert.Empty(t, messages)
		})
	}
}
