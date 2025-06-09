package internal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	zlog "github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
	"golang.org/x/sync/errgroup"

	"github.com/mfelipe/go-feijoada/kafka-consumer/config"
	schemavalidator "github.com/mfelipe/go-feijoada/schema-validator"
	streambuffer "github.com/mfelipe/go-feijoada/stream-buffer"
	sbmodels "github.com/mfelipe/go-feijoada/stream-buffer/models"
)

// This implementation is based on examples from the frans-go module, more specifically the one for consuming with a
// go routine per partition and manual batch commiting: https://github.com/twmb/franz-go/blob/master/examples/goroutine_per_partition_consuming/

type pconsumer struct {
	stream    streambuffer.Stream
	validator schemavalidator.SchemaValidator
	kcli      *kgo.Client
	topic     string
	partition int32

	quit    chan struct{}
	done    chan struct{}
	records chan []*kgo.Record
}

type consumerKey struct {
	topic     string
	partition int32
}

type Consumer struct {
	cfg       config.Consumer
	stream    streambuffer.Stream
	validator schemavalidator.SchemaValidator
	kcli      *kgo.Client
	consumers map[consumerKey]*pconsumer
}

func NewConsumer(cfg config.Consumer) *Consumer {
	c := &Consumer{
		cfg:       cfg,
		stream:    streambuffer.New(cfg.Stream),
		validator: schemavalidator.New(cfg.SchemaValidator),
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(strings.Split(cfg.Kafka.Brokers, ",")...),
		kgo.ConsumeTopics(strings.Split(cfg.Kafka.Topics, ",")...),
		kgo.ConsumerGroup(cfg.Kafka.Group),
		kgo.OnPartitionsAssigned(c.assigned),
		kgo.OnPartitionsRevoked(c.lost),
		kgo.OnPartitionsLost(c.lost),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll())
	if err != nil {
		log.Fatal(err)
	}

	// check connectivity to cluster
	if err = c.kcli.Ping(context.Background()); err != nil {
		log.Fatal(err)
	}

	c.kcli = client

	return c
}

func (pc *pconsumer) consume(_ context.Context) {
	defer close(pc.done)
	fmt.Printf("Starting consume for topic %s and partition %d\n", pc.topic, pc.partition)
	defer fmt.Printf("Closing consume for topic %s and partition %d\n", pc.topic, pc.partition)
	for {
		select {
		case <-pc.quit:
			return
		case recs := <-pc.records:
			ctx := context.Background()
			err := pc.addToStream(ctx, recs)
			if err != nil {
				return
			}
			fmt.Printf("Some sort of work done, about to commit on topic %s and partition %d\n", pc.topic, pc.partition)
			err = pc.kcli.CommitRecords(ctx, recs...)
			if err != nil {
				fmt.Printf("Error when committing offsets to kafka err: %v topic: %s partition: %d offset %d\n", err, pc.topic, pc.partition, recs[len(recs)-1].Offset+1)
			}
		}
	}
}

func (pc *pconsumer) addToStream(ctx context.Context, records []*kgo.Record) error {
	//TODO: Use a transaction (stream client side) for all records or nothing?
	eg := &errgroup.Group{}
	var err error
	defer func(eg *errgroup.Group) {
		err = eg.Wait()
	}(eg)

	for _, r := range records {
		eg.Go(func() error {
			var schemaURI string
			for _, h := range r.Headers {
				if h.Key == "schemaURI" {
					schemaURI = string(h.Value)
					break
				}
			}

			msg := sbmodels.Message{
				SchemaURI: schemaURI,
				Data:      r.Value,
			}

			zlog.Info().EmbedObject(&msg).Msgf("pooled record with id %s from topic %s", r.Key, r.Topic)

			//TODO: move to another func
			if schemaURI != "" {
				vResult, vErr := pc.validator.Validate(schemaURI, r.Value)
				if vErr != nil {
					return vErr
				}

				if !vResult.IsValid() {
					resultList := vResult.ToList()
					var rErr error
					for e := range resultList.Errors {
						errors.Join(rErr, errors.New(e))
					}

					fmt.Printf("record data is not a valid %s schema: %v", schemaURI, rErr)
					return nil
				}
			}

			// If it's not invalid, we add to the stream
			err := pc.stream.Add(ctx, msg)
			if err != nil {
				fmt.Printf("failed to add record to stream: %v", err)
			}
			return err
		})
	}
	return err
}

func (c *Consumer) Close() {
	c.kcli.Close()
}

func (c *Consumer) Poll() {
	defer c.kcli.Close()
	for {
		// PollRecords is strongly recommended when using
		// BlockRebalanceOnPoll. You can tune how many records to
		// process at once (upper bound -- could all be on one
		// partition), ensuring that your processor loops complete fast
		// enough to not block a rebalance too long.
		fetches := c.kcli.PollRecords(context.Background(), c.cfg.MaxPoolRecords)
		if fetches.IsClientClosed() {
			return
		}
		fetches.EachError(func(_ string, _ int32, err error) {
			// Note: you can delete this block, which will result
			// in these errors being sent to the partition
			// consumers, and then you can handle the errors there.
			panic(err)
		})
		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			k := consumerKey{p.Topic, p.Partition}

			// Since we are using BlockRebalanceOnPoll, we can be
			// sure this partition consumer exists:
			//
			// * onAssigned is guaranteed to be called before we
			// fetch offsets for newly added partitions
			//
			// * onRevoked waits for partition consumers to quit
			// and be deleted before re-allowing polling.
			c.consumers[k].records <- p.Records
		})
		c.kcli.AllowRebalance()
	}
}

func (c *Consumer) assigned(ctx context.Context, cl *kgo.Client, assigned map[string][]int32) {
	for topic, partitions := range assigned {
		for _, partition := range partitions {
			pc := &pconsumer{
				kcli:      cl,
				stream:    c.stream,
				validator: c.validator,
				topic:     topic,
				partition: partition,

				quit:    make(chan struct{}),
				done:    make(chan struct{}),
				records: make(chan []*kgo.Record, c.cfg.PartitionRecordsChannelSize),
			}
			c.consumers[consumerKey{topic, partition}] = pc
			go pc.consume(ctx)
		}
	}
}

// In this example, each partition consumer commits itself. Those commits will
// fail if partitions are lost, but will succeed if partitions are revoked. We
// only need one revoked or lost function (and we name it "lost").
func (c *Consumer) lost(ctx context.Context, _ *kgo.Client, lost map[string][]int32) {
	eg, _ := errgroup.WithContext(ctx)
	defer func(eg *errgroup.Group) {
		if err := eg.Wait(); err != nil {
			fmt.Println(err)
		}
	}(eg)

	for topic, partitions := range lost {
		for _, partition := range partitions {
			tp := consumerKey{topic, partition}
			pc := c.consumers[tp]
			delete(c.consumers, tp)
			close(pc.quit)
			fmt.Printf("waiting for work to finish on topic %s and partition %d\n", topic, partition)
			eg.Go(func() error {
				select {
				case <-time.After(c.cfg.CloseTimeout):
					return fmt.Errorf("timeout while trying to close consumer for topic %s and partition %d", topic, partition)
				case <-pc.done:
					fmt.Printf("succesfully closed consumer for topic %s and partition %d\n", topic, partition)
					return nil
				}
			})
		}
	}
}
