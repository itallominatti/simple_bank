package application

import (
	"context"
	"simple_bank/internal/domain"
)

type Repositories struct {
	Accounts     domain.AccountRepository
	Transactions domain.TransactionRepository
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(repos Repositories) error) error
}
