package storage

import (
	"errors"
	"time"

	"github.com/glynternet/go-accounting/account"
	gtime "github.com/glynternet/go-time"
)

// Account holds logic for an Account item that is held within a Storage
type Account struct {
	ID uint
	account.Account
	deletedAt gtime.NullTime
}

func DeletedAt(t time.Time) func(*Account) error {
	return func(a *Account) error {
		a.deletedAt = gtime.NullTime{Valid: true, Time: t}
		return nil
	}
}

// Accounts holds multiple Account items.
type Accounts []Account

// Equal return true if two Accounts are identical.
func (a Account) Equal(b Account) (bool, error) {
	if a.ID != b.ID {
		return false, nil
	}
	if a.Account == nil || b.Account == nil {
		return false, errors.New("nil account.Account")
	}
	if !a.Account.Equal(b.Account) {
		return false, nil
	}
	if !a.deletedAt.Equal(b.deletedAt) {
		return false, errors.New("accounts are equal but one has been deleted")
	}
	return true, nil
}
