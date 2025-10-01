package mocks

import (
	"cloud/transaction"
	"context"
)

type MockTransactor struct{}

func (t *MockTransactor) WritePut(context.Context, string, string) error {
	return nil
}

func (t *MockTransactor) WriteDelete(context.Context, string) error {
	return nil
}

func (t *MockTransactor) Close() error {
	return nil
}

func (t *MockTransactor) ReadEvents() (<-chan transaction.Event, <-chan error) {
	outEvent := make(chan transaction.Event)
	outError := make(chan error, 1)
	close(outError)
	close(outEvent)
	return outEvent, outError
}
