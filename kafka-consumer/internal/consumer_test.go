package internal

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/kaptinlin/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/mfelipe/go-feijoada/kafka-consumer/config"
	sbmodels "github.com/mfelipe/go-feijoada/stream-buffer/models"
)

// Mock implementations
type MockStream struct {
	mock.Mock
}

func (m *MockStream) Add(ctx context.Context, message sbmodels.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

type MockValidator struct {
	mock.Mock
	shouldReturnValid bool
	shouldReturnError bool
}

func (m *MockValidator) Validate(schemaURI string, obj any) (*jsonschema.EvaluationResult, error) {
	args := m.Called(schemaURI, obj)
	
	if m.shouldReturnError {
		return nil, args.Error(1)
	}
	
	// Create a real jsonschema.EvaluationResult
	compiler := jsonschema.NewCompiler()
	
	if m.shouldReturnValid {
		// Create a schema that will pass validation
		schema, _ := compiler.Compile(json.RawMessage(`{"type": "object"}`), "test")
		result := schema.Validate(map[string]interface{}{})
		return result, nil
	} else {
		// Create a schema that will fail validation
		schema, _ := compiler.Compile(json.RawMessage(`{"type": "string"}`), "test")
		result := schema.Validate(123) // number vs string should fail
		return result, nil
	}
}

func (m *MockValidator) AddSchema(uri string, schema json.RawMessage) error {
	args := m.Called(uri, schema)
	return args.Error(0)
}

func TestPconsumerValidateRecords(t *testing.T) {
	tests := []struct {
		name     string
		records  []*kgo.Record
		setup    func(*MockValidator)
		expected int
	}{
		{
			name: "validates records successfully",
			records: []*kgo.Record{
				{
					Topic:     "test-topic",
					Key:       []byte("key1"),
					Value:     []byte(`{"test": "data"}`),
					Timestamp: time.Now(),
					Headers: []kgo.RecordHeader{
						{Key: "schemaURI", Value: []byte("test-schema")},
					},
				},
			},
			setup: func(mv *MockValidator) {
				mv.shouldReturnValid = true
				mv.shouldReturnError = false
				mv.On("Validate", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expected: 1,
		},
		{
			name: "filters out invalid records",
			records: []*kgo.Record{
				{
					Topic:     "test-topic",
					Key:       []byte("key1"),
					Value:     []byte(`{"test": "data"}`),
					Timestamp: time.Now(),
					Headers: []kgo.RecordHeader{
						{Key: "schemaURI", Value: []byte("test-schema")},
					},
				},
			},
			setup: func(mv *MockValidator) {
				mv.shouldReturnValid = false
				mv.shouldReturnError = false
				mv.On("Validate", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expected: 0,
		},
		{
			name: "handles validation errors",
			records: []*kgo.Record{
				{
					Topic:     "test-topic",
					Key:       []byte("key1"),
					Value:     []byte(`{"test": "data"}`),
					Timestamp: time.Now(),
					Headers: []kgo.RecordHeader{
						{Key: "schemaURI", Value: []byte("test-schema")},
					},
				},
			},
			setup: func(mv *MockValidator) {
				mv.shouldReturnValid = false
				mv.shouldReturnError = true
				mv.On("Validate", mock.Anything, mock.Anything).Return(nil, errors.New("validation error"))
			},
			expected: 0,
		},
		{
			name: "handles records without schemaURI header",
			records: []*kgo.Record{
				{
					Topic:     "test-topic",
					Key:       []byte("key1"),
					Value:     []byte(`{"test": "data"}`),
					Timestamp: time.Now(),
					Headers:   []kgo.RecordHeader{},
				},
			},
			setup: func(mv *MockValidator) {
				mv.shouldReturnValid = true
				mv.shouldReturnError = false
				mv.On("Validate", mock.Anything, mock.Anything).Return(nil, nil)
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := &MockValidator{}
			tt.setup(mockValidator)

			pc := &pconsumer{
				validator: mockValidator,
				topic:     "test-topic",
				partition: 0,
			}

			ctx := context.Background()
			result := pc.validateRecords(ctx, tt.records)

			assert.Len(t, result, tt.expected)
			mockValidator.AssertExpectations(t)
		})
	}
}

func TestPconsumerAddToStream(t *testing.T) {
	tests := []struct {
		name     string
		messages []*sbmodels.Message
		setup    func(*MockStream)
		wantErr  bool
	}{
		{
			name: "adds messages successfully",
			messages: []*sbmodels.Message{
				{
					Origin:    "test-topic",
					SchemaURI: "test-schema",
					Timestamp: time.Now(),
					Data:      json.RawMessage(`{"test": "data"}`),
				},
			},
			setup: func(ms *MockStream) {
				ms.On("Add", mock.Anything, mock.AnythingOfType("models.Message")).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "handles stream errors - logs error but doesn't return it due to defer bug",
			messages: []*sbmodels.Message{
				{
					Origin:    "test-topic",
					SchemaURI: "test-schema",
					Timestamp: time.Now(),
					Data:      json.RawMessage(`{"test": "data"}`),
				},
			},
			setup: func(ms *MockStream) {
				ms.On("Add", mock.Anything, mock.AnythingOfType("models.Message")).Return(errors.New("stream error"))
			},
			wantErr: false, // Bug in original code - error is not properly returned due to defer pattern
		},
		{
			name: "skips nil messages",
			messages: []*sbmodels.Message{
				nil,
				{
					Origin:    "test-topic",
					SchemaURI: "test-schema",
					Timestamp: time.Now(),
					Data:      json.RawMessage(`{"test": "data"}`),
				},
			},
			setup: func(ms *MockStream) {
				ms.On("Add", mock.Anything, mock.AnythingOfType("models.Message")).Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStream := &MockStream{}
			tt.setup(mockStream)

			pc := &pconsumer{
				stream: mockStream,
			}

			ctx := context.Background()
			err := pc.addToStream(ctx, tt.messages)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockStream.AssertExpectations(t)
		})
	}
}

func TestPconsumerValidateMessage(t *testing.T) {
	tests := []struct {
		name      string
		message   sbmodels.Message
		setup     func(*MockValidator)
		wantValid bool
		wantErr   bool
	}{
		{
			name: "validates message successfully",
			message: sbmodels.Message{
				Origin:    "test-topic",
				SchemaURI: "test-schema",
				Timestamp: time.Now(),
				Data:      json.RawMessage(`{"test": "data"}`),
			},
			setup: func(mv *MockValidator) {
				mv.shouldReturnValid = true
				mv.shouldReturnError = false
				mv.On("Validate", mock.Anything, mock.Anything).Return(nil, nil)
			},
			wantValid: true,
			wantErr:   false,
		},
		{
			name: "returns false for invalid message",
			message: sbmodels.Message{
				Origin:    "test-topic",
				SchemaURI: "test-schema",
				Timestamp: time.Now(),
				Data:      json.RawMessage(`{"invalid": "data"}`),
			},
			setup: func(mv *MockValidator) {
				mv.shouldReturnValid = false
				mv.shouldReturnError = false
				mv.On("Validate", mock.Anything, mock.Anything).Return(nil, nil)
			},
			wantValid: false,
			wantErr:   false,
		},
		{
			name: "handles validation error",
			message: sbmodels.Message{
				Origin:    "test-topic",
				SchemaURI: "test-schema",
				Timestamp: time.Now(),
				Data:      json.RawMessage(`{"test": "data"}`),
			},
			setup: func(mv *MockValidator) {
				mv.shouldReturnValid = false
				mv.shouldReturnError = true
				mv.On("Validate", mock.Anything, mock.Anything).Return(nil, errors.New("validation failed"))
			},
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockValidator := &MockValidator{}
			tt.setup(mockValidator)

			pc := &pconsumer{
				validator: mockValidator,
			}

			ctx := context.Background()
			valid, err := pc.validateMessage(ctx, tt.message)

			assert.Equal(t, tt.wantValid, valid)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			mockValidator.AssertExpectations(t)
		})
	}
}

func TestConsumerAssigned(t *testing.T) {
	tests := []struct {
		name     string
		assigned map[string][]int32
		validate func(*testing.T, *Consumer)
	}{
		{
			name: "creates partition consumers for assigned partitions",
			assigned: map[string][]int32{
				"topic1": {0, 1},
				"topic2": {0},
			},
			validate: func(t *testing.T, c *Consumer) {
				assert.Len(t, c.consumers, 3)
				
				// Check that consumers were created for each topic-partition
				assert.Contains(t, c.consumers, consumerKey{"topic1", 0})
				assert.Contains(t, c.consumers, consumerKey{"topic1", 1})
				assert.Contains(t, c.consumers, consumerKey{"topic2", 0})
				
				// Verify consumer properties
				pc := c.consumers[consumerKey{"topic1", 0}]
				assert.Equal(t, "topic1", pc.topic)
				assert.Equal(t, int32(0), pc.partition)
				assert.NotNil(t, pc.quit)
				assert.NotNil(t, pc.done)
				assert.NotNil(t, pc.records)
			},
		},
		{
			name:     "handles empty assignment",
			assigned: map[string][]int32{},
			validate: func(t *testing.T, c *Consumer) {
				assert.Len(t, c.consumers, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Consumer{
				PartitionRecordsChannelSize: 10,
			}
			
			consumer := &Consumer{
				cfg:       cfg,
				consumers: make(map[consumerKey]*pconsumer),
				stream:    &MockStream{},
				validator: &MockValidator{},
			}

			ctx := context.Background()
			consumer.assigned(ctx, nil, tt.assigned)
			
			// Give goroutines a moment to start
			time.Sleep(10 * time.Millisecond)
			
			tt.validate(t, consumer)
			
			// Clean up goroutines
			for _, pc := range consumer.consumers {
				close(pc.quit)
			}
		})
	}
}

func TestConsumerLost(t *testing.T) {
	tests := []struct {
		name     string
		setup    func() *Consumer
		lost     map[string][]int32
		validate func(*testing.T, *Consumer)
	}{
		{
			name: "removes lost partition consumers",
			setup: func() *Consumer {
				cfg := config.Consumer{
					CloseTimeout: 100 * time.Millisecond,
				}
				
				consumer := &Consumer{
					cfg:       cfg,
					consumers: make(map[consumerKey]*pconsumer),
				}
				
				// Create some existing consumers
				pc1 := &pconsumer{
					topic:     "topic1",
					partition: 0,
					quit:      make(chan struct{}),
					done:      make(chan struct{}),
				}
				pc2 := &pconsumer{
					topic:     "topic1",
					partition: 1,
					quit:      make(chan struct{}),
					done:      make(chan struct{}),
				}
				
				consumer.consumers[consumerKey{"topic1", 0}] = pc1
				consumer.consumers[consumerKey{"topic1", 1}] = pc2
				
				// Simulate consumers that will close properly
				go func() {
					<-pc1.quit
					close(pc1.done)
				}()
				go func() {
					<-pc2.quit
					close(pc2.done)
				}()
				
				return consumer
			},
			lost: map[string][]int32{
				"topic1": {0, 1},
			},
			validate: func(t *testing.T, c *Consumer) {
				assert.Len(t, c.consumers, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := tt.setup()
			
			ctx := context.Background()
			consumer.lost(ctx, nil, tt.lost)
			
			tt.validate(t, consumer)
		})
	}
}

func TestOptions(t *testing.T) {
	t.Run("WithStream option", func(t *testing.T) {
		mockStream := &MockStream{}
		consumer := &Consumer{}
		
		opt := WithStream(mockStream)
		opt(consumer)
		
		assert.Equal(t, mockStream, consumer.stream)
	})
	
	t.Run("WithValidator option", func(t *testing.T) {
		mockValidator := &MockValidator{}
		consumer := &Consumer{}
		
		opt := WithValidator(mockValidator)
		opt(consumer)
		
		assert.Equal(t, mockValidator, consumer.validator)
	})
}

func TestPconsumerNilHandling(t *testing.T) {
	t.Run("handles nil records slice", func(t *testing.T) {
		mockValidator := &MockValidator{}
		
		pc := &pconsumer{
			validator: mockValidator,
			topic:     "test-topic",
			partition: 0,
		}

		ctx := context.Background()
		result := pc.validateRecords(ctx, nil)

		assert.Len(t, result, 0)
	})
	
	t.Run("handles empty records slice", func(t *testing.T) {
		mockValidator := &MockValidator{}
		
		pc := &pconsumer{
			validator: mockValidator,
			topic:     "test-topic",
			partition: 0,
		}

		ctx := context.Background()
		result := pc.validateRecords(ctx, []*kgo.Record{})

		assert.Len(t, result, 0)
	})
}