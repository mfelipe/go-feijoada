package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/mfelipe/go-feijoada/kafka-consumer/config"
	"github.com/mfelipe/go-feijoada/kafka-consumer/internal"
)

func main() {
	cfg := config.Load()
	consumer := internal.NewConsumer(cfg)

	go consumer.Poll()

	sigs := make(chan os.Signal, 2)
	signal.Notify(sigs, os.Interrupt)

	<-sigs
	fmt.Println("received interrupt signal; closing client")
	done := make(chan struct{})
	go func() {
		defer close(done)
		consumer.Close()
	}()

	select {
	case <-sigs:
		fmt.Println("received second interrupt signal; quitting without waiting for graceful close")
	case <-done:
	}
}
