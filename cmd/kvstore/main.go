package main

import (
	"cloud/internal/config"
	"cloud/internal/core"
	"cloud/internal/handlers"
	"cloud/internal/server"
	"cloud/internal/transaction"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	cfg := config.MustLoad()

	transactorFactory := transaction.NewTransactorFactory(cfg)
	transactor, err := transactorFactory.Create(ctx, transaction.TransactorTypePostgres)
	if err != nil {
		log.Fatalf("failed to create transaction logger: %s", err)
	}
	defer func() {
		if err := transactor.Close(); err != nil {
			log.Fatalf("failed to close transactor: %v", err)
		}
	}()

	store := core.NewStore(transactor)
	handler := handlers.NewHandler(store)

	routes := server.NewRouter(handler)
	srv := server.NewServer(cfg.HTTP, routes)
	srv.Start()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Println(sig)
	case err := <-srv.ErrChan():
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
