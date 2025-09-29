package transaction

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func TestCreateTransactor(t *testing.T) {
	const filename = "test_kv"
	defer os.Remove(filename)

	ctx := context.Background()
	tr, err := NewFileTransactor(ctx, filename)
	if err != nil {
		t.Fatalf("cannot create transactor: %v", err)
	}
	defer tr.Close()

	if tr == nil {
		t.Fatal("transactor is nil")
	}
	if !fileExists(filename) {
		t.Fatalf("file %s does not exist after creation", filename)
	}
}

func TestConcurrentWritesAndRead(t *testing.T) {
	const filename = "test_kv"
	defer os.Remove(filename)

	ctx := context.Background()

	transactor, err := NewFileTransactor(ctx, filename)
	if err != nil {
		t.Fatalf("failed to create transactor: %v", err)
	}

	defer func() {
		if err := transactor.Close(); err != nil {
			t.Fatalf("close error: %v", err)
		}
		os.Remove(transactor.file.Name())
	}()

	wg := &sync.WaitGroup{}
	const workers = 5

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			key := fmt.Sprintf("worker:%d", id)
			val := fmt.Sprintf("value-%d", id)

			if err := transactor.WritePut(ctx, key, val); err != nil {
				t.Errorf("worker %d error: %v", id, err)
			}
		}(i)
	}
	wg.Wait()

	transactor1, err := NewFileTransactor(ctx, filename)
	if err != nil {
		t.Fatalf("failed to create transactor1: %v", err)
	}

	defer func() {
		if err := transactor1.Close(); err != nil {
			t.Fatalf("close error: %v", err)
		}
		os.Remove(transactor.file.Name())
	}()

	var readEvents []Event
	eventsCh, errCh := transactor1.readEvents()
	for event := range eventsCh {
		readEvents = append(readEvents, event)
	}

	if err := <-errCh; err != nil {
		t.Fatalf("read error: %v", err)
	}
	if len(readEvents) != workers {
		t.Fatalf("expected %d rows, got %d", workers, len(readEvents))
	}
}

func TestSendClosedTransactor(t *testing.T) {
	const filename = "test_kv"
	defer os.Remove(filename)

	ctx := context.Background()
	tr, err := NewFileTransactor(ctx, filename)
	if err != nil {
		t.Fatalf("cannot create transactor: %v", err)
	}
	tr.Close()

	const key = "key"
	const value = "value"

	if err := tr.WritePut(ctx, key, value); !errors.Is(err, ErrTransactorClosed) {
		t.Fatal("transactor is not closed")
	}

	if err := tr.WriteDelete(ctx, key); !errors.Is(err, ErrTransactorClosed) {
		t.Fatal("transactor is not closed")
	}
}

func TestClosedTransactor(t *testing.T) {
	const filename = "test_kv"
	defer os.Remove(filename)

	ctx := context.Background()
	tr, err := NewFileTransactor(ctx, filename)
	if err != nil {
		t.Fatalf("cannot create transactor: %v", err)
	}

	if err := tr.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	if err := tr.Close(); err != ErrTransactorClosed {
		t.Fatalf("got: %v, but expected: %v", err, ErrTransactorClosed)
	}
}
