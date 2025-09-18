package testcontainers

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	tckafka "github.com/testcontainers/testcontainers-go/modules/kafka"
)

func StartKafka(ctx context.Context) *tckafka.KafkaContainer {
	kafkaContainer, err := tckafka.Run(ctx,
		"confluentinc/confluent-local:7.5.0",
		tckafka.WithClusterID("test-cluster"),
		testcontainers.WithLogConsumers(NewStdoutLogConsumer("kafka")),
	)
	validateContainerStart(ctx, kafkaContainer, err)
	return kafkaContainer
}
