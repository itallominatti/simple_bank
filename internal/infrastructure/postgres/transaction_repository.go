package postgres

import (
	"context"

	"simple_bank/internal/domain"
)

type TransactionRepository struct {
	db DBTX
}

func NewTransactionRepository(db DBTX) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(ctx context.Context, t *domain.Transaction) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO transactions (id, account_id, type, amount, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		t.ID,
		t.AccountID,
		string(t.Type),
		int64(t.Amount),
		t.CreatedAt,
	)
	return err
}

func (r *TransactionRepository) ListByAccountID(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, account_id, type, amount, created_at
		 FROM transactions
		 WHERE account_id = $1
		 ORDER BY created_at DESC`, accountID)
	if err != nil {
		return nil, err
	}
	// "defer" agenda algo para rodar quando a função terminar.
	// Aqui garantimos que o resultado da consulta seja sempre fechado.
	defer rows.Close()

	transactions := make([]*domain.Transaction, 0)
	for rows.Next() {
		var (
			t      domain.Transaction
			txType string
			amount int64
		)
		if err := rows.Scan(&t.ID, &t.AccountID, &txType, &amount, &t.CreatedAt); err != nil {
			return nil, err
		}
		t.Type = domain.TransactionType(txType)
		t.Amount = domain.Money(amount)
		transactions = append(transactions, &t)
	}
	return transactions, rows.Err()
}
