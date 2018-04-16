package test

import (
	"testing"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storagetest"
	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/common"
	"github.com/stretchr/testify/assert"
)

const numOfAccounts = 2

func Test(t *testing.T, store storage.Storage) {
	tests := []struct {
		title string
		run   func(t *testing.T, c storage.Storage)
	}{
		{
			title: "inserting and retrieving accounts",
			run:   InsertAndRetrieveAccounts,
		},
		{
			title: "inserting balances",
			run:   InsertAndRetrieveBalances,
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

func InsertAndRetrieveAccounts(t *testing.T, store storage.Storage) {
	as, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts")

	if !assert.Len(t, *as, 0) {
		t.FailNow()
	}

	a := accountingtest.NewAccount(t, "A", accountingtest.NewCurrencyCode(t, "YEN"), time.Now())
	insertedA, err := store.InsertAccount(a)
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

	for _, as := range []struct {
		A, B *storage.Account
	}{
		{
			A: insertedA, B: selectedA,
		},
		{
			A: insertedA, B: selectedByIDA,
		},
		{
			A: selectedA, B: selectedByIDA,
		},
	} {
		equal, err := as.A.Equal(*as.B)
		common.FatalIfErrorf(t, err, "equalling accounts %+v", as)
		if !assert.True(t, equal) {
			t.FailNow()
		}
	}

	b := accountingtest.NewAccount(t, "B", accountingtest.NewCurrencyCode(t, "EUR"), time.Now().Add(-1*time.Hour))

	insertedB, err := store.InsertAccount(b)
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

	for _, as := range []struct {
		A, B *storage.Account
	}{
		{
			A: insertedB, B: selectedB,
		},
		{
			A: insertedB, B: selectedByIDB,
		},
		{
			A: selectedB, B: selectedByIDB,
		},
	} {
		equal, err := as.A.Equal(*as.B)
		common.FatalIfErrorf(t, err, "equalling accounts %+v", as)
		if !assert.True(t, equal) {
			t.FailNow()
		}
	}

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

func InsertAndRetrieveBalances(t *testing.T, store storage.Storage) {
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
		b, err := balance.New(abs[i].Opened())
		common.FatalIfError(t, err, "creating new Balance")
		inserted, err := store.InsertBalance(abs[i].Account, *b)
		common.FatalIfError(t, err, "inserting Balance")
		equal := b.Equal(inserted.Balance)
		if !assert.True(t, equal) {
			t.FailNow()
		}

		bs, err := store.SelectAccountBalances(abs[i].Account)
		common.FatalIfError(t, err, "selecting account balances")
		assert.Len(t, *bs, 1)
		abs[i].Balances = *bs

		invalidBalance, err := balance.New(abs[i].Opened().Add(-time.Second))
		common.FatalIfError(t, err, "creating new invalid Balance")
		inserted, err = store.InsertBalance(abs[i].Account, *invalidBalance)
		if !assert.Error(t, err, "inserting Balance") {
			t.FailNow()
		}
		assert.Nil(t, inserted)
	}
}

func selectAccounts(t *testing.T, store storage.Storage) *storage.Accounts {
	as, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts after inserting one")
	return as
}
