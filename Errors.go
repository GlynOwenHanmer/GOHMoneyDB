package GOHMoneyDB

import (
	"fmt"
	"bytes"
)

// NoAccountWithIdError is an error returned when no account with a given ID can be found within a DB
type NoAccountWithIdError uint

// Error ensures that NoAccountWithIdError adheres to the error interface
func (e NoAccountWithIdError) Error() string {
	return fmt.Sprintf("No account with id: %d", uint(e))
}

// NewAccountFieldError holds zero or more descriptions of things that are wrong with potential new Account items.
type NewAccountFieldError []string

// Error ensures that NewAccountFieldError adheres to the error interface.
func (e NewAccountFieldError) Error() string {
	var errorString bytes.Buffer
	errorString.WriteString("NewAccountFieldError: ")
	for i, field := range e {
		errorString.WriteString(field)
		if i < len(e) - 1 {
			errorString.WriteByte(' ')
		}
	}
	return string(errorString.String())
}

// Various error strings describing possible errors with potential new Account items.
const (
	EmptyNameError = "Empty name."
	ZeroDateOpenedError = "No opened date given."
	ZeroValidDateClosedError = "Closed date marked as valid but not set."
	DateClosedBeforeDateOpenedError = "Closed date is before opened date."
)

// NewBalanceFieldError is an error returned there is a fault with a field of a givem potential new Balance item.
type NewBalanceFieldError string

// A collection of possible NewBalanceFieldErrors
const (
	BalanceZeroDate = NewBalanceFieldError("Date of balance is zero.")
)

// Error ensures that NewBalanceFieldError adheres to the error interface.
func (e NewBalanceFieldError) Error() string {
	return string(e)
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
