package testcontainers

import (
	"context"
	"errors"

	"github.com/rs/zerolog/log"
	"github.com/testcontainers/testcontainers-go"
)

func validateContainerStart(ctx context.Context, container testcontainers.Container, runErr error) {
	if runErr != nil {
		log.Err(runErr).Msg("failed to start container")
		if container != nil {
			runErr = errors.Join(runErr, container.Terminate(ctx))
		}
		panic(runErr)
	}

	state, err := container.State(ctx)
	if err != nil {
		log.Err(err).Msg("failed to get container state")
		err = errors.Join(err, container.Terminate(ctx))
		panic(err)
	}

	log.Info().Msgf("%T status: %s", container, state.Status)
}
