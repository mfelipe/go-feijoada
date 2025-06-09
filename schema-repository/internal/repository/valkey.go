package repository

import (
	"context"
	"errors"

	"github.com/valkey-io/valkey-go"

	"github.com/mfelipe/go-feijoada/schema-repository/internal/clients"
)

type valkeyClient struct {
	client clients.Valkey
}

func (v *valkeyClient) Set(ctx context.Context, key string, value string) error {
	return v.client.Do(ctx, v.client.B().Set().Key(key).Value(value).Build()).Error()
}

func (v *valkeyClient) Del(ctx context.Context, keys ...string) error {
	val, err := v.client.Do(ctx, v.client.B().Del().Key(keys...).Build()).ToInt64()

	if val == 0 && err == nil {
		return errors.New(ErrorKeyNotFound)
	}

	return err
}

func (v *valkeyClient) Get(ctx context.Context, key string) (string, error) {
	val, err := v.client.Do(ctx, v.client.B().Get().Key(key).Build()).ToString()

	if val == "" && valkey.IsValkeyNil(err) {
		return val, errors.New(ErrorKeyNotFound)
	}

	return val, err
}
