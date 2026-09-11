package domain

func Transfer(from, to *Account, amount Money) error {
	if from.ID() == to.ID() {
		return ErrSameAccountTransfer
	}

	if err := from.Withdraw(amount); err != nil {
		return err
	}

	return to.Deposit(amount)
}
