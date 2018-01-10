package postgres2

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
	store := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(store)
	accounts, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts")
	if !assert.NotNil(t, accounts) {
		t.FailNow()
	}
	assert.Len(t, *accounts, 0)
	checkAccountsSortedByIdAscending(*accounts, t)
}

func Test_CreateAccount(t *testing.T) {
	store := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(store)
	numOfAccounts := 10
	as := newTestAccounts(t, numOfAccounts)
	for _, a := range as {
		dba, err := store.InsertAccount(a)
		common.FatalIfError(t, err, "inserting account")
		assert.Equal(t, a, dba.Account)
	}
	accounts, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts")
	if !assert.NotNil(t, accounts) {
		t.FailNow()
	}
	assert.Len(t, *accounts, numOfAccounts)
}

func checkAccountsSortedByIdAscending(accounts storage.Accounts, t *testing.T) {
	for i := 0; i+1 < len(accounts); i++ {
		account := accounts[i]
		nextAccount := accounts[i+1]
		switch {
		case account.ID > nextAccount.ID:
			var message bytes.Buffer
			fmt.Fprintf(&message, "Accounts not returned sorted by ID. ID %d appears before %d.\n", account.ID, nextAccount.ID)
			fmt.Fprintf(&message, "accounts[%d]: %s", i, account)
			fmt.Fprintf(&message, "accounts[%d]: %s", i+1, nextAccount)
			t.Errorf(message.String())
		}
	}
}

func newTestAccountOpen(t *testing.T) account.Account {
	c, err := currency.NewCode("EUR")
	common.FatalIfError(t, err, "creating currency code")
	a, err := account.New("TEST ACCOUNT", *c, time.Now())
	common.FatalIfError(t, err, "creating account")
	return *a
}

func newTestDBAccountOpen(t *testing.T, s storage.Storage) storage.Account {
	a := newTestAccountOpen(t)
	dba, err := s.InsertAccount(a)
	common.FatalIfError(t, err, "inserting account for testing")
	return *dba
}

// newTestAccounts creates an account with a random currency and random name
func newTestAccounts(t *testing.T, count int) []account.Account {
	as := make([]account.Account, count)
	for i := 0; i < count; i++ {
		c, err := currency.NewCode(fmt.Sprintf("C%02d", i))
		common.FatalIfError(t, err, "creating currency code")
		name := fmt.Sprintf("TEST ACCOUNT %02d", i)
		a, err := account.New(name, *c, time.Now())
		common.FatalIfError(t, err, "creating account")
		as[i] = *a
	}
	return as
}
