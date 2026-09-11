package handler

import (
	"time"

	"simple_bank/internal/domain"
)

type CreateAccountRequest struct {
	OwnerName string `json:"owner_name" binding:"required"`
}

type AmountRequest struct {
	Amount int64 `json:"amount"`
}

type TransferRequest struct {
	FromAccountID string `json:"from_account_id" binding:"required"`
	ToAccountID   string `json:"to_account_id" binding:"required"`
	Amount        int64  `json:"amount"`
}

type AccountResponse struct {
	ID               string    `json:"id"`
	OwnerName        string    `json:"owner_name"`
	Balance          int64     `json:"balance"`
	BalanceFormatted string    `json:"balance_formatted"`
	CreatedAt        time.Time `json:"created_at"`
}

type TransactionResponse struct {
	ID              string    `json:"id"`
	Type            string    `json:"type"`
	Amount          int64     `json:"amount"`
	AmountFormatted string    `json:"amount_formatted"`
	CreatedAt       time.Time `json:"created_at"`
}

func toAccountResponse(a *domain.Account) AccountResponse {
	return AccountResponse{
		ID:               a.ID(),
		OwnerName:        a.OwnerName(),
		Balance:          int64(a.Balance()),
		BalanceFormatted: a.Balance().String(),
		CreatedAt:        a.CreatedAt(),
	}
}

func toTransactionResponses(list []*domain.Transaction) []TransactionResponse {
	result := make([]TransactionResponse, 0, len(list))
	for _, t := range list {
		result = append(result, TransactionResponse{
			ID:              t.ID,
			Type:            string(t.Type),
			Amount:          int64(t.Amount),
			AmountFormatted: t.Amount.String(),
			CreatedAt:       t.CreatedAt,
		})
	}
	return result
}
