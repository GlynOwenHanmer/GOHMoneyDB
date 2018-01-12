package storage

import (
	"testing"

	"time"

	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/currency"
	gtime "github.com/glynternet/go-time"
	"github.com/stretchr/testify/assert"
)

type mockAccountAccount struct {
	equal bool
}

func (a mockAccountAccount) Name() (s string)                              { return }
func (a mockAccountAccount) Opened() (t time.Time)                         { return }
func (a mockAccountAccount) Closed() (nt gtime.NullTime)                   { return }
func (a mockAccountAccount) TimeRange() (r gtime.Range)                    { return }
func (a mockAccountAccount) IsOpen() (b bool)                              { return }
func (a mockAccountAccount) CurrencyCode() (c currency.Code)               { return }
func (a mockAccountAccount) ValidateBalance(b balance.Balance) (err error) { return }
func (a mockAccountAccount) Equal(b account.Account) bool                  { return a.equal }

func TestAccount_Equal(t *testing.T) {
	// if account a is true, account.Account.Equal will evaluate to true
	for _, test := range []struct {
		a, b                       Account
		name                       string
		equal, error, accountEqual bool
	}{
		{
			name:  "both nil account",
			error: true,
		},
		{
			name:         "unequal account.Account",
			a:            Account{Account: mockAccountAccount{equal: false}},
			b:            Account{Account: mockAccountAccount{}},
			accountEqual: false,
			equal:        false,
		},
		{
			name:         "unequal ID",
			a:            Account{ID: 1},
			b:            Account{ID: 2},
			accountEqual: true,
		},
		{
			name:         "unequal deletedAt",
			a:            Account{deletedAt: gtime.NullTime{Valid: false}},
			b:            Account{deletedAt: gtime.NullTime{Valid: false}},
			accountEqual: true,
			error:        true,
		},
		{
			name:         "equal",
			a:            Account{Account: mockAccountAccount{equal: true}, deletedAt: gtime.NullTime{Valid: true}},
			b:            Account{Account: mockAccountAccount{}, deletedAt: gtime.NullTime{Valid: true}},
			accountEqual: true,
			equal:        true,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			equal, err := test.a.Equal(test.b)
			if test.error {
				assert.Error(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, test.equal, equal)
		})
	}
}
