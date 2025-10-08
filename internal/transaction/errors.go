package transaction

import "errors"

const (
	filename = "transactor.journal"
)

var (
	ErrTransactorClosed = errors.New("file transactor is closed")
	ErrOutOfSequence    = errors.New("transaction numbers out of sequence")
	ErrEmptyJournal     = errors.New("empty journal")
)
