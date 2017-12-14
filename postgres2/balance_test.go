package postgres2_test

import (
	"testing"
	"time"

	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/common"
	"github.com/stretchr/testify/assert"
)

func TestPostgres_InsertBalance(t *testing.T) {
	s := createTestDB(t)
	defer deleteTestDB(t)
	defer nonReturningCloseStorage(t, s)

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
		if err == nil {
			assert.Equal(t, b, dbb.Balance)
		}
	}
}

func newTestBalance(t *testing.T, time time.Time, os ...balance.Option) balance.Balance {
	b, err := balance.New(time, os...)
	common.FatalIfError(t, err, "creating test balance")
	return *b
}
