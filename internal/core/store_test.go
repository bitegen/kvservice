package core

import (
	"cloud/internal/mocks"
	"context"
	"errors"
	"testing"
)

func TestPut(t *testing.T) {
	var (
		ctx   = context.Background()
		store = NewStore(&mocks.MockTransactor{})
	)

	const key = "create-key-put"
	const value = "create-value-put"

	t.Run("Successful Put", func(t *testing.T) {
		_, contains := store.m[key]
		if contains {
			t.Error("key/value already exists")
		}

		err := store.Put(ctx, key, value)
		if err != nil {
			t.Error(err)
		}
		defer store.delete(key)

		val, contains := store.m[key]
		if !contains {
			t.Error("create failed")
		}

		if val != value {
			t.Error("val/value mismatch")
		}
	})

	t.Run("Empty Key", func(t *testing.T) {
		err := store.Put(ctx, "", value)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("Empty Value", func(t *testing.T) {
		err := store.Put(ctx, key, "")
		if err != nil {
			t.Error(err)
		}

		val, contains := store.m[key]
		if !contains {
			t.Error("create failed for empty value")
		}

		if val != "" {
			t.Error("val/value mismatch for empty value")
		}
	})
}

func TestGet(t *testing.T) {
	var (
		ctx   = context.Background()
		store = NewStore(&mocks.MockTransactor{})
	)

	const key = "create-key-get"
	const value = "create-value-get"

	t.Run("Successful Get", func(t *testing.T) {
		err := store.Put(ctx, key, value)
		if err != nil {
			t.Error(err)
		}
		defer store.Delete(ctx, key)

		val, err := store.Get(ctx, key)
		if err != nil {
			t.Error(err)
		}

		if val != value {
			t.Errorf("val/value mismatch: got %s, want %s", val, value)
		}
	})

	t.Run("Key Not Found", func(t *testing.T) {
		_, err := store.Get(ctx, "non-existent-key")
		if !errors.Is(err, ErrKeyNotFound) {
			t.Errorf("expected error %v, got %v", ErrKeyNotFound, err)
		}
	})

	t.Run("Empty Key", func(t *testing.T) {
		_, err := store.Get(ctx, "")
		if !errors.Is(err, ErrEmptyKey) {
			t.Errorf("expected error %v, got %v", ErrEmptyKey, err)
		}
	})
}

func TestDelete(t *testing.T) {
	var (
		ctx   = context.Background()
		store = NewStore(&mocks.MockTransactor{})
	)

	const key = "create-key-delete"
	const value = "create-value-delete"

	t.Run("Successful Delete", func(t *testing.T) {
		err := store.Put(ctx, key, value)
		if err != nil {
			t.Error(err)
		}

		err = store.Delete(ctx, key)
		if err != nil {
			t.Error(err)
		}

		_, contains := store.m[key]
		if contains {
			t.Error("key still exists after deletion")
		}
	})

	t.Run("Delete Non-Existent Key", func(t *testing.T) {
		err := store.Delete(ctx, "non-existent-key")
		if err != nil {
			t.Error("expected no error, got:", err)
		}
	})

	t.Run("Delete Empty Key", func(t *testing.T) {
		err := store.Delete(ctx, "")
		if !errors.Is(err, ErrEmptyKey) {
			t.Errorf("expected error %v, got %v", ErrEmptyKey, err)
		}
	})
}
