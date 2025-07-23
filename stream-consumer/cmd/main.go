package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"

	"github.com/mfelipe/go-feijoada/stream-consumer/config"
	"github.com/mfelipe/go-feijoada/stream-consumer/internal/consumer"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

func main() {
	cfg := config.Load()

	// Set global log level
	utilslog.InitializeGlobal(cfg.Log)

	// Create consumer
	w, err := consumer.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create consumer")
	}

	// Setup context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info().Str("signal", sig.String()).Msg("received shutdown signal")
		cancel()
	}()

	// Start the consumer
	log.Info().Msg("starting Stream Consumer...")
	if err = w.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("consumer failed")
	}

	// Close consumer
	w.Close()

	log.Info().Msg("Stream Consumer shutdown complete")
}
