// +build integration

package postgres

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-money/common"
	"github.com/glynternet/go-money/currency"
	"github.com/stretchr/testify/assert"
)

func Test_SelectAccounts(t *testing.T) {
	deleteTestDBIgnorantly(t)
	store := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(store)
	accounts, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts")
	if !assert.NotNil(t, accounts) {
		t.FailNow()
	}
	assert.Len(t, *accounts, 0)
	checkAccountsSortedByIDAscending(t, *accounts)
}

func TestPostgres_InsertAccount(t *testing.T) {
	deleteTestDBIgnorantly(t)
	store := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(store)
	numOfAccounts := 10
	as := newTestAccounts(t, numOfAccounts)
	for _, a := range as {
		dba, err := store.InsertAccount(a)
		common.FatalIfError(t, err, "inserting account")
		assert.Equal(t, a.Name(), dba.Account.Name())
		assert.Equal(t, a.CurrencyCode(), dba.Account.CurrencyCode())
		assert.True(t, a.Opened().Sub(dba.Account.Opened()) < time.Millisecond)
		assert.Equal(t, a.Closed().Valid, dba.Account.Closed().Valid)
		if a.Closed().Valid {
			assert.True(t, a.Closed().Time.Sub(dba.Account.Closed().Time) < time.Millisecond)
		}
	}
	accounts, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts")
	if !assert.NotNil(t, accounts) {
		t.FailNow()
	}
	assert.Len(t, *accounts, numOfAccounts)
}

func Test_InsertAccount_SelectAccount(t *testing.T) {
	deleteTestDBIgnorantly(t)
	store := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(store)
	asOpen := newTestAccounts(t, 5)
	asClosed := newTestAccounts(t, 5, account.CloseTime(time.Now().Add(time.Hour)))
	var dbas []*storage.Account
	for _, a := range append(asOpen, asClosed...) {
		inserted, err := store.InsertAccount(a)
		common.FatalIfError(t, err, "inserting account")
		dbas = append(dbas, inserted)
		selected, err := store.SelectAccount(inserted.ID)
		assert.NoError(t, err)
		assert.Equal(t, inserted, selected)
		equal, err := inserted.Equal(*selected)
		assert.NoError(t, err)
		assert.True(t, equal)
	}
}

func checkAccountsSortedByIDAscending(t *testing.T, accounts storage.Accounts) {
	for i := 0; i+1 < len(accounts); i++ {
		account := accounts[i]
		nextAccount := accounts[i+1]
		switch {
		case account.ID >= nextAccount.ID:
			var message bytes.Buffer
			fmt.Fprintf(&message, "Accounts not returned sorted by ID. ID %d appears before %d.\n", account.ID, nextAccount.ID)
			fmt.Fprintf(&message, "accounts[%d]: %v", i, account)
			fmt.Fprintf(&message, "accounts[%d]: %v", i+1, nextAccount)
			t.Errorf(message.String())
		}
	}
}

func newTestAccount(t *testing.T, options ...account.Option) account.Account {
	c, err := currency.NewCode("EUR")
	common.FatalIfError(t, err, "creating currency code")
	a, err := account.New("TEST ACCOUNT", *c, time.Now(), options...)
	common.FatalIfError(t, err, "creating account")
	return *a
}

func newTestDBAccountOpen(t *testing.T, s storage.Storage) storage.Account {
	a := newTestAccount(t)
	dba, err := s.InsertAccount(a)
	common.FatalIfError(t, err, "inserting account for testing")
	return *dba
}

// newTestAccounts creates a given number of account.Account.
// Each account will have a different Currency Code, its index formatted as C%02D.
// To create only accounts with the same index, use another function.
// TODO: extend this to take currencyCount and perCurrencyCount? Where a given number of accounts per currency are generated, and a given number of currencies are generated.
func newTestAccounts(t *testing.T, count int, options ...account.Option) []account.Account {
	as := make([]account.Account, count)
	for i := 0; i < count; i++ {
		c, err := currency.NewCode(fmt.Sprintf("C%02d", i))
		common.FatalIfError(t, err, "creating currency code")
		name := fmt.Sprintf("TEST ACCOUNT %02d", i)
		a, err := account.New(name, *c, time.Now(), options...)
		common.FatalIfError(t, err, "creating account")
		as[i] = *a
	}
	return as
}

// newTestInsertedStorageAccounts is the same as newTestAccounts except that
// it inserts each account into the given storage
func newTestInsertedStorageAccounts(t *testing.T, s storage.Storage, count int, options ...account.Option) []*storage.Account {
	as := newTestAccounts(t, count, options...)
	dbas := make([]*storage.Account, count)
	for i := 0; i < count; i++ {
		dba, err := s.InsertAccount(as[i])
		common.FatalIfError(t, err, "inserting account")
		dbas[i] = dba
	}
	return dbas
}
