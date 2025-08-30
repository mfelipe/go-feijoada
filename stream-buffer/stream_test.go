package streambuffer

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mfelipe/go-feijoada/stream-buffer/config"
)

func TestNew_Redis(t *testing.T) {
	cfg := config.Config{
		Redis: &config.Server{
			IsCluster:  false,
			Address:    "localhost:6379",
			Username:   "testuser",
			Password:   "testpass",
			ClientName: "test-client",
		},
		Stream: config.Stream{
			Name:      "test-stream",
			Group:     "test-group",
			Consumer:  "test-consumer",
			ReadCount: 10,
			Block:     time.Second,
		},
	}

	stream := New(cfg)
	require.NotNil(t, stream)
	// We can't directly test the type since stream is unexported
	// but we can verify it's not nil and implements the Stream interface
	assert.NotNil(t, stream)
}

func TestNew_Valkey(t *testing.T) {
	cfg := config.Config{
		Valkey: &config.Server{
			IsCluster:  false,
			Address:    "valkey://localhost:6379",
			Username:   "testuser",
			Password:   "testpass",
			ClientName: "test-client",
		},
		Stream: config.Stream{
			Name:      "test-stream",
			Group:     "test-group",
			Consumer:  "test-consumer",
			ReadCount: 10,
			Block:     time.Second,
		},
	}

	// Panics as valkey client tries to connect to a non-existing server
	assert.Panics(t, func() {
		New(cfg)
	})
}

func TestNew_NeitherRedisNorValkey(t *testing.T) {
	cfg := config.Config{
		Stream: config.Stream{
			Name:      "test-stream",
			Group:     "test-group",
			Consumer:  "test-consumer",
			ReadCount: 10,
			Block:     time.Second,
		},
	}

	stream := New(cfg)
	assert.Nil(t, stream)
}
