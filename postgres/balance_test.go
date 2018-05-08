// +build integration

package postgres_test

import (
	"testing"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/common"
	"github.com/stretchr/testify/assert"
)

func Test_BalancesForInvalidAccountId(t *testing.T) {
	validID := uint(1)
	store := prepareTestDB(t)
	defer nonReturningCloseStorage(t, store)
	as, err := store.SelectAccounts()
	assert.Nil(t, err)
	var selectedA storage.Account
	for _, a := range *as {
		if a.ID != validID {
			continue
		}
		selectedA = a
	}
	if !assert.NotNil(t, selectedA) {
		t.FailNow()
	}
	balances, err := store.SelectAccountBalances(selectedA)
	common.ErrorIfErrorf(t, err, "Getting balances for account %d", validID)
	minBalances := 91
	if len(*balances) < minBalances {
		t.Errorf("account ID: %d, expected at least %d balances but got: %d", validID, minBalances, len(*balances))
		return
	}
	expectedID := uint(1)
	actualID := (*balances)[0].ID
	if expectedID != actualID {
		t.Errorf(`Unexpected Balance ID.\nExpected: %d\nActual:  %d`, expectedID, actualID)
	}
	expectedAmount := 63641
	actualAmount := (*balances)[0].Amount
	assert.Equal(t, expectedAmount, actualAmount)
	expectedDate := time.Date(2016, 06, 17, 0, 0, 0, 0, time.UTC)
	actualDate := (*balances)[0].Date
	assert.True(t, expectedDate.Equal(actualDate))
}
