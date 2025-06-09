package testcontainers

import (
	"context"

	"github.com/testcontainers/testcontainers-go"
	tcvalkey "github.com/testcontainers/testcontainers-go/modules/valkey"
)

func StartValkey(ctx context.Context) *tcvalkey.ValkeyContainer {
	valkeyContainer, err := tcvalkey.Run(ctx,
		"valkey/valkey:7.2.5",
		tcvalkey.WithSnapshotting(10, 1),
		tcvalkey.WithLogLevel(tcvalkey.LogLevelVerbose),
		testcontainers.WithLogConsumers(NewStdoutLogConsumer("valkey")),
	)
	validateContainerStart(ctx, valkeyContainer, err)
	return valkeyContainer
}
