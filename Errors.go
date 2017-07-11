package GOHMoneyDB

import (
	"fmt"
)

// NoAccountWithIdError is an error returned when no account with a given ID can be found within a DB
type NoAccountWithIdError uint

// Error ensures that NoAccountWithIdError adheres to the error interface
func (e NoAccountWithIdError) Error() string {
	return fmt.Sprintf("No account with id: %d", uint(e))
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
