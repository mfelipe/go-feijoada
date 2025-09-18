package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"

	"github.com/mfelipe/go-feijoada/kafka-consumer/config"
	"github.com/mfelipe/go-feijoada/kafka-consumer/internal"
	"github.com/mfelipe/go-feijoada/utils/testcontainers"
)

var tc *testContext

type testContext struct {
	kafkaContainer *kafka.KafkaContainer
	redisContainer *redis.RedisContainer
	consumer       *internal.Consumer
	producer       *kgo.Client
	adminClient    *kgo.Client
}

func (tc *testContext) shutdown() {
	if tc.consumer != nil {
		tc.consumer.Close()
	}
	if tc.producer != nil {
		tc.producer.Close()
	}
	if tc.adminClient != nil {
		tc.adminClient.Close()
	}
	if tc.kafkaContainer != nil {
		if err := tc.kafkaContainer.Terminate(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to terminate kafka container")
		}
	}
	if tc.redisContainer != nil {
		if err := tc.redisContainer.Terminate(context.Background()); err != nil {
			log.Error().Err(err).Msg("failed to terminate redis container")
		}
	}
}

func setupTestContext() (*testContext, error) {
	ctx := context.Background()

	// Start Kafka container
	kafkaContainer := testcontainers.StartKafka(ctx)
	brokers, err := kafkaContainer.Brokers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get kafka brokers: %w", err)
	}

	// Start Redis container for stream buffer
	redisContainer := testcontainers.StartRedis(ctx)
	redisAddress, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get redis endpoint: %w", err)
	}

	// Set environment variables for kafka-consumer config
	_ = os.Setenv("KC_KAFKA_BROKERS", strings.Join(brokers, ","))
	_ = os.Setenv("KC_KAFKA_TOPICS", "test-topic")
	_ = os.Setenv("KC_KAFKA_GROUP", "test-consumer-group")
	_ = os.Setenv("KC_REPOSITORY_REDIS_ADDRESS", redisAddress)
	_ = os.Setenv("KC_REPOSITORY_REDIS_CLIENTNAME", "kafka-consumer-test")
	_ = os.Setenv("KC_MAXPOLLRECORDS", "10")

	// Create admin client for topic management
	adminClient, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka admin client: %w", err)
	}

	// Create producer for sending test messages
	producer, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka producer: %w", err)
	}

	tc := &testContext{
		kafkaContainer: kafkaContainer,
		redisContainer: redisContainer,
		producer:       producer,
		adminClient:    adminClient,
	}

	return tc, nil
}

func (tc *testContext) createTopic(topicName string) error {
	ctx := context.Background()
	
	// Create topic with 1 partition
	req := kmsg.NewCreateTopicsRequest()
	req.Topics = []kmsg.CreateTopicsRequestTopic{
		{
			Topic:             topicName,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}
	
	_, err := tc.adminClient.Request(ctx, &req)
	if err != nil {
		return fmt.Errorf("failed to create topic %s: %w", topicName, err)
	}
	
	// Wait for topic creation to propagate
	time.Sleep(2 * time.Second)
	log.Info().Str("topic", topicName).Msg("created test topic")
	return nil
}

func TestMain(m *testing.M) {
	var err error
	defer func() {
		if tc != nil {
			tc.shutdown()
		}
	}()

	if tc, err = setupTestContext(); err != nil {
		log.Fatal().Err(err).Msg("failed to setup test context")
	}

	code := m.Run()
	os.Exit(code)
}

func TestKafkaConsumerConfiguration(t *testing.T) {
	t.Run("consumer loads configuration correctly", func(t *testing.T) {
		cfg := config.Load()

		// Verify configuration was loaded from environment variables
		if cfg.Kafka.Brokers == "" {
			t.Error("kafka brokers not configured")
		}
		if cfg.Kafka.Topics == "" {
			t.Error("kafka topics not configured")
		}
		if cfg.Kafka.Group == "" {
			t.Error("kafka group not configured")
		}
		if cfg.MaxPollRecords <= 0 {
			t.Error("max poll records not configured properly")
		}

		log.Info().
			Str("brokers", cfg.Kafka.Brokers).
			Str("topics", cfg.Kafka.Topics).
			Str("group", cfg.Kafka.Group).
			Int("maxPollRecords", cfg.MaxPollRecords).
			Msg("consumer configuration verified")
	})
}

func TestKafkaConsumerIntegration(t *testing.T) {
	// Create the test topic first
	if err := tc.createTopic("test-topic"); err != nil {
		t.Fatalf("failed to create test topic: %v", err)
	}

	// Load config and create consumer
	cfg := config.Load()
	consumer := internal.NewConsumer(*cfg)
	defer consumer.Close()

	ctx := context.Background()
	
	// Test message data
	testMessages := []struct {
		key       string
		value     string
		schemaURI string
	}{
		{
			key:       "user-1",
			value:     `{"name": "John Doe", "age": 30}`,
			schemaURI: "user-schema-v1",
		},
		{
			key:       "user-2", 
			value:     `{"name": "Jane Smith", "age": 25}`,
			schemaURI: "user-schema-v1",
		},
	}

	t.Run("consumer processes messages successfully", func(t *testing.T) {
		// Start consumer in background
		consumerDone := make(chan struct{})
		go func() {
			defer close(consumerDone)
			// Run consumer for a limited time
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			
			// This would normally run indefinitely, but we'll cancel it
			go consumer.Poll()
			<-ctx.Done()
			consumer.Close()
		}()

		// Give consumer time to start
		time.Sleep(3 * time.Second)

		// Produce test messages
		for _, msg := range testMessages {
			record := &kgo.Record{
				Topic: "test-topic",
				Key:   []byte(msg.key),
				Value: []byte(msg.value),
				Headers: []kgo.RecordHeader{
					{Key: "schemaURI", Value: []byte(msg.schemaURI)},
				},
			}

			if err := tc.producer.ProduceSync(ctx, record).FirstErr(); err != nil {
				t.Fatalf("failed to produce message: %v", err)
			}
			log.Info().Str("key", msg.key).Msg("produced test message")
		}

		// Wait a bit for messages to be consumed
		time.Sleep(5 * time.Second)

		// Verify messages were processed (this is a basic integration test)
		// In a real scenario, you'd verify the messages were written to the stream buffer
		log.Info().Msg("integration test completed - messages were produced and consumer was running")
	})
}

func TestKafkaConsumerWithInvalidMessages(t *testing.T) {
	// Ensure topic exists
	if err := tc.createTopic("test-topic"); err != nil {
		t.Fatalf("failed to create test topic: %v", err)
	}

	// Load config and create consumer
	cfg := config.Load()
	consumer := internal.NewConsumer(*cfg)
	defer consumer.Close()

	ctx := context.Background()

	// Test invalid message data
	invalidMessages := []struct {
		key       string
		value     string
		schemaURI string
		reason    string
	}{
		{
			key:       "invalid-json",
			value:     `{"name": "John", "age":}`, // Invalid JSON
			schemaURI: "user-schema-v1",
			reason:    "invalid JSON format",
		},
		{
			key:       "missing-schema",
			value:     `{"name": "Jane"}`,
			schemaURI: "", // No schema URI
			reason:    "missing schema URI",
		},
	}

	t.Run("consumer handles invalid messages gracefully", func(t *testing.T) {
		// Start consumer in background
		consumerDone := make(chan struct{})
		go func() {
			defer close(consumerDone)
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			
			go consumer.Poll()
			<-ctx.Done()
			consumer.Close()
		}()

		// Give consumer time to start
		time.Sleep(2 * time.Second)

		// Produce invalid messages
		for _, msg := range invalidMessages {
			record := &kgo.Record{
				Topic: "test-topic",
				Key:   []byte(msg.key),
				Value: []byte(msg.value),
			}

			// Only add schema header if provided
			if msg.schemaURI != "" {
				record.Headers = []kgo.RecordHeader{
					{Key: "schemaURI", Value: []byte(msg.schemaURI)},
				}
			}

			if err := tc.producer.ProduceSync(ctx, record).FirstErr(); err != nil {
				t.Fatalf("failed to produce invalid message: %v", err)
			}
			log.Info().Str("key", msg.key).Str("reason", msg.reason).Msg("produced invalid test message")
		}

		// Wait for messages to be processed
		time.Sleep(3 * time.Second)

		log.Info().Msg("invalid message test completed - consumer should have handled errors gracefully")
	})
}

func TestKafkaProducerHelper(t *testing.T) {
	// Ensure topic exists
	if err := tc.createTopic("test-topic"); err != nil {
		t.Fatalf("failed to create test topic: %v", err)
	}

	ctx := context.Background()

	t.Run("helper function to produce test messages", func(t *testing.T) {
		// This is a helper test to verify our producer setup works
		testMessage := map[string]interface{}{
			"id":        "test-123",
			"name":      "Test Message",
			"timestamp": time.Now().Unix(),
		}

		messageBytes, err := json.Marshal(testMessage)
		if err != nil {
			t.Fatalf("failed to marshal test message: %v", err)
		}

		record := &kgo.Record{
			Topic: "test-topic",
			Key:   []byte("helper-test"),
			Value: messageBytes,
			Headers: []kgo.RecordHeader{
				{Key: "schemaURI", Value: []byte("helper-test-schema")},
				{Key: "contentType", Value: []byte("application/json")},
			},
		}

		if err := tc.producer.ProduceSync(ctx, record).FirstErr(); err != nil {
			t.Fatalf("failed to produce helper test message: %v", err)
		}

		log.Info().Str("key", "helper-test").Msg("helper test message produced successfully")
	})
}

func TestKafkaContainerSetup(t *testing.T) {
	t.Run("kafka and redis containers are running", func(t *testing.T) {
		ctx := context.Background()
		
		// Test Kafka connection
		brokers, err := tc.kafkaContainer.Brokers(ctx)
		if err != nil {
			t.Fatalf("failed to get kafka brokers: %v", err)
		}
		
		if len(brokers) == 0 {
			t.Fatal("no kafka brokers found")
		}
		
		log.Info().Strs("brokers", brokers).Msg("kafka container is running")
		
		// Test Redis connection
		redisEndpoint, err := tc.redisContainer.Endpoint(ctx, "")
		if err != nil {
			t.Fatalf("failed to get redis endpoint: %v", err)
		}
		
		if redisEndpoint == "" {
			t.Fatal("redis endpoint is empty")
		}
		
		log.Info().Str("endpoint", redisEndpoint).Msg("redis container is running")
	})
}