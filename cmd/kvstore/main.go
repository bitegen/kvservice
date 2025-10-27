package main

import (
	"cloud/internal/config"
	"cloud/internal/core"
	"cloud/internal/handlers"
	"cloud/internal/logger"
	"cloud/internal/server"
	"cloud/internal/transaction"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()

	cfg := config.MustLoad()
	log := logger.NewLogger(cfg.Env)

	transactorFactory := transaction.NewTransactorFactory(cfg)
	transactor, err := transactorFactory.Create(ctx, transaction.TransactorTypePostgres)
	if err != nil {
		log.Error("failed to create transaction logger",
			slog.Any("error", err))
		os.Exit(1)
	}
	defer func() {
		if err := transactor.Close(); err != nil {
			log.Error("failed to close transactor", slog.Any("error", err))
		}
	}()

	store, err := core.NewStore(transactor, log)
	if err != nil {
		log.Error("failed to create store", slog.Any("err", err))
		os.Exit(1)
	}
	handler := handlers.NewHandler(store, log)

	routes := server.NewRouter(handler, log)
	srv := server.NewServer(cfg.HTTP, routes)
	srv.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("got sycall to finish service")
	case err := <-srv.ErrChan():
		log.Error("got err from server", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("got err from server shutdown", slog.Any("err", err))
		os.Exit(1)
	}
}
