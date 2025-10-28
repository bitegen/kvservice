package server

import (
	"cloud/internal/handlers"
	"cloud/internal/middleware"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(h *handlers.Handler, logger *slog.Logger) http.Handler {
	r := mux.NewRouter()

	r.HandleFunc("/", h.HelloGoHandler)
	r.HandleFunc("/v1/{key}", h.PutHandler).Methods(http.MethodPut)
	r.HandleFunc("/v1/{key}", h.GetHandler).Methods(http.MethodGet)
	r.HandleFunc("/v1/{key}", h.DeleteHandler).Methods(http.MethodDelete)

	chain := middleware.Logging(logger)(
		middleware.Recover(logger)(r),
	)

	return chain
}
