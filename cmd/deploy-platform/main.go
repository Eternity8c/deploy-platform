package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_server "github.com/Eternity8c/deploy-platform/internal/core/server"
	core_transport_middleware "github.com/Eternity8c/deploy-platform/internal/core/transport/middleware"
	features_handler "github.com/Eternity8c/deploy-platform/internal/features/handler"
)

func main() {
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	configLog := core_logger.NewConfigMust()
	logger, err := core_logger.NewLogger(configLog)
	if err != nil {
		fmt.Printf("created logger: %w", err)
		os.Exit(1)
	}
	defer logger.Close()

	configServer := core_server.NewConfigMust()
	server := core_server.NewServer(
		configServer,
		logger,
		core_transport_middleware.Logger(logger),
		core_transport_middleware.Panic(),
	)

	router := core_server.NewRouter()
	router.RegisterRoutes(core_server.Route{
		Method:  http.MethodGet,
		Path:    "/healthz",
		Handler: features_handler.Healthz,
	})

	server.RegisterRoutes(router)

	server.Run(ctx)
}
