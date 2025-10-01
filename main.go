package main

import (
	"cloud/config"
	"cloud/core"
	"cloud/handlers"
	"cloud/transaction"
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

const yamlPath = "./configs/config.yaml"
const envPath = "./.env"

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cfg, err := config.LoadConfig(envPath, yamlPath)
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}
	log.Println(cfg)

	transactor, err := transaction.NewPostgresTransactor(ctx, cfg.Postgres)
	if err != nil {
		log.Fatalf("failed to create transaction logger: %s", err)
	}
	defer transactor.Close()

	store := core.NewStore(transactor)
	handler := handlers.NewHandler(store)

	r := mux.NewRouter()
	r.HandleFunc("/", handler.HelloGoHandler)
	r.HandleFunc("/v1/{key}", handler.PutHandler).Methods("PUT")
	r.HandleFunc("/v1/{key}", handler.GetHandler).Methods("GET")
	r.HandleFunc("/v1/{key}", handler.DeleteHandler).Methods("DELETE")

	log.Println("server is starting...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
