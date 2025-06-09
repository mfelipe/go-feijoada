package clients

import (
	"context"

	"github.com/valkey-io/valkey-go"

	"github.com/mfelipe/go-feijoada/schema-repository/config"
)

type Valkey interface {
	B() valkey.Builder
	Do(ctx context.Context, cmd valkey.Completed) (resp valkey.ValkeyResult)
}

func NewValkeyClient(cfg config.RepoServer) Valkey {
	opts := valkey.MustParseURL(cfg.Address)
	opts.Username = cfg.Username
	opts.Password = cfg.Password
	opts.ClientName = cfg.ClientName

	client, err := valkey.NewClient(opts)
	if err != nil {
		panic(err)
	}

	return client
}
