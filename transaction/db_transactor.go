package transaction

import (
	"cloud/config"
	"cloud/utils"
	"context"
	"fmt"
	"sync/atomic"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTransactor struct {
	events       chan Event
	errors       chan error
	done         chan struct{}
	lastSequence uint64
	closed       uint32 // 0 if open, 1 if closed
	pool         *pgxpool.Pool
}

func NewPostgresTransactor(ctx context.Context, cfg config.PostgresConfig) (*PostgresTransactor, error) {
	dsn := utils.MakeDSN(cfg)

	psqlConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pg config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, psqlConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to db: %w", err)
	}

	t := &PostgresTransactor{
		events: make(chan Event, 128),
		errors: make(chan error, 1),
		done:   make(chan struct{}),
		pool:   pool,
	}

	return t, nil
}

func (t *PostgresTransactor) Close() error {
	if !atomic.CompareAndSwapUint32(&t.closed, 0, 1) {
		return nil
	}
	close(t.done) // release all goroutines

	t.pool.Close()
	return nil
}

func (t *PostgresTransactor) WritePut(ctx context.Context, key, value string) error {
	return t.send(ctx, Event{Key: key, Value: value})
}

func (t *PostgresTransactor) WriteDelete(ctx context.Context, key string) error {
	return t.send(ctx, Event{Key: key})
}

func (t *PostgresTransactor) send(ctx context.Context, event Event) error {
	if atomic.LoadUint32(&t.closed) == 1 {
		return ErrTransactorClosed
	}

	select {
	case <-ctx.Done():
		return ErrTransactorClosed
	case t.events <- event:
		return nil
	case <-t.done:
		return nil
	}
}

// func (t *PostgresTransactor) run(ctx context.Context) {
// 	go func() {
// 		for {
// 			select {
// 			case event := <-t.events:
// 				t.lastSequence++

// 				journalRow := fmt.Sprintf(
// 					"%d\t%d\t%s\t%s\n",
// 					t.lastSequence, event.EventType, event.Key, event.Value)

// 				_, err := t.file.WriteString(journalRow)
// 				if err != nil {
// 					t.errors <- err
// 				}
// 			case <-t.done:
// 				return
// 			case <-ctx.Done():
// 				return
// 			}
// 		}
// 	}()
// }

func (t *PostgresTransactor) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)
	close(outError)
	close(outEvent)
	return outEvent, outError
}
