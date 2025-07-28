package consumer

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"

	"github.com/mfelipe/go-feijoada/stream-buffer"
	"github.com/mfelipe/go-feijoada/stream-consumer/config"
	"github.com/mfelipe/go-feijoada/stream-consumer/internal/dynamo"
)

type Consumer struct {
	stream    streambuffer.Stream
	dynamo    *dynamo.Client
	batchSize int
	interval  time.Duration
	done      chan struct{}
}

func New(cfg *config.Config) (*Consumer, error) {
	// Create DynamoDB client
	dynamoClient := dynamo.New(cfg.DynamoDB)

	// Create stream client
	stream := streambuffer.New(cfg.Repository)

	return &Consumer{
		stream:    stream,
		dynamo:    dynamoClient,
		batchSize: cfg.Consumer.BatchSize,
		interval:  cfg.Consumer.Interval,
		done:      make(chan struct{}),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	for {
		select {
		case <-c.done:
			return nil
		case <-ctx.Done():
			return nil
		case <-time.After(c.interval):
			c.processTick(ctx)
		}
	}
}

func (c *Consumer) processTick(ctx context.Context) {
	//TODO: parallel processing
	if err := c.processBatch(ctx); err != nil {
		zlog.Error().Err(err).Msg("failed to process batch")
	}
}

func (c *Consumer) processBatch(ctx context.Context) error {
	// Read messages from stream
	messages, err := c.stream.ReadGroup(ctx)
	if err != nil {
		return fmt.Errorf("failed to read from stream: %c", err)
	}

	if len(messages) == 0 {
		zlog.Info().Msg("no messages were read from the stream")
		return nil
	}

	// Write messages to DynamoDB
	unpersisted, err := c.dynamo.BatchWrite(ctx, messages)

	// Compile what was persisted and what was not
	var persistedLogEvent = zerolog.Arr()
	var unpersistedLogEvent = zerolog.Arr()
	var persisted = make([]string, 0)
	for id := range messages {
		if slices.Contains(unpersisted, id) {
			unpersistedLogEvent.Str(id)
		} else {
			persistedLogEvent.Str(id)
			persisted = append(persisted, id)
		}
	}

	// Check if no item was persisted, we only return error in this scenario
	if len(persisted) == 0 {
		if err == nil {
			err = fmt.Errorf("no items were persisted but client didn't throw any error")
		}

		zlog.Error().Err(err).Array("unpersistedStreamIds", unpersistedLogEvent).Msg("failed to batch write items to DynamoDB")
		return err
	}

	// Log unpersisted messages, if any
	if len(unpersisted) > 0 {
		zlog.Error().Err(err).Array("unpersistedStreamIds", unpersistedLogEvent).Msg("failed to write one of more items to DynamoDB")
	}

	logEvent := zerolog.Dict().
		Array("persistedStreamIds", persistedLogEvent).
		Array("unpersistedStreamIds", unpersistedLogEvent).
		Int("persistedCount", len(persisted)).
		Int("unpersistedCount", len(unpersisted))
	var ackErr error
	defer func() {
		if ackErr == nil {
			zlog.Info().Dict("persistence", logEvent).Msg("message batch processed successfully")
		} else {
			zlog.Error().Dict("persistence", logEvent).Err(ackErr).Msg("failed to process message batch")
		}
	}()

	// Acknowledge messages in stream
	if ackErr = c.stream.Ack(ctx, persisted...); ackErr != nil {
		zlog.Info().Array("persisted stream ids", persistedLogEvent).Msg("failed to acknowledge messages in the stream")
		return ackErr
	}

	return nil
}

func (c *Consumer) Close() {
	close(c.done)
}
