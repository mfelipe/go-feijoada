package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/mfelipe/go-feijoada/kafka-producer/config"
	"github.com/mfelipe/go-feijoada/schemas/models/v1_0_0"
	"github.com/mfelipe/go-feijoada/schemas/models/v2_0_0"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
	zlog "github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func main() {
	cfg := config.Load()

	// Set global log level
	utilslog.InitializeGlobal(cfg.Log)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(cfg.Kafka.Brokers, ",")...),
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("failed to create kafka client")
	}
	defer client.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	zlog.Info().Msg("Starting kafka producer...")

	for {
		time.Sleep(100 * time.Millisecond)
		select {
		case <-ctx.Done():
			zlog.Info().Msg("Shutting down kafka producer...")
			return
		default:
			record, err := newRecord()
			if err != nil {
				zlog.Error().Err(err).Msg("failed to create new record")
				continue
			}

			zlog.Info().Str("topic", record.Topic).Msg("Producing record")
			client.Produce(ctx, record, func(r *kgo.Record, err error) {
				if err != nil {
					zlog.Error().Err(err).Msg("failed to produce record")
				}
			})
		}
	}
}

func newRecord() (*kgo.Record, error) {
	model, name, version := generateModel()

	modelBytes, err := json.Marshal(&model)
	if err != nil {
		zlog.Error().Err(err).Msg("failed to marshal model")
		return nil, err
	}

	return &kgo.Record{
		Value: modelBytes,
		Headers: []kgo.RecordHeader{
			{Key: "schemaURI", Value: []byte(fmt.Sprintf("http://schema-repository:8080/schemas/%s/%s", name, version))},
		},
		//Key:   []byte(name + version),
		Topic: fmt.Sprintf("%s-topic", name),
	}, nil
}

var genIndex uint8

func generateModel() (model any, name string, version string) {
	genIndex++

	switch genIndex {
	case 1:
		model = newFakeModel[v1_0_0.User]()
		name = "user"
		version = "1.0.0"
	case 2:
		model = newFakeModel[v1_0_0.Address]()
		name = "address"
		version = "1.0.0"
	case 3:
		model = newFakeModel[v1_0_0.Order]()
		name = "order"
		version = "1.0.0"
	case 4:
		model = newFakeModel[v1_0_0.Payment]()
		name = "payment"
		version = "1.0.0"
	case 5:
		model = newFakeModel[v1_0_0.Product]()
		name = "product"
		version = "1.0.0"
	case 6:
		model = newFakeModel[v2_0_0.Address]()
		name = "address"
		version = "2.0.0"
	case 7:
		model = newFakeModel[v2_0_0.Order]()
		model = v2_0_0.Order{}
		name = "order"
		version = "2.0.0"
	case 8:
		model = newFakeModel[v2_0_0.Payment]()
		model = v2_0_0.Payment{}
		name = "payment"
		version = "2.0.0"
	default:
		genIndex = 0
		model = newFakeModel[v2_0_0.Product]()
		model = v2_0_0.Product{}
		name = "product"
		version = "2.0.0"
	}
	return
}

func newFakeModel[T any]() any {
	var model T
	if err := gofakeit.Struct(&model); err != nil {
		zlog.Error().Err(err).Msg("failed to create fake model")
		return nil
	}
	return model
}
