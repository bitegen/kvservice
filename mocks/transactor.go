package mocks

import "context"

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
