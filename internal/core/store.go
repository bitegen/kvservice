package core

import (
	"cloud/internal/transaction"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrEmptyKey    = errors.New("key is empty")
)

type inMemoryStore struct {
	m          map[string]string
	log        *slog.Logger
	transactor transaction.Transactor
	sync.RWMutex
}

func NewStore(transactor transaction.Transactor, logger *slog.Logger) (*inMemoryStore, error) {
	st := &inMemoryStore{
		m:          make(map[string]string),
		log:        logger,
		transactor: transactor,
	}

	if err := st.restoreState(); err != nil {
		st.log.Error("failed to restore state",
			slog.Any("err", err))
		return nil, err
	}
	st.log.Debug("state is restored succesfull")

	return st, nil
}

// helper to check if key is empty
func (s *inMemoryStore) isKeyValid(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	return nil
}

func (s *inMemoryStore) Put(ctx context.Context, key string, value string) error {
	const op = "inMemoryStore.Put"

	log := s.log.With(
		slog.String("op", op),
	)

	if err := s.isKeyValid(key); err != nil {
		log.Error("empty key", slog.Any("error", ErrEmptyKey))
		return err
	}

	s.put(key, value) // add pair in lock

	err := s.transactor.WritePut(context.TODO(), key, value)
	if err != nil {
		s.delete(key)
		log.Error("journal write failed, rollback", slog.Any("error", err))
		return fmt.Errorf("failed to log put operation: %w", err)
	}

	log.Info("put succeeded")
	return nil

}

func (s *inMemoryStore) Delete(ctx context.Context, key string) error {
	const op = "inMemoryStore.Delete"

	log := s.log.With(
		slog.String("op", op),
	)

	if err := s.isKeyValid(key); err != nil {
		log.Error("empty key", slog.Any("error", ErrEmptyKey))
		return err
	}

	s.delete(key) // delete pair in lock

	err := s.transactor.WriteDelete(context.TODO(), key)
	if err != nil {
		s.delete(key)
		log.Error("journal write failed, rollback", slog.Any("error", err))
		return fmt.Errorf("failed to log delete operation: %w", err)
	}

	log.Info("delete succeeded")
	return nil
}

func (s *inMemoryStore) Get(ctx context.Context, key string) (string, error) {
	const op = "inMemoryStore.Get"

	log := s.log.With(
		slog.String("op", op),
	)

	if err := s.isKeyValid(key); err != nil {
		log.Error("empty key", slog.Any("error", ErrEmptyKey))
		return "", err
	}

	s.RLock()
	defer s.RUnlock()

	value, ok := s.m[key]
	if !ok {
		log.Error("wrong key", slog.Any("error", ErrKeyNotFound))
		return "", ErrKeyNotFound
	}

	log.Info("get succeeded")
	return value, nil
}

// put data in lock
func (s *inMemoryStore) put(key string, value string) {
	s.Lock()
	defer s.Unlock()

	s.m[key] = value
}

// delete data in lock
func (s *inMemoryStore) delete(key string) {
	s.Lock()
	defer s.Unlock()

	delete(s.m, key)
}

func (s *inMemoryStore) restoreState() error {
	eventsCh, errCh := s.transactor.ReadEvents()
	if eventsCh == nil || errCh == nil {
		return transaction.ErrEmptyJournal
	}

	for event := range eventsCh {
		switch event.EventType {
		case transaction.EventDelete:
			s.delete(event.Key)
		case transaction.EventPut:
			s.put(event.Key, event.Value)
		default:
			return errors.New("unknown event to restore")
		}
	}

	if err := <-errCh; err != nil {
		return err
	}
	return nil
}
