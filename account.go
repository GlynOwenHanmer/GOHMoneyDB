package storage

import (
	"github.com/glynternet/go-accounting/account"
)

// Account holds logic for an Account item that is held within a Storage
type Account struct {
	ID uint
	account.Account
	deletedAt time.NullTime
}

// Accounts holds multiple Account items.
type Accounts []Account

// Equal return true if two Accounts are identical.
func (a Account) Equal(b Account) (bool, error) {
	if a.ID != b.ID {
		return false, nil
	}
	if !a.Account.Equal(b.Account) {
		return false, nil
	}
	if !a.deletedAt.Equal(b.deletedAt) {
		return false, errors.New("accounts are equal but one has been deleted")
	}
	return true, nil
