package GOHMoneyDB

import (
	"errors"
	"fmt"
)

// NoAccountWithIdError is an error returned when no account with a given ID can be found within a DB
type NoAccountWithIdError uint

// Error ensures that NoAccountWithIdError adheres to the error interface
func (e NoAccountWithIdError) Error() string {
	return fmt.Sprintf("No account with Id: %d", uint(e))
}

// BalancesError is an error type that can be returned when no Balance items are returned but there would have been Balance items expected to have returned.
type BalancesError string

// Error ensures that BalancesError adheres to the error interface
func (e BalancesError) Error() string {
	return string(e)
}

// A collection of possible BalancesErrors
const (
	NoBalances = BalancesError("No balances exist.")
)

// InvalidAccountBalanceError is an error type used to describe when a mistmatch of logical Account and Balance occurs.
type InvalidAccountBalanceError struct {
	AccountId, BalanceId uint
}

// Describes InvalidAccountBalanceError to ensure that InvalidAccountBalanceError adheres to the error interface.
func (e InvalidAccountBalanceError) Error() string {
	return fmt.Sprintf(`Invalid balance (id: %d) for account (id: %d).`, e.BalanceId, e.AccountId)
}

var AccountDeleted = errors.New("account is deleted")
var AccountDifferentInDbAndRuntime = errors.New("account in DB different to Account in runtime")
