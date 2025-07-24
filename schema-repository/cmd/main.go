package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	zlog "github.com/rs/zerolog/log"
	"github.com/tavsec/gin-healthcheck/checks"

	healthcheck "github.com/tavsec/gin-healthcheck"
	hcconfig "github.com/tavsec/gin-healthcheck/config"

	"github.com/mfelipe/go-feijoada/schema-repository/config"
	"github.com/mfelipe/go-feijoada/schema-repository/internal/handlers"
	"github.com/mfelipe/go-feijoada/schema-repository/internal/repository"
	"github.com/mfelipe/go-feijoada/schema-repository/internal/service"
	utilslog "github.com/mfelipe/go-feijoada/utils/log"
)

func main() {
	select {
	case <-startServer():
		zlog.Info().Msg("server stopped")
	}
}

func startServer() chan error {
	cfg := config.Load()

	//Set global log level
	utilslog.InitializeGlobal(cfg.Log)

	// Initialize Gin router
	router := gin.New(utilslog.CustomLoggerRecovery())

	err := healthcheck.New(router, hcconfig.DefaultConfig(), []checks.Check{})
	if err != nil {
		panic(err)
	}

	// Create the repository client based on the configuration
	repo := repository.NewRepository(cfg.Repository)

	// Create the SchemaService
	schemaSvc := service.NewSchemaService(cfg.Repository.Data, repo)

	// Create the handler instance
	// The NewHandler function now accepts the service.
	apiHandler := handlers.NewHandler(schemaSvc)

	// Registers custom validation functions for using during request binding
	handlers.RegisterCustomValidators()

	// Register routes
	router.Group("/schemas/:name/:version").
		GET("", apiHandler.GetSchemaHandler).
		POST("", apiHandler.CreateSchemaHandler).
		DELETE("", apiHandler.DeleteSchemaHandler)

	// Start the server
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	zlog.Info().Msgf("starting server on %s", serverAddr)

	run := make(chan error)
	go func() {
		run <- router.Run(serverAddr)
	}()
	return run
}
