package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-money/currency"
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

// MarshalJSON marshals an Account into a json blob, returning the blob with any errors that occur during the marshalling.
func (a Account) MarshalJSON() ([]byte, error) {
	_, err := json.Marshal(a.Account)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal inner account to json: %v", err)
	}
	type Alias Account
	return json.Marshal(&struct {
		*Alias
		DeletedAt gtime.NullTime
	}{
		Alias:     (*Alias)(&a),
		DeletedAt: a.deletedAt,
	})
}

// UnmarshalJSON attempts to unmarshal a json blob into an Account object, returning any errors that occur during the unmarshalling.
func (a *Account) UnmarshalJSON(data []byte) (err error) {
	aux := new(struct {
		ID      uint
		Account struct {
			Name     string
			Opened   time.Time
			Closed   gtime.NullTime
			Currency string
		}
		DeletedAt gtime.NullTime
	})
	err = json.Unmarshal(data, &aux)
	if err != nil {
		return fmt.Errorf("error unmarshalling into auxilliary account struct: %v", err)
	}
	a.ID = aux.ID
	a.deletedAt = aux.DeletedAt
	c, err := currency.NewCode(aux.Account.Currency)
	if err != nil {
		return fmt.Errorf("error creating new currency code from auxilliary account struct: %v", err)
	}
	var o account.Option
	if aux.Account.Closed.Valid {
		o = account.CloseTime(aux.Account.Closed.Time)
	}
	inner, err := account.New(aux.Account.Name, *c, aux.Account.Opened, o)
	if err != nil {
		return fmt.Errorf("error creating inner account: %v", err)
	}
	a.Account = inner
	return
}
