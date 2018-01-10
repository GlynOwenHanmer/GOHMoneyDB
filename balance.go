package storage

import (
	"github.com/glynternet/go-accounting/balance"
)

// Balance holds logic for an Account item that is held within a go-money database.
type Balance struct {
	balance.Balance
	ID uint
}

// Equal returns true if two Balance items are logically identical
func (b Balance) Equal(ob Balance) bool {
	if b.ID != ob.ID || !b.Balance.Equal(ob.Balance) {
		return false
	}
	return true
}

// Balances holds multiple Balance items
type Balances []Balance
