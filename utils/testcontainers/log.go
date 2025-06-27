package testcontainers

import (
	"fmt"

	"github.com/testcontainers/testcontainers-go"
)

func NewStdoutLogConsumer(prefix string) testcontainers.LogConsumer {
	return &stdoutLogConsumer{prefix: prefix}
}

// StdoutLogConsumer is a LogConsumer that prints the log to stdout
type stdoutLogConsumer struct {
	prefix string
}

// Accept prints the log to stdout
func (lc *stdoutLogConsumer) Accept(l testcontainers.Log) {
	fmt.Printf("[%s] %s", lc.prefix, string(l.Content))
}
