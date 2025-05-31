package valkey

import (
	"context"
	"encoding/json"

	"github.com/valkey-io/valkey-go"

	"github.com/mfelipe/go-feijoada/stream-buffer/config"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
)

func New(serverCfg config.Server, streamCfg config.Stream) *stream {
	opts := valkey.MustParseURL(serverCfg.Address)
	opts.Username = serverCfg.Username
	opts.Password = serverCfg.Password
	opts.ClientName = serverCfg.ClientName

	client, err := valkey.NewClient(opts)
	if err != nil {
		panic(err)
	}

	// create stream and consumer group if not exists
	client.Do(context.Background(), client.B().XgroupCreate().Key(streamCfg.Name).Group(streamCfg.Group).Id("0").Mkstream().Build())

	return &stream{
		cfg:          streamCfg,
		consumerName: serverCfg.ClientName,
		client:       client,
	}
}

type stream struct {
	cfg          config.Stream
	client       valkey.Client
	consumerName string
}

func (s *stream) Add(ctx context.Context, message models.Message) error {
	return s.client.Do(ctx, s.client.B().Xadd().Key(s.cfg.Name).Nomkstream().Id("*").FieldValue().FieldValue(models.DataFieldName, string(message.Data)).FieldValue(models.SchemaFieldName, message.SchemaURI).Build()).Error()
}

func (s *stream) ReadGroup(ctx context.Context) (map[string]models.Message, error) {
	resp := s.client.Do(ctx, s.client.B().Xreadgroup().Group(s.cfg.Group, s.consumerName).Block(s.cfg.Block.Milliseconds()).Streams().Key(s.cfg.Name, s.cfg.Name).Id(">", "0").Build())
	if resp.Error() != nil {
		return nil, resp.Error()
	}

	entriesArrayMap, err := resp.AsXRead()
	if err != nil {
		return nil, err
	}

	messageMap := make(map[string]models.Message)
	for _, entriesMap := range entriesArrayMap {
		for _, message := range entriesMap {
			messageMap[message.ID] = models.Message{
				SchemaURI: message.FieldValues[models.SchemaFieldName],
				Data:      json.RawMessage(message.FieldValues[models.DataFieldName]),
			}
		}
	}

	return messageMap, nil
}

func (s *stream) Ack(ctx context.Context, ids ...string) error {
	return s.client.Do(ctx, s.client.B().Xack().Key(s.cfg.Name).Group(s.cfg.Group).Id(ids...).Build()).Error()
}

func (s *stream) Delete(ctx context.Context, ids ...string) error {
	return s.client.Do(ctx, s.client.B().Xdel().Key(s.cfg.Name).Id(ids...).Build()).Error()
}
