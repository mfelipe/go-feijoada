package dynamo

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
	"github.com/mfelipe/go-feijoada/stream-consumer/config"
	"github.com/stretchr/testify/assert"
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

func TestClient_BatchWrite_Structure(t *testing.T) {
	// Test the structure without actual DynamoDB calls
	client := &Client{
		tableName: "test-table",
	}

	messages := map[string]models.Message{
		"msg1": {
			Origin:    "kafka",
			SchemaURI: "http://schema-repo/user/1.0.0",
			Data:      json.RawMessage(`{"id": "1", "name": "John"}`),
			Timestamp: time.Now(),
		},
	}

	// Test messageToItem function
	for id, msg := range messages {
		item := client.messageToItem(id, msg)

		// Verify all expected attributes are present
		assert.Contains(t, item, "id")
		assert.Contains(t, item, "origin")
		assert.Contains(t, item, "schemaURI")
		assert.Contains(t, item, "data")
		assert.Contains(t, item, "timestamp")

		// Verify attribute values
		assert.Equal(t, &types.AttributeValueMemberS{Value: id}, item["id"])
		assert.Equal(t, &types.AttributeValueMemberS{Value: msg.Origin}, item["origin"])
		assert.Equal(t, &types.AttributeValueMemberS{Value: msg.SchemaURI}, item["schemaURI"])
		assert.Equal(t, &types.AttributeValueMemberS{Value: string(msg.Data)}, item["data"])
		assert.Equal(t, &types.AttributeValueMemberS{Value: msg.Timestamp.Format(time.RFC3339)}, item["timestamp"])
	}
}
