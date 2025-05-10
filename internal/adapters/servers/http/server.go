package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/handlers"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/routes"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/server"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
)

const (
	defaultIdleTimeout     = 2 * time.Minute
	defaultShutdownTimeout = 10 * time.Second
)

type Server struct {
	server     *http.Server
	config     server.Config
	authAPI    auth.UseCaseUser
	orchAPI    orchestrator.UseCaseCalculation
	handlers   *handlers.Handlers
	shutdownCh chan struct{}
}

func NewServer(config server.Config, authAPI auth.UseCaseUser, orchAPI orchestrator.UseCaseCalculation) *Server {
	return &Server{
		config:     config,
		authAPI:    authAPI,
		orchAPI:    orchAPI,
		handlers:   handlers.NewHandlers(authAPI, orchAPI),
		shutdownCh: make(chan struct{}),
	}
}

func (s *Server) Start(ctx context.Context) error {
	log := logger.ContextLogger(ctx, nil)
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	log.Info("Starting HTTP server",
		zap.String("address", addr),
		zap.Duration("read_timeout", s.config.ReadTimeout),
		zap.Duration("write_timeout", s.config.WriteTimeout))

	router := routes.NewRouter(s.authAPI, s.orchAPI)

	s.server = &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       s.config.ReadTimeout,
		WriteTimeout:      s.config.WriteTimeout,
		IdleTimeout:       defaultIdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error", zap.Error(err))
		}
		close(s.shutdownCh)
	}()

	log.Info("HTTP server started successfully")
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log := logger.ContextLogger(ctx, nil)
	log.Info("Stopping HTTP server")

	shutdownCtx, cancel := context.WithTimeout(ctx, defaultShutdownTimeout)
	defer cancel()

	err := s.server.Shutdown(shutdownCtx)
	if err != nil {
		log.Error("HTTP server shutdown error", zap.Error(err))
		return fmt.Errorf("server shutdown error: %w", err)
	}

	select {
	case <-s.shutdownCh:
		log.Info("HTTP server stopped successfully")
	case <-shutdownCtx.Done():
		log.Warn("HTTP server shutdown timed out")
		return fmt.Errorf("HTTP server shutdown context timed out: %w", ctx.Err())
	}

	return nil
}

func (s *Server) WaitForShutdown() {
	<-s.shutdownCh
}
