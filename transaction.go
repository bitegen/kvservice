package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type EventType byte

const (
	EventDelete EventType = iota + 1
	EventPut
)

type Event struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}

type FileTransactionLogger struct {
	events       chan<- Event
	errors       chan error
	lastSequence uint64
	file         *os.File
	wg           *sync.WaitGroup
}

func NewFileTransactionLogger(filename string) (*FileTransactionLogger, error) {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("cannot create directory: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)
	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	logger := &FileTransactionLogger{
		file: file,
		wg:   &sync.WaitGroup{},
	}
	logger.Run()

	return logger, nil
}

func (f *FileTransactionLogger) WritePut(key string, value string) {
	f.wg.Add(1)
	f.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (f *FileTransactionLogger) WriteDelete(key string) {
	f.wg.Add(1)
	f.events <- Event{EventType: EventDelete, Key: key}
}

func (f *FileTransactionLogger) Err() <-chan error {
	return f.errors
}

func (f *FileTransactionLogger) Close() error {
	f.Wait()

	if f.events != nil {
		close(f.events)
		f.events = nil
	}

	return f.file.Close()
}

func (f *FileTransactionLogger) Wait() {
	f.wg.Wait()
}

func (f *FileTransactionLogger) Run() {
	events := make(chan Event, 16)
	f.events = events

	errors := make(chan error, 1)
	f.errors = errors

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()

		for event := range events {
			f.lastSequence++

			journalRow := fmt.Sprintf(
				"%d\t%d\t%s\t%s\n",
				f.lastSequence, event.EventType, event.Key, event.Value)

			_, err := f.file.WriteString(journalRow)
			if err != nil {
				f.errors <- err
				return
			}

			f.wg.Done() // do for each event from WriteDelete/WritePut
		}
	}()
}

func (f *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(f.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	f.wg.Add(1)
	go func() {
		defer func() {
			close(outError)
			close(outEvent)
			f.wg.Done()
		}()

		var event Event

		for scanner.Scan() {
			line := scanner.Text()

			if _, err := fmt.Sscanf(line,
				"%d\t%d\t%s\t%s\n",
				&event.Sequence, &event.EventType, &event.Key, &event.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}

			if f.lastSequence >= event.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			f.lastSequence = event.Sequence

			outEvent <- event
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}
