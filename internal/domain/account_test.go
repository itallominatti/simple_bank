package domain

import (
	"errors"
	"testing"
)

func TestNewAccount_RequiresOwnerName(t *testing.T) {
	_, err := NewAccount("  ")
	if !errors.Is(err, ErrInvalidOwnerName) {
		t.Fatalf("esperava ErrInvalidOwnerName, recebi %v", err)
	}
}

func TestDepositAndWithdraw(t *testing.T) {
	account, _ := NewAccount("Maria")

	if err := account.Deposit(1000); err != nil {
		t.Fatalf("deposito falhou: %v", err)
	}

	if err := account.Withdraw(300); err != nil {
		t.Fatalf("Saque falhou %v", err)
	}

	if account.Balance() != 700 {
		t.Fatalf("Saldo esperado 700, recebi %d", account.Balance())
	}
}

func TestWithdraw_InsufficientFunds(t *testing.T) {
	account, _ := NewAccount("João")
	err := account.Withdraw(100)
	if !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("esperava ErrInsufficientFunds, recebi %v", err)
	}
	if account.Balance() != 0 {
		t.Fatalf("o saldo não deveria ter mudado")
	}
}

func TestTransfer(t *testing.T) {
	from, _ := NewAccount("Ana")
	to, _ := NewAccount("Beto")
	_ = from.Deposit(5000)

	if err := Transfer(from, to, 2000); err != nil {
		t.Fatalf("transferência falhou: %v", err)
	}
	if from.Balance() != 3000 || to.Balance() != 2000 {
		t.Fatalf("saldos errados: origem=%d destino=%d", from.Balance(), to.Balance())
	}
}

func TestTransfer_SameAccount(t *testing.T) {
	account, _ := NewAccount("Ana")
	_ = account.Deposit(5000)

	err := Transfer(account, account, 100)
	if !errors.Is(err, ErrSameAccountTransfer) {
		t.Fatalf("esperava ErrSameAccountTransfer, recebi %v", err)
	}
}

func TestMoneyString(t *testing.T) {
	if got := Money(1050).String(); got != "R$ 10,50" {
		t.Fatalf("esperava R$ 10,50, recebi %s", got)
	}
}
