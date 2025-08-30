package redis

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/mfelipe/go-feijoada/stream-buffer/config"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
)

var (
	defaultStreamConfig = config.Stream{
		Name:      "test-stream",
		Group:     "test-group",
		Consumer:  "test-consumer",
		ReadCount: 10,
		Block:     time.Second,
	}
)

func setupTestStream(t *testing.T, setupFunc func(*mockClient)) *stream {
	mCli := newMockClient(t)
	setupFunc(mCli)

	s := &stream{
		cfg:    defaultStreamConfig,
		client: mCli,
	}

	return s
}

func TestRedisStream_Add(t *testing.T) {
	tests := []struct {
		name        string
		message     models.Message
		setupMock   func(*mockClient)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful add",
			message: models.Message{
				Origin:    "test-origin",
				SchemaURI: "test-schema",
				Timestamp: time.Now(),
				Data:      json.RawMessage(`{"test": "data"}`),
			},
			setupMock: func(m *mockClient) {
				cmd := &redis.StringCmd{}
				cmd.SetVal("1234567890-1")
				m.EXPECT().XAdd(mock.Anything, mock.MatchedBy(func(args *redis.XAddArgs) bool {
					return args.Stream == "test-stream" && args.NoMkStream == true
				})).Return(cmd)
			},
			expectError: false,
		},
		{
			name: "add with error",
			message: models.Message{
				Origin:    "test-origin",
				SchemaURI: "test-schema",
				Timestamp: time.Now(),
				Data:      json.RawMessage(`{"test": "data"}`),
			},
			setupMock: func(m *mockClient) {
				cmd := &redis.StringCmd{}
				cmd.SetErr(errors.New("redis error"))
				m.EXPECT().XAdd(mock.Anything, mock.Anything).Return(cmd)
			},
			expectError: true,
			errorMsg:    "redis error",
		},
		{
			name: "add with nil result",
			message: models.Message{
				Origin:    "test-origin",
				SchemaURI: "test-schema",
				Timestamp: time.Now(),
				Data:      json.RawMessage(`{"test": "data"}`),
			},
			setupMock: func(m *mockClient) {
				m.EXPECT().XAdd(mock.Anything, mock.Anything).Return(nil)
			},
			expectError: true,
			errorMsg:    nilResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupTestStream(t, tt.setupMock)

			err := s.Add(context.Background(), tt.message)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisStream_ReadGroup(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*mockClient)
		expectedMsgs int
		expectError  bool
		errorMsg     string
	}{
		{
			name: "successful read with messages",
			setupMock: func(m *mockClient) {
				cmd := &redis.XStreamSliceCmd{}
				streams := []redis.XStream{
					{
						Stream: "test-stream",
						Messages: []redis.XMessage{
							{
								ID: "1234567890-0",
								Values: map[string]interface{}{
									"origin":    "test-origin",
									"schemaURI": "test-schema",
									"timestamp": time.Now().Format(time.RFC3339),
									"data":      `{"test": "data"}`,
								},
							},
							{
								ID: "1234567890-1",
								Values: map[string]interface{}{
									"origin":    "test-origin-2",
									"schemaURI": "test-schema-2",
									"timestamp": time.Now().Format(time.RFC3339),
									"data":      `{"test": "data2"}`,
								},
							},
						},
					},
				}
				cmd.SetVal(streams)
				m.EXPECT().XReadGroup(mock.Anything, mock.MatchedBy(func(args *redis.XReadGroupArgs) bool {
					return args.Group == "test-group" && args.Consumer == "test-consumer"
				})).Return(cmd)
			},
			expectedMsgs: 2,
			expectError:  false,
		},
		{
			name: "successful read with no messages",
			setupMock: func(m *mockClient) {
				cmd := &redis.XStreamSliceCmd{}
				cmd.SetVal([]redis.XStream{})
				m.EXPECT().XReadGroup(mock.Anything, mock.Anything).Return(cmd)
			},
			expectedMsgs: 0,
			expectError:  false,
		},
		{
			name: "read with redis error",
			setupMock: func(m *mockClient) {
				cmd := &redis.XStreamSliceCmd{}
				cmd.SetErr(errors.New("redis read error"))
				m.EXPECT().XReadGroup(mock.Anything, mock.Anything).Return(cmd)
			},
			expectedMsgs: 0,
			expectError:  true,
			errorMsg:     "redis read error",
		},
		{
			name: "read with nil result",
			setupMock: func(m *mockClient) {
				m.EXPECT().XReadGroup(mock.Anything, mock.Anything).Return(nil)
			},
			expectedMsgs: 0,
			expectError:  true,
			errorMsg:     nilResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupTestStream(t, tt.setupMock)

			messages, err := s.ReadGroup(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.Len(t, messages, tt.expectedMsgs)
			}
		})
	}
}

func TestRedisStream_Ack(t *testing.T) {
	tests := []struct {
		name        string
		ids         []string
		setupMock   func(*mockClient)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful ack single message",
			ids:  []string{"1234567890-0"},
			setupMock: func(m *mockClient) {
				cmd := &redis.IntCmd{}
				cmd.SetVal(1)
				m.EXPECT().XAck(mock.Anything, "test-stream", "test-group", []string{"1234567890-0"}).Return(cmd)
			},
			expectError: false,
		},
		{
			name: "successful ack multiple messages",
			ids:  []string{"1234567890-0", "1234567890-1"},
			setupMock: func(m *mockClient) {
				cmd := &redis.IntCmd{}
				cmd.SetVal(2)
				m.EXPECT().XAck(mock.Anything, "test-stream", "test-group", []string{"1234567890-0", "1234567890-1"}).Return(cmd)
			},
			expectError: false,
		},
		{
			name: "ack with error",
			ids:  []string{"1234567890-0"},
			setupMock: func(m *mockClient) {
				cmd := &redis.IntCmd{}
				cmd.SetErr(errors.New("redis ack error"))
				m.EXPECT().XAck(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(cmd)
			},
			expectError: true,
			errorMsg:    "redis ack error",
		},
		{
			name: "ack with nil result",
			ids:  []string{"1234567890-0"},
			setupMock: func(m *mockClient) {
				m.EXPECT().XAck(mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: true,
			errorMsg:    nilResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupTestStream(t, tt.setupMock)

			err := s.Ack(context.Background(), tt.ids...)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisStream_Delete(t *testing.T) {
	tests := []struct {
		name        string
		ids         []string
		setupMock   func(*mockClient)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful delete single message",
			ids:  []string{"1234567890-0"},
			setupMock: func(m *mockClient) {
				cmd := &redis.IntCmd{}
				cmd.SetVal(1)
				m.EXPECT().XDel(mock.Anything, "test-stream", []string{"1234567890-0"}).Return(cmd)
			},
			expectError: false,
		},
		{
			name: "successful delete multiple messages",
			ids:  []string{"1234567890-0", "1234567890-1"},
			setupMock: func(m *mockClient) {
				cmd := &redis.IntCmd{}
				cmd.SetVal(2)
				m.EXPECT().XDel(mock.Anything, "test-stream", []string{"1234567890-0", "1234567890-1"}).Return(cmd)
			},
			expectError: false,
		},
		{
			name: "delete with error",
			ids:  []string{"1234567890-0"},
			setupMock: func(m *mockClient) {
				cmd := &redis.IntCmd{}
				cmd.SetErr(errors.New("redis delete error"))
				m.EXPECT().XDel(mock.Anything, mock.Anything, mock.Anything).Return(cmd)
			},
			expectError: true,
			errorMsg:    "redis delete error",
		},
		{
			name: "delete with nil result",
			ids:  []string{"1234567890-0"},
			setupMock: func(m *mockClient) {
				m.EXPECT().XDel(mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			expectError: true,
			errorMsg:    nilResult,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupTestStream(t, tt.setupMock)

			err := s.Delete(context.Background(), tt.ids...)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRedisStream_New(t *testing.T) {
	type args struct {
		serverCfg config.Server
		streamCfg config.Stream
		opts      []Option
	}
	tests := []struct {
		name string
		args args
		want *stream
	}{
		{
			name: "Non-cluster configuration",
			args: args{
				serverCfg: config.Server{
					IsCluster:  false,
					Address:    "localhost:6379",
					ClientName: "test-client",
				},
				streamCfg: defaultStreamConfig,
			},
			want: &stream{
				cfg: defaultStreamConfig,
			},
		},
		{
			name: "Cluster configuration",
			args: args{
				serverCfg: config.Server{
					IsCluster:  true,
					Address:    "localhost:6379",
					ClientName: "test-client",
				},
				streamCfg: defaultStreamConfig,
			},
			want: &stream{
				cfg: defaultStreamConfig,
			},
		},
		{
			name: "WithClient option",
			args: args{
				streamCfg: defaultStreamConfig,
				opts: []Option{
					WithClient(newMockClient(t)),
				},
			},
			want: &stream{
				cfg: defaultStreamConfig,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cli := New(tt.args.serverCfg, tt.args.streamCfg, tt.args.opts...)

			assert.Equal(t, tt.want.cfg, cli.cfg)
			// We cannot directly compare the client field as it's an interface and may contain unexported fields.
			if tt.want.client != nil {
				// If a custom client is provided, ensure it's set.
				assert.Equal(t, tt.want.client, cli.client)
			} else {
				// If no custom client is provided, ensure a client is created.
				assert.NotNil(t, cli.client)
			}
		})
	}
}
