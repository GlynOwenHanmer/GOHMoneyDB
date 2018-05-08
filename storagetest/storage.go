package storagetest

import (
	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-accounting/balance"
)

type Storage struct {
	IsAvailable bool
	Err         error

	*storage.Account
	AccountErr error

	*storage.Accounts

	*storage.Balance
	BalanceErr error

	*storage.Balances
	BalancesErr error
}

func (s *Storage) Available() bool { return s.IsAvailable }
func (s *Storage) Close() error    { return s.Err }

func (s *Storage) InsertAccount(a account.Account) (*storage.Account, error) {
	return s.Account, s.AccountErr
}

func (s *Storage) UpdateAccount(a *storage.Account, updates *account.Account) (*storage.Account, error) {
	return s.Account, s.AccountErr
}

func (s *Storage) SelectAccount(u uint) (*storage.Account, error) { return s.Account, s.AccountErr }
func (s *Storage) SelectAccounts() (*storage.Accounts, error)     { return s.Accounts, s.Err }

func (s *Storage) InsertBalance(a storage.Account, b balance.Balance) (*storage.Balance, error) {
	return s.Balance, s.BalanceErr
}
func (s *Storage) SelectAccountBalances(storage.Account) (*storage.Balances, error) {
	return s.Balances, s.BalancesErr
}
