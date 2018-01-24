package storage

import (
	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-accounting/balance"
)

// Storage is something that can be used to store certain go-accounting types
type Storage interface {
	Available() bool
	Close() error
	InsertAccount(a account.Account) (*Account, error)
	SelectAccount(u uint) (*Account, error)
	SelectAccounts() (*Accounts, error)
	//UpdateAccount(a *Account, us account.Account) error
	//DeleteAccount(a *Account) error
	//
	InsertBalance(a Account, b balance.Balance) (*Balance, error)
	SelectAccountBalances(Account) (*Balances, error)
	//UpdateBalance(a Account, b *Balance, us balance.Balance) error
	//DeleteBalance(a Account, b *Balance) error
}

//type AccountQuery func (Storage, ...AccountFilter) (*Account, error)
