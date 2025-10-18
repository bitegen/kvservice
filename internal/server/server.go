package server

import (
	"cloud/internal/config"
	"context"
	"fmt"
	"net/http"
)

type HTTPServer struct {
	srv   *http.Server
	errCh chan error
}

func NewServer(cfg config.ServerConfig, routes http.Handler) *HTTPServer {
	server := &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		Handler:      routes,
	}
	return &HTTPServer{
		srv:   server,
		errCh: make(chan error, 1),
	}
}

func (s *HTTPServer) Start() {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.errCh <- fmt.Errorf("listen error: %v", err)
		}
	}()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.srv.Shutdown(ctx)
}

func (s *HTTPServer) ErrChan() <-chan error {
	return s.errCh
}
