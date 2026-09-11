package domain

import "context"

type AccountRepository interface {
	Create(ctx context.Context, account *Account) error
	FindByID(ctx context.Context, id string) (*Account, error)
	FindByIDForUpdate(ctx context.Context, id string) (*Account, error)
	Update(ctx context.Context, account *Account) error
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *Transaction) error
	ListByAccountID(ctx context.Context, accountID string) ([]*Transaction, error)
}
