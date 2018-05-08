package test

import (
	"testing"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storagetest"
	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/common"
	"github.com/stretchr/testify/assert"
)

const numOfAccounts = 2

// Test will run a suite of tests again a given Storage
func Test(t *testing.T, store storage.Storage) {
	tests := []struct {
		title string
		run   func(t *testing.T, c storage.Storage)
	}{
		{
			title: "inserting and retrieving accounts",
			run:   insertAndRetrieveAccounts,
		},
		{
			title: "inserting balances",
			run:   insertAndRetrieveBalances,
		},
		{
			title: "update account",
			run:   updateAccounts,
		},
	}
	for _, test := range tests {
		success := t.Run(test.title, func(t *testing.T) {
			test.run(t, store)
		})
		if !success {
			t.Fail()
			return
		}
	}
}

func insertAndRetrieveAccounts(t *testing.T, store storage.Storage) {
	as, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts")

	if !assert.Len(t, *as, 0) {
		t.FailNow()
	}

	a := accountingtest.NewAccount(t, "A", accountingtest.NewCurrencyCode(t, "YEN"), time.Now())
	insertedA, err := store.InsertAccount(*a)
	common.FatalIfError(t, err, "inserting account")

	as = selectAccounts(t, store)

	if !assert.Len(t, *as, 1) {
		t.FailNow()
	}
	selectedA := &(*as)[0]
	equal, err := insertedA.Equal(*selectedA)
	common.FatalIfError(t, err, "equaling inserted and retrieved")
	if !assert.True(t, equal) {
		t.FailNow()
	}

	selectedByIDA, err := store.SelectAccount(insertedA.ID)
	common.FatalIfError(t, err, "selecting account by ID")

	assertThreeAccountsEqual(t, insertedA, selectedA, selectedByIDA)

	b := accountingtest.NewAccount(t,
		"B",
		accountingtest.NewCurrencyCode(t, "EUR"),
		time.Now().Add(-1*time.Hour),
	)

	insertedB, err := store.InsertAccount(*b)
	common.FatalIfError(t, err, "inserting account")

	as, err = store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts after inserting two")

	if !assert.Len(t, *as, numOfAccounts) {
		t.FailNow()
	}
	selectedB := &(*as)[1]
	equal, err = insertedB.Equal(*selectedB)
	common.FatalIfError(t, err, "equaling inserted and retrieved")
	if !assert.True(t, equal) {
		t.FailNow()
	}

	selectedByIDB, err := store.SelectAccount(insertedB.ID)
	common.FatalIfError(t, err, "selecting account by ID")

	assertThreeAccountsEqual(t, insertedB, selectedB, selectedByIDB)

	equal, err = insertedA.Equal(*insertedB)
	common.FatalIfError(t, err, "equaling insertedA and insertedB")
	if !assert.False(t, equal) {
		t.FailNow()
	}

	equal, err = selectedA.Equal(*selectedB)
	common.FatalIfError(t, err, "equaling selectedA and selectedB")
	if !assert.False(t, equal) {
		t.FailNow()
	}
}

// assertThreeAccountsEqual will compare three accounts against each other and fail the
// test if any of them are not equal against each other.
func assertThreeAccountsEqual(t *testing.T, a, b, c *storage.Account) {
	for _, as := range []struct {
		A, B *storage.Account
	}{
		{
			A: a, B: b,
		},
		{
			A: a, B: c,
		},
		{
			A: b, B: c,
		},
	} {
		equal, err := as.A.Equal(*as.B)
		common.FatalIfErrorf(t, err, "equalling accounts %+v", as)
		if !equal {
			t.Fatal("not equal")
		}
	}
}

func insertAndRetrieveBalances(t *testing.T, store storage.Storage) {
	as := selectAccounts(t, store)
	assert.Len(t, *as, numOfAccounts)

	type accountBalances struct {
		storage.Account
		storage.Balances
	}

	abs := make([]accountBalances, numOfAccounts)
	for i, a := range *as {
		bs, err := store.SelectAccountBalances(a)
		common.FatalIfError(t, err, "selecting account balances")
		assert.Len(t, *bs, 0)
		abs[i] = accountBalances{
			Account:  (*as)[i],
			Balances: *bs,
		}
	}

	for i := 0; i < numOfAccounts; i++ {
		b := newTestBalance(t, abs[i].Account.Account.Opened())
		inserted, err := store.InsertBalance(abs[i].Account, b)
		common.FatalIfError(t, err, "inserting Balance")
		equal := b.Equal(inserted.Balance)
		if !assert.True(t, equal) {
			t.FailNow()
		}

		bs, err := store.SelectAccountBalances(abs[i].Account)
		common.FatalIfError(t, err, "selecting account balances")
		assert.Len(t, *bs, 1)
		abs[i].Balances = *bs

		invalidBalance, err := balance.New(abs[i].Account.Account.Opened().Add(-time.Second))
		common.FatalIfError(t, err, "creating new invalid Balance")
		inserted, err = store.InsertBalance(abs[i].Account, *invalidBalance)
		if !assert.Error(t, err, "inserting Balance") {
			t.FailNow()
		}
		assert.Nil(t, inserted)
	}
}

func updateAccounts(t *testing.T, store storage.Storage) {
	initial := accountingtest.NewAccount(t, "A", accountingtest.NewCurrencyCode(t, "YEN"), time.Now())

	t.Run("valid without balances", func(t *testing.T) {
		inserted, err := store.InsertAccount(*initial)
		common.FatalIfError(t, err, "inserting account to store")

		// Here we truncate to the closest second to avoid the issue where
		// postgres stores times down to only the closest millisecond or so
		// TODO: Sort out rounding of times logic and document it properly. For
		// TODO: the moment, it is assumed that accounts won't be updated and
		// TODO: then compared against their original down to such a fine grain
		updates := accountingtest.NewAccount(t,
			"B",
			accountingtest.NewCurrencyCode(t, "GBP"),
			time.Now().Truncate(time.Second),
			account.CloseTime(time.Now().Add(24*time.Hour).Truncate(time.Second)),
		)

		updatedA, err := store.UpdateAccount(inserted, updates)
		common.FatalIfError(t, err, "updating account")
		assert.Equal(t, updatedA.ID, inserted.ID)
		assert.True(t,
			updatedA.Account.Equal(*updates),
			"inserted: %+v\nupdates: %+v\nupdatedA: %+v", inserted.Account, updates, updatedA.Account,
		)
	})

	t.Run("valid with balances", func(t *testing.T) {
		inserted, err := store.InsertAccount(*initial)
		common.FatalIfError(t, err, "inserting account to store")

		for _, b := range newTestBalances(t, 10, inserted.Account.Opened(), time.Hour) {
			_, err := store.InsertBalance(*inserted, b)
			common.FatalIfError(t, err, "inserting balance")
		}

		// Here we truncate to the closest second to avoid the issue where
		// postgres stores times down to only the closest millisecond or so
		// TODO: Sort out rounding of times logic and document it properly. For
		// TODO: the moment, it is assumed that accounts won't be updated and
		// TODO: then compared against their original down to such a fine grain
		updates := accountingtest.NewAccount(t,
			"B",
			accountingtest.NewCurrencyCode(t, "GBP"),
			time.Now().Add(-time.Hour).Truncate(time.Second),
			account.CloseTime(time.Now().Add(200*time.Hour).Truncate(time.Second)),
		)

		updatedA, err := store.UpdateAccount(inserted, updates)
		common.FatalIfError(t, err, "updating account")
		assert.Equal(t, updatedA.ID, inserted.ID)
		assert.True(t,
			updatedA.Account.Equal(*updates),
			"inserted: %+v\nupdates: %+v\nupdatedA: %+v",
			inserted.Account, updates, updatedA.Account,
		)
	})

	t.Run("invalid with balances", func(t *testing.T) {
		inserted, err := store.InsertAccount(*initial)
		common.FatalIfError(t, err, "inserting account to store")

		for _, b := range newTestBalances(t, 10, inserted.Account.Opened(), time.Hour) {
			_, err := store.InsertBalance(*inserted, b)
			common.FatalIfError(t, err, "inserting balance")
		}

		// Here we truncate to the closest second to avoid the issue where
		// postgres stores times down to only the closest millisecond or so
		// TODO: Sort out rounding of times logic and document it properly. For
		// TODO: the moment, it is assumed that accounts won't be updated and
		// TODO: then compared against their original down to such a fine grain
		updates := accountingtest.NewAccount(t,
			"B",
			accountingtest.NewCurrencyCode(t, "GBP"),
			time.Now().Add(-time.Hour).Truncate(time.Second),
			account.CloseTime(time.Now().Add(time.Hour).Truncate(time.Second)),
		)

		updatedA, err := store.UpdateAccount(inserted, updates)
		assert.Error(t, err)
		assert.Nil(t, updatedA)
	})
}

func selectAccounts(t *testing.T, store storage.Storage) *storage.Accounts {
	as, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts after inserting one")
	return as
}

func newTestBalance(t *testing.T, time time.Time, os ...balance.Option) balance.Balance {
	b, err := balance.New(time, os...)
	common.FatalIfError(t, err, "creating test balance")
	return *b
}

func newTestBalances(
	t *testing.T, count int, start time.Time, interval time.Duration, os ...balance.Option,
) []balance.Balance {
	bs := make([]balance.Balance, count)
	for i := 0; i < count; i++ {
		bs[i] = newTestBalance(t, start.Add(time.Duration(i)*interval), os...)
	}
	return bs
}
