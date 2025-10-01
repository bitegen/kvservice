package core

import (
	"cloud/transaction"
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrEmptyKey    = errors.New("key is empty")
)

type Store struct {
	m          map[string]string
	transactor transaction.Transactor
	sync.RWMutex
}

func NewStore(transactor transaction.Transactor) *Store {
	st := &Store{
		m:          make(map[string]string),
		transactor: transactor,
	}

	if err := st.restoreState(); err != nil {
		log.Fatal(err)
	}

	return st
}

// helper to check if key is empty
func (s *Store) isKeyValid(key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	return nil
}

func (s *Store) Put(key string, value string) error {
	if err := s.isKeyValid(key); err != nil {
		return err
	}

	s.put(key, value) // add pair in lock

	err := s.transactor.WritePut(context.TODO(), key, value)
	if err != nil {
		s.delete(key)
		return fmt.Errorf("failed to log put operation: %w", err)
	}
	return nil
}

func (s *Store) Delete(key string) error {
	if err := s.isKeyValid(key); err != nil {
		return err
	}

	s.delete(key) // delete pair in lock

	err := s.transactor.WriteDelete(context.TODO(), key)
	if err != nil {
		s.delete(key)
		return fmt.Errorf("failed to log delete operation: %w", err)
	}
	return nil
}

func (s *Store) Get(key string) (string, error) {
	if err := s.isKeyValid(key); err != nil {
		return "", err
	}

	s.RLock()
	defer s.RUnlock()

	value, ok := s.m[key]
	if !ok {
		return "", ErrKeyNotFound
	}
	return value, nil
}

// put data in lock
func (s *Store) put(key string, value string) {
	s.Lock()
	defer s.Unlock()

	s.m[key] = value
}

// delete data in lock
func (s *Store) delete(key string) {
	s.Lock()
	defer s.Unlock()

	delete(s.m, key)
}

func (s *Store) restoreState() error {
	eventsCh, errCh := s.transactor.ReadEvents()
	if eventsCh == nil || errCh == nil {
		return nil
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
