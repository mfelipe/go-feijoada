package redis

import (
	"context"
	"errors"
	"sync"
	
	"github.com/redis/go-redis/v9"
	zlog "github.com/rs/zerolog/log"

	"github.com/mfelipe/go-feijoada/stream-buffer/config"
	"github.com/mfelipe/go-feijoada/stream-buffer/models"
)

const nilResult = "got an unexpected nil result from stream operation"

type client interface {
	XAdd(ctx context.Context, a *redis.XAddArgs) *redis.StringCmd
	XAck(ctx context.Context, stream, group string, ids ...string) *redis.IntCmd
	XDel(ctx context.Context, stream string, ids ...string) *redis.IntCmd
	XReadGroup(ctx context.Context, a *redis.XReadGroupArgs) *redis.XStreamSliceCmd
}

//goland:noinspection GoExportedFuncWithUnexportedType
func New(serverCfg config.Server, streamCfg config.Stream, opts ...Option) *stream {
	s := stream{
		cfg: streamCfg,
	}

	for _, opt := range opts {
		opt(&s)
	}

	if s.client == nil {
		var once sync.Once
		onConnFunc := func(ctx context.Context, cn *redis.Conn) error {
			var err error
			once.Do(func() {
				zlog.Debug().Str("stream", s.cfg.Name).Str("group", s.cfg.Group).Msg("trying to create redis stream group")
				status := cn.XGroupCreateMkStream(ctx, s.cfg.Name, s.cfg.Group, "0")
				if status.Err() != nil && !errors.Is(status.Err(), redis.Nil) {
					err = status.Err()
				}
			})
			return err
		}
		if serverCfg.IsCluster {
			c := redis.NewClusterClient(&redis.ClusterOptions{
				Addrs:      []string{serverCfg.Address},
				ClientName: serverCfg.ClientName,
				OnConnect:  onConnFunc,
				NewClient: func(opt *redis.Options) *redis.Client {
					opt.CredentialsProvider = func() (username string, password string) {
						return serverCfg.Username, serverCfg.Password
					}
					return redis.NewClient(opt)
				},
			})

			s.client = c
		} else {
			c := redis.NewClient(&redis.Options{
				Addr:      serverCfg.Address,
				OnConnect: onConnFunc,
				CredentialsProvider: func() (username string, password string) {
					return serverCfg.Username, serverCfg.Password
				},
			})

			s.client = c
		}
	}

	return &s
}

type stream struct {
	cfg    config.Stream
	client client
}

func (s *stream) Add(ctx context.Context, message models.Message) error {
	result := s.client.XAdd(ctx, &redis.XAddArgs{
		Stream:     s.cfg.Name,
		NoMkStream: true,
		Values:     message.ToValue(),
	})
	return resultError(result)
}

func (s *stream) ReadGroup(ctx context.Context) (map[string]models.Message, error) {
	result := s.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    s.cfg.Group,
		Consumer: s.cfg.Consumer,
		Streams:  []string{s.cfg.Name, s.cfg.Name, ">", "0"},
		Count:    s.cfg.ReadCount,
		Block:    s.cfg.Block,
	})

	if result == nil {
		return nil, errors.New(nilResult)
	}

	xStreams, err := result.Result()
	if err != nil {
		return nil, err
	}

	messageMap := make(map[string]models.Message)
	for _, xStream := range xStreams {
		for _, xMessage := range xStream.Messages {
			messageMap[xMessage.ID] = models.MessageFromRedisValue(xMessage.Values)
		}
	}

	return messageMap, nil

}

func (s *stream) Ack(ctx context.Context, ids ...string) error {
	result := s.client.XAck(ctx, s.cfg.Name, s.cfg.Group, ids...)
	return resultError(result)
}

func (s *stream) Delete(ctx context.Context, ids ...string) error {
	result := s.client.XDel(ctx, s.cfg.Name, ids...)
	return resultError(result)
}

type cmdErr interface {
	*redis.IntCmd | *redis.StringCmd
	Err() error
}

func resultError[T cmdErr](result T) error {
	var err error
	if result == nil {
		err = errors.New(nilResult)
	} else {
		err = result.Err()
	}

	return err
}
