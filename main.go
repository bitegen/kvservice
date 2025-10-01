package main

import (
	"cloud/core"
	"cloud/handlers"
	"cloud/transaction"
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	ctx := context.Background()

	transactor, err := transaction.NewFileTransactor(ctx)
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
