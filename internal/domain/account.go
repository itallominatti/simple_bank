package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Account struct {
	id        string
	ownerName string
	balance   Money
	createdAt time.Time
	updatedAt time.Time
}

func NewAccount(ownerName string) (*Account, error) {
	ownerName = strings.TrimSpace(ownerName)
	if ownerName == "" {
		return nil, ErrInvalidOwnerName
	}

	now := time.Now().UTC()
	return &Account{
		id:        uuid.NewString(),
		ownerName: ownerName,
		balance:   0,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func RestoreAccount(id, ownerName string, balance Money, createdAt, updatedAt time.Time) *Account {
	return &Account{
		id:        id,
		ownerName: ownerName,
		balance:   balance,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (a *Account) Deposit(amount Money) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if a.balance < amount {
		return ErrInsufficientFunds
	}
	a.balance += amount
	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) Withdraw(amount Money) error {
	if !amount.IsPositive() {
		return ErrInvalidAmount
	}
	if a.balance < amount {
		return ErrInsufficientFunds
	}
	a.balance -= amount
	a.updatedAt = time.Now().UTC()
	return nil
}

func (a *Account) ID() string           { return a.id }
func (a *Account) OwnerName() string    { return a.ownerName }
func (a *Account) Balance() Money       { return a.balance }
func (a *Account) CreatedAt() time.Time { return a.createdAt }
func (a *Account) UpdatedAt() time.Time { return a.updatedAt }
