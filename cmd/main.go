package main

import (
	"cloud/internal/config"
	"cloud/internal/core"
	"cloud/internal/handlers"
	"cloud/internal/transaction"
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

var (
	defaultYAML = "./configs/config.yaml"
	defaultEnv  = "./.env"
)

var (
	yamlPathParsed = flag.String("config", defaultYAML, "path to the YAML configuration file")
	envPathParsed  = flag.String("env", defaultEnv, "path to the .env file")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	cfg, err := config.LoadConfig(*envPathParsed, *yamlPathParsed)
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}

	transactor, err := transaction.NewFileTransactor(ctx)
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

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HelloGoHandler)
	r.HandleFunc("/v1/{key}", handler.PutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", handler.GetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", handler.DeleteHandler).Methods("DELETE")

	log.Println("server is starting...")
	srv := NewServer(cfg.HTTP)
	srv.Handler = r

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("server is listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	<-stop
	log.Println("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped gracefully")
}

func NewServer(cfg config.ServerConfig) *http.Server {
	server := &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return server
}
