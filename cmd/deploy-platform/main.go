package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_repo_postgres "github.com/Eternity8c/deploy-platform/internal/core/repository/postgres"
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
		fmt.Println("created logger:", err)
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

	configBD := core_repo_postgres.NewConfigMust()
	pool, err := core_repo_postgres.NewConnectionPool(ctx, configBD)
	if err != nil {
		logger.Error("pool:", zap.Error(err))
		os.Exit(1)
	}
	defer pool.Close()

	router := core_server.NewRouter()
	router.RegisterRoutes(
		core_server.Route{
			Method:  http.MethodGet,
			Path:    "/healthz",
			Handler: features_handler.Healthz,
		},
		core_server.Route{
			Method:  http.MethodGet,
			Path:    "/readyz",
			Handler: features_handler.Readyz(pool),
		},
		core_server.Route{
			Method:  http.MethodGet,
			Path:    "/metrics",
			Handler: promhttp.Handler().ServeHTTP,
		},
	)

	server.RegisterRoutes(router)

	if err := server.Run(ctx); err != nil {
		logger.Error("run server: ", zap.Error(err))
	}
}
