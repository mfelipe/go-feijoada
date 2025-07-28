package testcontainers

import (
	"context"
	"errors"

	zlog "github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
)

func validateContainerStart(ctx context.Context, container testcontainers.Container, runErr error) {
	if runErr != nil {
		zlog.Err(runErr).Msg("failed to start container")
		if container != nil {
			runErr = errors.Join(runErr, container.Terminate(ctx))
		}
		panic(runErr)
	}

	state, err := container.State(ctx)
	if err != nil {
		zlog.Err(err).Msg("failed to get container state")
		err = errors.Join(err, container.Terminate(ctx))
		panic(err)
	}

	zlog.Info().Msgf("%T status: %s", container, state.Status)
}
