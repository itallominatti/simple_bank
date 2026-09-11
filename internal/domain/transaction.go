package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionDeposit     TransactionType = "DEPOSIT"      // depósito
	TransactionWithdraw    TransactionType = "WITHDRAW"     // saque
	TransactionTransferIn  TransactionType = "TRANSFER_IN"  // transferência recebida
	TransactionTransferOut TransactionType = "TRANSFER_OUT" // transferência enviada
)

type Transaction struct {
	ID        string
	AccountID string
	Type      TransactionType
	Amount    Money
	CreatedAt time.Time
}

func NewTransaction(accountID string, txType TransactionType, amount Money) *Transaction {
	return &Transaction{
		ID:        uuid.NewString(),
		AccountID: accountID,
		Type:      txType,
		Amount:    amount,
		CreatedAt: time.Now().UTC(),
	}
}
