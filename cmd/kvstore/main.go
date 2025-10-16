package main

import (
	"cloud/internal/config"
	"cloud/internal/core"
	"cloud/internal/handlers"
	"cloud/internal/server"
	"cloud/internal/transaction"
	"context"
	"log"
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
	srv.Run(ctx)
}
