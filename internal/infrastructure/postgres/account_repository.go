package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"

	"simple_bank/internal/domain"
)

type AccountRepository struct {
	db DBTX
}

func NewAccountRepository(db DBTX) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Create(ctx context.Context, account *domain.Account) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO accounts (id, owner_name, balance, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		account.ID(),
		account.OwnerName(),
		int64(account.Balance()),
		account.CreatedAt(),
		account.UpdatedAt(),
	)
	return err
}

func (r *AccountRepository) FindByID(ctx context.Context, id string) (*domain.Account, error) {
	return r.findOne(ctx,
		`SELECT id, owner_name, balance, created_at, updated_at
		 FROM accounts WHERE id = $1`, id)
}

func (r *AccountRepository) FindByIDForUpdate(ctx context.Context, id string) (*domain.Account, error) {
	return r.findOne(ctx,
		`SELECT id, owner_name, balance, created_at, updated_at
		 FROM accounts WHERE id = $1 FOR UPDATE`, id)
}

func (r *AccountRepository) Update(ctx context.Context, account *domain.Account) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE accounts SET balance = $1, updated_at = $2 WHERE id = $3`,
		int64(account.Balance()),
		account.UpdatedAt(),
		account.ID(),
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrAccountNotFound
	}
	return nil
}

func (r *AccountRepository) findOne(ctx context.Context, query, id string) (*domain.Account, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, domain.ErrAccountNotFound
	}

	var (
		accountID string
		ownerName string
		balance   int64
		createdAt time.Time
		updatedAt time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&accountID, &ownerName, &balance, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrAccountNotFound
	}
	if err != nil {
		return nil, err
	}

	return domain.RestoreAccount(accountID, ownerName, domain.Money(balance), createdAt, updatedAt), nil
}
