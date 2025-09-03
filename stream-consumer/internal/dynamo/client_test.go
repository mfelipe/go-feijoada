package dynamo

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
	"github.com/mfelipe/go-feijoada/stream-consumer/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNew(t *testing.T) {
	cfg := config.DynamoDB{
		TableName:    "test-table",
		RetryMax:     3,
		RetryWaitMax: time.Second * 5,
	}

	// This test mainly ensures New doesn't panic
	// In a real scenario, this would require AWS credentials
	assert.NotPanics(t, func() {
		New(cfg)
	})
}

func TestMessageToItem(t *testing.T) {
	client := &Client{tableName: "test-table"}
	timestamp := time.Now()
	message := models.Message{
		Origin:    "kafka",
		SchemaURI: "http://schema-repo/user/1.0.0",
		Data:      json.RawMessage(`{"id": "123", "name": "John Doe"}`),
		Timestamp: timestamp,
	}

	item := client.messageToItem("test-id", message)

	// Verify all expected attributes are present
	assert.Contains(t, item, "id")
	assert.Contains(t, item, "origin")
	assert.Contains(t, item, "schemaURI")
	assert.Contains(t, item, "data")
	assert.Contains(t, item, "timestamp")

	// Verify attribute values
	assert.Equal(t, &types.AttributeValueMemberS{Value: "test-id"}, item["id"])
	assert.Equal(t, &types.AttributeValueMemberS{Value: "kafka"}, item["origin"])
	assert.Equal(t, &types.AttributeValueMemberS{Value: "http://schema-repo/user/1.0.0"}, item["schemaURI"])
	assert.Equal(t, &types.AttributeValueMemberS{Value: `{"id": "123", "name": "John Doe"}`}, item["data"])
	assert.Equal(t, &types.AttributeValueMemberS{Value: timestamp.Format(time.RFC3339)}, item["timestamp"])
}

func setupTestClient(t *testing.T, setupFunc func(*mockClient)) *Client {
	mCli := newMockClient(t)
	setupFunc(mCli)

	c := &Client{
		tableName: "test-table",
		db:        mCli,
	}

	return c
}

func TestClient_BatchWrite(t *testing.T) {
	tests := []struct {
		name                string
		messages            map[string]models.Message
		setupMock           func(*mockClient)
		expectedUnpersisted []string
		expectError         bool
		errorMsg            string
	}{
		{
			name: "successful batch write",
			messages: map[string]models.Message{
				"1": {
					Origin:    "test-origin",
					SchemaURI: "test-schema",
				},
			},
			setupMock: func(m *mockClient) {
				m.EXPECT().BatchWriteItem(mock.Anything, mock.MatchedBy(func(input *dynamodb.BatchWriteItemInput) bool {
					if reqs, ok := input.RequestItems["test-table"]; ok {
						return len(reqs) == 1
					}
					return false
				})).Return(&dynamodb.BatchWriteItemOutput{
					UnprocessedItems: map[string][]types.WriteRequest{},
				}, nil)
			},
			expectedUnpersisted: []string{},
			expectError:         false,
		},
		{
			name: "partial failure in batch write",
			messages: map[string]models.Message{
				"1": {
					Origin:    "test-origin",
					SchemaURI: "test-schema",
				},
				"2": {
					Origin:    "test-origin-2",
					SchemaURI: "test-schema-2",
				},
			},
			setupMock: func(m *mockClient) {
				m.EXPECT().BatchWriteItem(mock.Anything, mock.Anything).Return(&dynamodb.BatchWriteItemOutput{
					UnprocessedItems: map[string][]types.WriteRequest{
						"test-table": {
							{
								PutRequest: &types.PutRequest{
									Item: map[string]types.AttributeValue{
										"id": &types.AttributeValueMemberS{Value: "2"},
									},
								},
							},
						},
					},
				}, nil)
			},
			expectedUnpersisted: []string{"2"},
			expectError:         false,
		},
		{
			name: "complete failure in batch write",
			messages: map[string]models.Message{
				"1": {
					Origin:    "test-origin",
					SchemaURI: "test-schema",
				},
			},
			setupMock: func(m *mockClient) {
				m.EXPECT().BatchWriteItem(mock.Anything, mock.Anything).Return(nil, assert.AnError)
			},
			expectedUnpersisted: []string{"1"},
			expectError:         true,
			errorMsg:            assert.AnError.Error(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := setupTestClient(t, tt.setupMock)

			unpersisted, err := c.BatchWrite(context.Background(), tt.messages)

			assert.ElementsMatch(t, tt.expectedUnpersisted, unpersisted)

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
