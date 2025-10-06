package main

import (
	"cloud/config"
	"net/http"
)

func NewServer(cfg config.ServerConfig) *http.Server {
	server := &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return server
}
