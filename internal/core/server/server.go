package core_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	core_logger "github.com/Eternity8c/deploy-platform/internal/core/logger"
	core_transport_middleware "github.com/Eternity8c/deploy-platform/internal/core/transport/middleware"
)

type Server struct {
	mux    *http.ServeMux
	config Config
	log    *core_logger.Logger

	middleware []core_transport_middleware.Middleware
}

func NewServer(config Config, log *core_logger.Logger, middleware ...core_transport_middleware.Middleware) *Server {
	return &Server{
		mux:        http.NewServeMux(),
		config:     config,
		log:        log,
		middleware: middleware,
	}
}

func (s *Server) RegisterRoutes(routers ...*Router) {
	for _, router := range routers {
		s.mux.Handle(
			"/",
			router,
		)
	}
}

func (s *Server) Run(ctx context.Context) error {
	mux := core_transport_middleware.ChainMiddleware(s.mux, s.middleware...)

	server := &http.Server{
		ReadHeaderTimeout: 5 * time.Second,
		Addr:              s.config.Addr,
		Handler:           mux,
	}

	ch := make(chan error, 1)

	go func() {
		defer close(ch)

		s.log.Warn("start server", zap.String("addr", s.config.Addr))

		err := server.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			ch <- err
		}
	}()

	select {
	case err := <-ch:
		if err != nil {
			return fmt.Errorf("listen and serve: %w", err)
		}
	case <-ctx.Done():
		s.log.Warn("shutdown server...")

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			s.config.ShutdownTimeout,
		)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()

			return fmt.Errorf("shutdown server err: %w", err)
		}
	}
	return nil
}
