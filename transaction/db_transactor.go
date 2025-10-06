package transaction

import (
	"cloud/config"
	"cloud/migrator"
	"cloud/utils"
	"context"
	"fmt"
	"log"
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

	if err := migrator.RunMigrations(cfg, cfg.MigrationsDir); err != nil {
		return nil, fmt.Errorf("cannot run migrations: %w", err)
	}

	t := &PostgresTransactor{
		events: make(chan Event, 128),
		errors: make(chan error, 1),
		done:   make(chan struct{}),
		pool:   pool,
	}
	t.run(ctx)

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
	return t.send(ctx, Event{Key: key, Value: value, EventType: EventPut})
}

func (t *PostgresTransactor) WriteDelete(ctx context.Context, key string) error {
	return t.send(ctx, Event{Key: key, EventType: EventDelete})
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
		return ErrTransactorClosed
	}
}

func (t *PostgresTransactor) run(ctx context.Context) {
	go func() {
		query := `INSERT INTO transactions
			(event_type, key, value)
			VALUES ($1, $2, $3)`

		for event := range t.events {
			_, err := t.pool.Exec(
				context.TODO(),
				query,
				event.EventType, event.Key, event.Value)

			log.Println("store event: ", event)

			if err != nil {
				t.errors <- err
			}
		}
	}()
}

func (t *PostgresTransactor) ReadEvents() (<-chan Event, <-chan error) {
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	query := "SELECT sequence, event_type, key, value FROM transactions"

	go func() {
		defer close(outEvent)
		defer close(outError)

		rows, err := t.pool.Query(context.TODO(), query)
		if err != nil {
			outError <- fmt.Errorf("sql query error: %w", err)
			return
		}
		defer rows.Close()

		var e Event

		for rows.Next() {
			err = rows.Scan(&e.Sequence, &e.EventType, &e.Key, &e.Value)
			log.Println("get event: ", e)

			if err != nil {
				outError <- err
				return
			}

			outEvent <- e
		}

		err = rows.Err()
		if err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
		}
	}()

	return outEvent, outError
}
