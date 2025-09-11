package core

import (
	"cloud/transaction"
	"context"
	"errors"
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
	return &Store{
		m:          make(map[string]string),
		transactor: transactor,
	}
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

	s.Lock()
	defer s.Unlock()

	s.m[key] = value
	s.transactor.WritePut(context.TODO(), key, value)
	return nil
}

func (s *Store) Delete(key string) error {
	if err := s.isKeyValid(key); err != nil {
		return err
	}

	s.Lock()
	defer s.Unlock()

	delete(s.m, key)
	s.transactor.WriteDelete(context.TODO(), key)
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
