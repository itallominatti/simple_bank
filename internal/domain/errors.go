package domain

import "errors"

var (
	ErrAccountNotFound     = errors.New("Conta não encontrada.")
	ErrInvalidOwnerName    = errors.New("O nome do titular é obrigatório")
	ErrInvalidAmount       = errors.New("O valor deve ser maior que zero")
	ErrInsufficientFunds   = errors.New("Saldo insuficiente")
	ErrSameAccountTransfer = errors.New("Não pode transferir para a mesma conta")
)
