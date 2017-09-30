package GOHMoneyDB

import (
	"errors"
	"fmt"
)

// NoAccountWithIDError is an error returned when no account with a given ID can be found within a DB
type NoAccountWithIDError uint

// Error ensures that NoAccountWithIDError adheres to the error interface
func (e NoAccountWithIDError) Error() string {
	return fmt.Sprintf("No account with ID: %d", uint(e))
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
	AccountID, BalanceID uint
}

// Describes InvalidAccountBalanceError to ensure that InvalidAccountBalanceError adheres to the error interface.
func (e InvalidAccountBalanceError) Error() string {
	return fmt.Sprintf(`Invalid balance (id: %d) for account (id: %d).`, e.BalanceID, e.AccountID)
}

// ErrAccountDeleted is the error that is returned what a deleted account is Validated.
var ErrAccountDeleted = errors.New("account is deleted")

// ErrAccountDifferentInDbAndRuntime is the error that is returned when a method is called on an account that doesn't match the record in the DB.
var ErrAccountDifferentInDbAndRuntime = errors.New("account in DB different to Account in runtime")
