package main

import (
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/mfelipe/go-feijoada/kafka-consumer/config"
	"github.com/mfelipe/go-feijoada/kafka-consumer/internal"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

func main() {
	cfg := config.Load()

	// Set global log level
	utilslog.InitializeGlobal(cfg.Log)

	// Create the kafka consumer
	consumer := internal.NewConsumer(*cfg)

	stopped := make(chan byte)
	go func() {
		defer close(stopped)
		consumer.Poll()
	}()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt, os.Kill)

	done := make(chan struct{})
	select {
	case <-sigs:
		log.Info().Msg("received interrupt signal. Stopping polling...")
		go func() {
			defer close(done)
			consumer.Close()
		}()
	case <-stopped:
		log.Info().Msg("kafka polling stopped. Exiting...")
		return
	}

	select {
	case <-time.After(time.Minute):
		log.Info().Msg("kafka consumer polling stop timeout; quitting without waiting for graceful shutdown")
	case <-sigs:
		log.Info().Msg("received second interrupt signal; quitting without waiting for graceful shutdown")
	case <-done:
		log.Info().Msg("kafka consumer was gracefully shutdown")
	}
}
