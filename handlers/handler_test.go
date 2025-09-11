package handlers

import (
	"bytes"
	"cloud/core"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestHelloGoHandler(t *testing.T) {
	store := core.NewStore()
	handler := NewHandler(store)

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler.HelloGoHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler got, %v want %v", status, http.StatusOK)
	}

	expected := "Hello World!\n"
	if rr.Body.String() != expected {
		t.Errorf("handler got %v, want %v", rr.Body.String(), expected)
	}
}

func TestPutHandler(t *testing.T) {
	store := core.NewStore()
	handler := NewHandler(store)

	key := "key1"
	value := "value1"

	req, err := http.NewRequest("PUT", "/v1/{key}", bytes.NewBuffer([]byte(value)))
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{"key": key})

	rr := httptest.NewRecorder()
	handler.PutHandler(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler got, %v want %v", status, http.StatusCreated)
	}

	if storedValue, err := store.Get(key); (err != nil) || storedValue != value {
		t.Errorf("handler got, %v want %v", value, storedValue)
	}
}

func TestGetHandler(t *testing.T) {
	store := core.NewStore()
	handler := NewHandler(store)

	key := "key1"
	value := "value1"
	store.Put(key, value)

	req, err := http.NewRequest("GET", "/v1/{key}", nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{"key": key})

	rr := httptest.NewRecorder()
	handler.GetHandler(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler got, %v want %v", status, http.StatusOK)
	}

	if bodyValue := rr.Body.String(); bodyValue != value {
		t.Errorf("handler got, %v want %v", bodyValue, value)
	}
}
