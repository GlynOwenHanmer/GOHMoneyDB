package storage

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/common"
	"github.com/glynternet/go-money/currency"
	gtime "github.com/glynternet/go-time"
	"github.com/stretchr/testify/assert"
)

type mockAccountAccount struct {
	equal bool
	json  string
}

func (a mockAccountAccount) Name() (s string)                              { return }
func (a mockAccountAccount) Opened() (t time.Time)                         { return }
func (a mockAccountAccount) Closed() (nt gtime.NullTime)                   { return }
func (a mockAccountAccount) TimeRange() (r gtime.Range)                    { return }
func (a mockAccountAccount) IsOpen() (b bool)                              { return }
func (a mockAccountAccount) CurrencyCode() (c currency.Code)               { return }
func (a mockAccountAccount) ValidateBalance(b balance.Balance) (err error) { return }
func (a mockAccountAccount) Equal(b account.Account) bool                  { return a.equal }
func (a *mockAccountAccount) MarshalJSON() ([]byte, error)                 { return []byte(a.json), nil }

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

	t.Run("unequal deletedAt", func(t *testing.T) {
		a := Account{Account: mockAccountAccount{equal: true}}
		b := Account{Account: mockAccountAccount{equal: true}, deletedAt: gtime.NullTime{Valid: true}}
		var equal bool
		equal, err := a.Equal(b)
		assert.False(t, equal)
		assert.Error(t, err, "accounts are equal but one has been deleted")
	})
}

func TestDeletedAt(t *testing.T) {
	var a Account
	assert.Equal(t, gtime.NullTime{}, a.deletedAt)

	time := time.Date(1000, 0, 0, 0, 0, 0, 0, time.UTC)
	err := DeletedAt(time)(&a)
	assert.Nil(t, err)
	assert.Equal(t, gtime.NullTime{Valid: true, Time: time}, a.deletedAt)
}

func TestAccount_JSONLoop(t *testing.T) {
	c, err := currency.NewCode("NEO")
	common.FatalIfErrorf(t, err, "creating currency code")

	for _, test := range []struct {
		name  string
		code  currency.Code
		open  time.Time
		close gtime.NullTime
	}{
		{
			name: "with close time",
			open: time.Date(1999, 0, 0, 0, 0, 0, 0, time.UTC),
			close: gtime.NullTime{
				Valid: true,
				Time:  time.Date(2000, 0, 0, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "without close time",
			open: time.Date(1999, 0, 0, 0, 0, 0, 0, time.UTC),
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			inner, err := account.New(
				test.name,
				*c,
				test.open,
			)
			common.FatalIfError(t, err, "creating inner account")
			if test.close.Valid {
				err := account.CloseTime(test.close.Time)(inner)
				common.FatalIfError(t, err, "creating close time")
			}
			a := Account{
				ID: 147827,
				deletedAt: gtime.NullTime{
					Valid: true,
					Time:  time.Date(1000, 0, 0, 0, 0, 0, 0, time.UTC),
				},
				Account: inner,
			}
			bs, err := json.Marshal(a)
			common.FatalIfError(t, err, "marshalling json")
			var actual Account
			err = json.Unmarshal(bs, &actual)
			common.FatalIfError(t, err, "unmarshalling json")
			assert.Equal(t, a, actual, "intermediate stage: %s", bs)
		})
	}
}
