package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"simple_bank/internal/application"
)

type UnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Do(ctx context.Context, fn func(repos application.Repositories) error) error {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("Iniciar transação: %w", err)
	}

	defer func() { _ = tx.Rollback() }()

	repos := application.Repositories{
		Accounts:     NewAccountRepository(tx),
		Transactions: NewTransactionRepository(tx),
	}

	if err := fn(repos); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Confirmar transação: %w", err)
	}
	return nil
}
