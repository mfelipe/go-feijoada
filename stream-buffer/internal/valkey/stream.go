package valkey

import (
	"context"

	"github.com/valkey-io/valkey-go"

	"github.com/mfelipe/go-feijoada/stream-buffer/config"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
)

// client interface for mocking
type client interface {
	Do(ctx context.Context, cmd valkey.Completed) valkey.ValkeyResult
	B() valkey.Builder
}

//goland:noinspection GoExportedFuncWithUnexportedType
func New(serverCfg config.Server, streamCfg config.Stream, opts ...Option) *stream {
	s := stream{
		cfg: streamCfg,
	}

	for _, opt := range opts {
		opt(&s)
	}

	if s.cli == nil {
		vopts := valkey.MustParseURL(serverCfg.Address)
		vopts.Username = serverCfg.Username
		vopts.Password = serverCfg.Password
		vopts.ClientName = serverCfg.ClientName

		cli, err := valkey.NewClient(vopts)
		if err != nil {
			panic(err)
		}

		s.cli = cli
	}

	// create stream and consumer group if not exists
	s.cli.Do(context.Background(), s.cli.B().XgroupCreate().Key(streamCfg.Name).Group(streamCfg.Group).Id("0").Mkstream().Build())

	return &s
}

type stream struct {
	cfg config.Stream
	cli client
}

func (s *stream) Add(ctx context.Context, message models.Message) error {
	return s.cli.Do(ctx, s.cli.B().Xadd().Key(s.cfg.Name).Nomkstream().Id("*").FieldValue().FieldValueIter(message.Iter()).Build()).Error()
}

func (s *stream) ReadGroup(ctx context.Context) (map[string]models.Message, error) {
	resp := s.cli.Do(ctx, s.cli.B().Xreadgroup().Group(s.cfg.Group, s.cfg.Consumer).Block(s.cfg.Block.Milliseconds()).Streams().Key(s.cfg.Name, s.cfg.Name).Id(">", "0").Build())
	if resp.Error() != nil {
		return nil, resp.Error()
	}

	entriesArrayMap, err := resp.AsXRead()
	if err != nil {
		return nil, err
	}

	messageMap := make(map[string]models.Message)
	for _, entries := range entriesArrayMap {
		for _, entry := range entries {
			messageMap[entry.ID] = models.MessageFromValkeyValue(entry.FieldValues)
		}
	}

	return messageMap, nil
}

func (s *stream) Ack(ctx context.Context, ids ...string) error {
	return s.cli.Do(ctx, s.cli.B().Xack().Key(s.cfg.Name).Group(s.cfg.Group).Id(ids...).Build()).Error()
}

func (s *stream) Delete(ctx context.Context, ids ...string) error {
	return s.cli.Do(ctx, s.cli.B().Xdel().Key(s.cfg.Name).Id(ids...).Build()).Error()
}
