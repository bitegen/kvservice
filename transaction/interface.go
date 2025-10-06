package transaction

import (
	"context"
)

type Transactor interface {
	WritePut(ctx context.Context, key, value string) error
	WriteDelete(ctx context.Context, key string) error

	ReadEvents() (<-chan Event, <-chan error)

	Close() error
}
