package application

import (
	"context"
	"simple_bank/internal/domain"
)

type AccountService struct {
	accounts     domain.AccountRepository
	transactions domain.TransactionRepository
	uow          UnitOfWork
}

func NewAccountService(
	accounts domain.AccountRepository,
	transactions domain.TransactionRepository,
	uow UnitOfWork,
) *AccountService {
	return &AccountService{
		accounts:     accounts,
		transactions: transactions,
		uow:          uow,
	}
}

func (s *AccountService) CreateAccount(ctx context.Context, ownerName string) (*domain.Account, error) {
	account, err := domain.NewAccount(ownerName)
	if err != nil {
		return nil, err
	}

	if err := s.accounts.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) GetAccount(ctx context.Context, id string) (*domain.Account, error) {
	return s.accounts.FindByID(ctx, id)
}

func (s *AccountService) Deposit(ctx context.Context, accountID string, amount domain.Money) (*domain.Account, error) {
	var account *domain.Account

	err := s.uow.Do(ctx, func(repos Repositories) error {
		var err error
		account, err = repos.Accounts.FindByIDForUpdate(ctx, accountID)
		if err != nil {
			return err
		}

		if err := account.Deposit(amount); err != nil {
			return err
		}

		if err := repos.Accounts.Update(ctx, account); err != nil {
			return err
		}

		return repos.Transactions.Create(ctx,
			domain.NewTransaction(account.ID(), domain.TransactionDeposit, amount))
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) Withdraw(ctx context.Context, accountID string, amount domain.Money) (*domain.Account, error) {
	var account *domain.Account

	err := s.uow.Do(ctx, func(repos Repositories) error {
		var err error
		account, err = repos.Accounts.FindByIDForUpdate(ctx, accountID)
		if err != nil {
			return err
		}

		if err := account.Withdraw(amount); err != nil {
			return err
		}

		if err := repos.Accounts.Update(ctx, account); err != nil {
			return err
		}

		return repos.Transactions.Create(ctx,
			domain.NewTransaction(account.ID(), domain.TransactionWithdraw, amount))
	})
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (s *AccountService) Transfer(ctx context.Context, fromID, toID string, amount domain.Money) error {
	if fromID == toID {
		return domain.ErrSameAccountTransfer
	}

	if !amount.IsPositive() {
		return domain.ErrInvalidAmount
	}

	return s.uow.Do(ctx, func(repos Repositories) error {
		firstID, secondID := fromID, toID
		if secondID < firstID {
			firstID, secondID = secondID, firstID
		}

		first, err := repos.Accounts.FindByIDForUpdate(ctx, firstID)
		if err != nil {
			return err
		}

		second, err := repos.Accounts.FindByIDForUpdate(ctx, secondID)
		if err != nil {
			return err
		}

		from, to := first, second
		if from.ID() != fromID {
			from, to = second, first
		}

		if err := domain.Transfer(from, to, amount); err != nil {
			return err
		}

		if err := repos.Accounts.Update(ctx, from); err != nil {
			return err
		}

		if err := repos.Accounts.Update(ctx, to); err != nil {
			return err
		}

		if err := repos.Transactions.Create(ctx,
			domain.NewTransaction(from.ID(), domain.TransactionTransferOut, amount)); err != nil {
			return err
		}
		return repos.Transactions.Create(ctx,
			domain.NewTransaction(to.ID(), domain.TransactionTransferIn, amount))
	})
}

func (s *AccountService) ListTransactions(ctx context.Context, accountID string) ([]*domain.Transaction, error) {
	if _, err := s.accounts.FindByID(ctx, accountID); err != nil {
		return nil, err
	}
	return s.transactions.ListByAccountID(ctx, accountID)
}
