package functional

import (
	"log"
	"strings"
	"testing"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting-storage/postgres2"
	"github.com/glynternet/go-accounting-storagetest"
	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

const (
	// viper keys
	keyDBHost    = "db-host"
	keyDBUser    = "db-user"
	keyDBName    = "db-name"
	keyDBSSLMode = "db-sslmode"

	numOfAccounts = 2
)

func init() {
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv() // read in environment variables that match
}

func TestCreateStorage(t *testing.T) {
	const retries = 5

	var err error
	var i int
	for i = 1; i <= retries; i++ {
		err = postgres2.CreateStorage(
			viper.GetString(keyDBHost),
			viper.GetString(keyDBUser),
			viper.GetString(keyDBName),
			viper.GetString(keyDBSSLMode),
		)
		if err == nil {
			break
		}
		log.Printf("Attempt: %d, err: %v\n", i, err)
		time.Sleep(time.Second)
	}
	assert.NoError(t, err)
	t.Logf("Attempts: %d", i)
}

func TestInsertingAndRetrievingTwoAccounts(t *testing.T) {
	store := createStorage(t)

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

func TestInsertingBalances(t *testing.T) {
	store := createStorage(t)
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

func createStorage(t *testing.T) storage.Storage {
	cs, err := postgres2.NewConnectionString(
		viper.GetString(keyDBHost),
		viper.GetString(keyDBUser),
		viper.GetString(keyDBName),
		viper.GetString(keyDBSSLMode),
	)
	common.FatalIfError(t, err, "creating connection string")
	store, err := postgres2.New(cs)
	common.FatalIfError(t, err, "creating storage")
	if !assert.True(t, store.Available(), "store should be available") {
		t.FailNow()
	}
	return store
}

func selectAccounts(t *testing.T, store storage.Storage) *storage.Accounts {
	as, err := store.SelectAccounts()
	common.FatalIfError(t, err, "selecting accounts after inserting one")
	return as
}
