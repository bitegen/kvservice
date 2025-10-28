package server

import (
	"cloud/internal/config"
	"context"
	"fmt"
	"log/slog"
	"net/http"
)

type HTTPServer struct {
	srv       *http.Server
	logger    *slog.Logger
	errCh     chan error
	isRunning bool
}

func NewServer(cfg config.ServerConfig, logger *slog.Logger, routes http.Handler) *HTTPServer {
	server := &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      routes,
	}
	return &HTTPServer{
		srv:    server,
		logger: logger,
		errCh:  make(chan error, 1),
	}
}

func (s *HTTPServer) Start() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			s.errCh <- fmt.Errorf("listen error: %v", err)
		}
	}()

	s.logger.Info("http server started", "addr", s.srv.Addr)
}

func (s *HTTPServer) Stop(ctx context.Context) error {
	if !s.isRunning {
		return nil
	}

	s.isRunning = false
	if err := s.srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown error: %v", err)
	}

	s.logger.Info("http server stopped")
	return nil
}

func (s *HTTPServer) ErrChan() <-chan error {
	return s.errCh
}
