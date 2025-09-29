package transaction

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"sync/atomic"
)

const filename = "transactor.journal"

var (
	ErrTransactorClosed = errors.New("file transactor is closed")
	ErrOutOfSequence    = errors.New("transaction numbers out of sequence")
)

type FileTransactor struct {
	events       chan Event
	errors       chan error
	done         chan struct{}
	lastSequence uint64
	closed       uint32
	file         *os.File
}

func NewFileTransactor(ctx context.Context, filename string) (*FileTransactor, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	t := &FileTransactor{
		events: make(chan Event, 128),
		errors: make(chan error),
		done:   make(chan struct{}),
		file:   file,
	}
	t.run(ctx)

	return t, nil
}

func (t *FileTransactor) Close() error {
	if !atomic.CompareAndSwapUint32(&t.closed, 0, 1) {
		return ErrTransactorClosed
	}
	close(t.done)

	if err := t.file.Sync(); err != nil {
		return fmt.Errorf("sync error: %w", err)
	}
	return t.file.Close()
}

func (t *FileTransactor) WritePut(ctx context.Context, key, value string) error {
	return t.send(ctx, Event{Key: key, Value: value, EventType: EventPut})
}

func (t *FileTransactor) WriteDelete(ctx context.Context, key string) error {
	return t.send(ctx, Event{Key: key, Value: "", EventType: EventDelete})
}

func (t *FileTransactor) send(ctx context.Context, event Event) error {
	if atomic.LoadUint32(&t.closed) == 1 {
		return ErrTransactorClosed
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case t.events <- event:
		return nil
	}
}

func (t *FileTransactor) run(ctx context.Context) {
	go func() {
		for {
			select {
			case event := <-t.events:
				t.lastSequence++

				journalRow := fmt.Sprintf(
					"%d\t%d\t%s\t%s\n",
					t.lastSequence, event.EventType, event.Key, event.Value)

				_, err := t.file.WriteString(journalRow)
				if err != nil {
					t.errors <- err
				}
			case <-t.done:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

func (t *FileTransactor) readEvents() (chan Event, chan error) {
	scanner := bufio.NewScanner(t.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event

		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()

			fmt.Sscanf(
				line, "%d\t%d\t%s\t%s",
				&e.Sequence, &e.EventType, &e.Key, &e.Value)

			if t.lastSequence >= e.Sequence {
				outError <- ErrOutOfSequence
				return
			}

			uv, err := url.QueryUnescape(e.Value)
			if err != nil {
				outError <- fmt.Errorf("value decoding fail: %w", err)
				return
			}

			e.Value = uv
			t.lastSequence = e.Sequence

			outEvent <- e
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read fail: %w", err)
		}
	}()

	return outEvent, outError
}
