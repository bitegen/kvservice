package server

import (
	"cloud/internal/handlers"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(h *handlers.Handler) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/", h.HelloGoHandler)
	r.HandleFunc("/v1/{key}", h.PutHandler).Methods(http.MethodPut)
	r.HandleFunc("/v1/{key}", h.GetHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/{key}", h.DeleteHandler).Methods(http.MethodDelete)
	return r
}
