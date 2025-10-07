package handlers

import (
	"cloud/core"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

type Handler struct {
	store core.Store
}

func NewHandler(store core.Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) HelloGoHandler(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintln(w, "Hello World!")
}

func (h *Handler) PutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		http.Error(w, "empty key", http.StatusBadRequest)
		return
	}

	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.store.Put(r.Context(), key, string(value))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DeleteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	err := h.store.Delete(r.Context(), key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		http.Error(w, "empty key", http.StatusBadRequest)
		return
	}

	value, err := h.store.Get(r.Context(), key)
	if errors.Is(err, core.ErrKeyNotFound) {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(value))
}
