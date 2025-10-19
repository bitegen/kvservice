package handlers

import (
	"cloud/internal/core"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	store core.Store
	log   *slog.Logger
}

func NewHandler(store core.Store, log *slog.Logger) *Handler {
	return &Handler{
		store: store,
		log:   log,
	}
}

func (h *Handler) HelloGoHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "Hello World!")
}

func (h *Handler) PutHandler(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.PutHandler"

	log := h.log.With(
		slog.String("op", op),
	)

	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		log.Warn("empty key")
		http.Error(w, "empty key", http.StatusBadRequest)
		return
	}

	value, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("read body failed", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.store.Put(r.Context(), key, string(value))
	if err != nil {
		log.Error("put failed", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("value stored", slog.Int("size", len(value)))
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.DeleteHandler"

	log := h.log.With(
		slog.String("op", op),
	)

	vars := mux.Vars(r)
	key := vars["key"]

	err := h.store.Delete(r.Context(), key)
	if err != nil {
		log.Error("delete failed", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("value deleted")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	const op = "Handler.GetHandler"

	log := h.log.With(
		slog.String("op", op),
	)

	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		log.Warn("empty key")
		http.Error(w, "empty key", http.StatusBadRequest)
		return
	}

	value, err := h.store.Get(r.Context(), key)
	if errors.Is(err, core.ErrKeyNotFound) {
		log.Error("get failed", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Info("value retrieved")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(value))
}
