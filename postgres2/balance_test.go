package postgres2

import (
	"testing"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/common"
	"github.com/stretchr/testify/assert"
)

func TestPostgres_InsertBalance_selectBalanceByID(t *testing.T) {
	s := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(s)

	a := newTestDBAccountOpen(t, s)

	for i, test := range []struct {
		time.Time
		int
		error bool
	}{
		{
			Time:  a.Opened().AddDate(-1, 0, 0),
			error: true,
		},
		{
			Time: a.Opened(),
			int:  -999,
		},
		{
			Time: a.Opened().AddDate(1, 0, 0),
			int:  0137,
		},
	} {
		b := newTestBalance(t, test.Time, balance.Amount(test.int))
		dbb, err := s.InsertBalance(a, b)
		assert.Equal(t, test.error, err != nil, "[test: %d] %v", i, err)
		if err != nil {
			return
		}
		assert.Equal(t, b, dbb.Balance)
		dbbb, err := s.(*postgres).selectBalanceByID(dbb.ID)
		common.FatalIfError(t, err, "selecting balance to check against inserted")
		assert.Equal(t, dbb, dbbb)
	}
}

func TestPostgres_SelectAccountBalances(t *testing.T) {
	deleteTestDBIgnorantly(t)
	store := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(store)
	count := 10
	as := newTestInsertedStorageAccounts(t, store, count)
	for i := 0; i < count; i++ {
		numBalances := i
		inserted := make([]storage.Balance, numBalances)
		for j, b := range newTestBalances(t, numBalances, time.Now(), time.Hour) {
			err := balance.Amount(j)(&b)
			common.FatalIfError(t, err, "setting balance amount")
			dba, err := store.InsertBalance(*as[i], b)
			common.FatalIfError(t, err, "inserting Balance")
			inserted[j] = *dba
		}
		returned, err := store.SelectAccountBalances(*as[i])
		common.FatalIfError(t, err, "selecting account balances")
		for j := 0; j < i; j++ {
			assert.Equal(t, inserted[j], (*returned)[j])
		}
	}
}

func newTestBalance(t *testing.T, time time.Time, os ...balance.Option) balance.Balance {
	b, err := balance.New(time, os...)
	common.FatalIfError(t, err, "creating test balance")
	return *b
}

func newTestBalances(t *testing.T, count int, startTime time.Time, interval time.Duration, os ...balance.Option) []balance.Balance {
	bs := make([]balance.Balance, count)
	for i := 0; i < count; i++ {
		bs[i] = newTestBalance(t, startTime.Add(time.Duration(i)*interval), os...)
	}
	return bs
}
