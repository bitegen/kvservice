package transaction

import (
	"cloud/internal/config"
	"context"
	"errors"
)

const (
	TransactorTypeInMemory = "in_memory_transactor"
	TransactorTypePostgres = "postgres_transactor"
)

type TransactorFactory struct {
	cfg *config.Config
}

func NewTransactorFactory(cfg *config.Config) *TransactorFactory {
	return &TransactorFactory{cfg: cfg}
}

func (f *TransactorFactory) Create(ctx context.Context, transactorType string) (Transactor, error) {
	switch transactorType {
	case TransactorTypeInMemory:
		return NewFileTransactor(ctx)
	case TransactorTypePostgres:
		return NewPostgresTransactor(ctx, f.cfg.Postgres)
	default:
		return nil, errors.New("unknown transactor type: " + transactorType)
	}
}
