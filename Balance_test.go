package GOHMoneyDB_test

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"testing"
	"time"

	"encoding/json"

	"github.com/glynternet/GOHMoney/balance"
	"github.com/glynternet/GOHMoney/common"
	"github.com/glynternet/GOHMoney/money"
	"github.com/glynternet/GOHMoneyDB"
	"github.com/stretchr/testify/assert"
)

func Test_BalancesForInvalidAccountId(t *testing.T) {
	validID := uint(1)
	db := prepareTestDB(t)
	defer close(t, db)
	a, err := GOHMoneyDB.SelectAccountWithID(db, validID)
	if err != nil {
		t.Errorf("Error selecting account for testings: %s", err)
	}
	invalidIDs := []uint{0, 99999}
	for _, invalidID := range invalidIDs {
		a.ID = invalidID
		balances, err := a.Balances(db)
		if err == nil {
			t.Errorf("account ID: %d, expected error but got nil", a.ID)
		}
		if balances != nil && len(*balances) != 0 {
			t.Errorf("account ID: %d, expected no balances but got: %s", invalidID, balances)
		}
	}
}

func TestBalancesForValidAccountId(t *testing.T) {
	validID := uint(1)
	db := prepareTestDB(t)
	defer close(t, db)
	a, err := GOHMoneyDB.SelectAccountWithID(db, validID)
	common.ErrorIfError(t, err, "Selecting accounts for testing")
	balances, err := a.Balances(db)
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
	expectedAmount, err := money.New(63641, "GBP")
	common.FatalIfError(t, err, "Creating Money")
	actualAmount := (*balances)[0].Money()
	if eq, err := actualAmount.Equal(*expectedAmount); !eq || (err != nil) {
		t.Errorf("account ID: %d, first balance, expected balance amount of %+v but got %+v", validID, expectedAmount, actualAmount)
	}
	expectedDate := time.Date(2016, 06, 17, 0, 0, 0, 0, time.UTC)
	actualDate := (*balances)[0].Date()
	if !expectedDate.Equal(actualDate) {
		t.Errorf("account ID: %d, first balance, expected date of %s but got %s", validID, expectedDate, actualDate)
	}
}

func Test_BalanceInsert_InvalidBalance(t *testing.T) {
	accountID := uint(1)
	db := prepareTestDB(t)
	defer close(t, db)
	account, err := GOHMoneyDB.SelectAccountWithID(db, accountID)
	common.ErrorIfError(t, err, "Selecting account")
	dbAccount := GOHMoneyDB.Account(account)
	startingBalances, err := dbAccount.Balances(db)
	common.FatalIfErrorf(t, err, "Getting balances for account %+v", dbAccount)
	invalidBalance := balance.Balance{}
	insertedBalance, err := dbAccount.InsertBalance(db, invalidBalance)
	if err != balance.ZeroDate {
		t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s", balance.ZeroDate, err)
	}
	if insertedBalance.ID != 0 {
		t.Errorf("Expected uninitialised Balance id of 0, got %d", insertedBalance.ID)
	}
	if !insertedBalance.Date().IsZero() {
		t.Errorf("Inserted balance date should be zero but is: %s", insertedBalance.Date().String())
	}
	actual := insertedBalance.Money()
	assert.Equal(t, actual.Amount(), int64(0))
	_, err = actual.Currency()
	assert.Equal(t, money.ErrNoCurrency, err)
	assert.Equal(t, insertedBalance.ID, uint(0))
	balancesAfterTest, err := dbAccount.Balances(db)
	common.FatalIfError(t, err, "Getting Balances")
	assert.Equal(t, len(*startingBalances), len(*balancesAfterTest), "Number of Balances changed")
}

func TestAccount_InsertBalance_ValidBalance(t *testing.T) {
	accountID := uint(1)
	db := prepareTestDB(t)
	defer close(t, db)
	initialLastID := getHighestBalanceID(db, t)
	account, err := GOHMoneyDB.SelectAccountWithID(db, accountID)
	common.ErrorIfErrorf(t, err, "Error selecting account with ID %d for testing", accountID)
	dbAccount := GOHMoneyDB.Account(account)
	startingBalances, err := dbAccount.Balances(db)
	if err != nil {
		t.Fatalf("Unable to get balances for testing for account: %s", dbAccount)
	}
	validDate := time.Date(3000, 6, 1, 1, 1, 1, 1, time.UTC).Truncate(time.Hour * 24)
	m, err := money.New(123456, "USD")
	common.FatalIfError(t, err, "Creating Money")
	validBalance, _ := balance.New(validDate, *m)
	insertedBalance, err := dbAccount.InsertBalance(db, validBalance)
	common.ErrorIfError(t, err, "Expected nil error bit got")
	if insertedBalance.ID != initialLastID+1 {
		t.Errorf("Expected ID to incremement by 1.\nInitial last ID: %d\nInserted Balance ID: %d", initialLastID, insertedBalance.ID)
	}
	if !insertedBalance.Balance.Equal(validBalance) {
		t.Errorf("Inserted balance does not equal original.\nInserted: %+v\nOriginal: %+v", insertedBalance.Balance, validBalance)
	}
	err = dbAccount.ValidateBalance(db, insertedBalance)
	common.ErrorIfErrorf(t, err, "Expected inserted balance to be valid against account.\nError: %s\nAccount: %s\nBalance: %s", err, dbAccount, insertedBalance)
	balancesAfterTest, err := dbAccount.Balances(db)
	common.ErrorIfErrorf(t, err, "Unable to get balances for testing for account: %s", dbAccount)
	expectedDiff := 1
	balancesCountDiff := len(*balancesAfterTest) - len(*startingBalances)
	if balancesCountDiff != expectedDiff {
		t.Errorf("Number of balances should changed by %d but changed by %d", expectedDiff, balancesCountDiff)
	}
}

func getHighestBalanceID(db *sql.DB, t *testing.T) uint {
	initialLastID := uint(0)
	allAccounts, err := GOHMoneyDB.SelectAccounts(db)
	if err != nil {
		t.Fatalf("Error selecting all Accounts for testing: %s", err.Error())
	}
	if len(*allAccounts) < 1 {
		t.Fatalf("No Accounts were selected.")
	}
	for _, account := range *allAccounts {
		balances, err := account.Balances(db)
		if err == GOHMoneyDB.NoBalances {
			continue
		}
		if err != nil {
			t.Fatalf("Error selecting balances for testing: %s", err.Error())
		}
		for _, balance := range *balances {
			if balance.ID > initialLastID {
				initialLastID = balance.ID
			}
		}
	}
	return initialLastID
}

func TestAccount_ValidateBalance(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	m, err := money.New(0, "YEN")
	common.FatalIfError(t, err, "Creating Money")
	b, _ := balance.New(account.Start().AddDate(0, 0, 1), *m)
	validBalance, err := account.InsertBalance(db, b)
	if err != nil {
		t.Fatalf(`Error inserting new balance for testing. Error :%s`, err)
	}
	outOfDateRange := GOHMoneyDB.Balance{
		ID:      account.ID,
		Balance: newInnerBalanceIgnoreError(time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC), 0, "USD"),
	}
	balanceWithWrongOwner := GOHMoneyDB.Balance{
		ID:      account.ID - 1,
		Balance: validBalance.Balance,
	}
	testSets := []struct {
		account *GOHMoneyDB.Account
		balance GOHMoneyDB.Balance
		error
	}{
		{
			account: account,
			balance: validBalance,
			error:   nil,
		},
		{
			account: account,
			balance: outOfDateRange,
			error:   account.Account.ValidateBalance(outOfDateRange.Balance),
		},
		{
			account: account,
			balance: balanceWithWrongOwner,
			error:   GOHMoneyDB.InvalidAccountBalanceError{AccountID: account.ID, BalanceID: balanceWithWrongOwner.ID},
		},
	}
	for _, testSet := range testSets {
		err := testSet.account.ValidateBalance(db, testSet.balance)
		if err != testSet.error {
			t.Fatalf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", testSet.error, err)
		}
	}
}

func newInnerBalanceIgnoreError(t time.Time, a int64, cur string) balance.Balance {
	m, _ := money.New(a, cur)
	b, _ := balance.New(t, *m)
	return b
}

func Test_UpdateBalance_WrongAccount(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	accountA, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	accountB, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := newInnerBalanceIgnoreError(accountA.Start(), 0, "EUR")
	createdBalanceA, err := accountA.InsertBalance(db, newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := newInnerBalanceIgnoreError(time.Now(), 100, "EUR")
	updatedBalance, err := accountB.UpdateBalance(db, createdBalanceA, update)
	expectedError := GOHMoneyDB.InvalidAccountBalanceError{AccountID: accountB.ID, BalanceID: createdBalanceA.ID}
	if err != expectedError {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err)
	}
	expectedBalance := GOHMoneyDB.Balance{}
	if updatedBalance != expectedBalance {
		t.Errorf("Unexpected Balance.\n\tExpected: %s\n\tActual  : %s", expectedBalance, updatedBalance)
	}
}

func Test_UpdateBalance_InvalidUpdate(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := newInnerBalanceIgnoreError(account.Start(), 0, "CZK")
	createdBalance, err := account.InsertBalance(db, newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := balance.Balance{}
	updatedBalance, err := account.UpdateBalance(db, createdBalance, update)
	expectedError := errors.New(`Update Balance is not valid: ` + balance.ZeroDate.Error())
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err)
	}
	expectedBalance := GOHMoneyDB.Balance{}
	if updatedBalance != expectedBalance {
		t.Errorf("Unexpected Balance.\n\tExpected: %s\n\tActual  : %s", expectedBalance, updatedBalance)
	}
}

// Test for when the update Balance that is trying to be applied is not valid in the context of the account. For example, where the Data of the update is outside of the TimeRange of the account.
func Test_UpdateBalance_InvalidUpdateForAccount(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := newInnerBalanceIgnoreError(account.Start(), 0, "EUR")
	createdBalance, err := account.InsertBalance(db, newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := newInnerBalanceIgnoreError(account.Start().AddDate(-1, 0, 0), 0, "USD")
	_, err = account.UpdateBalance(db, createdBalance, update)
	expectedError := `Update is not valid for account: ` + balance.DateOutOfAccountTimeRange{}.Error()
	if err == nil {
		t.Errorf("Expected error but got nil.")
	} else if err.Error() != expectedError {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err.Error())
	}
}

func Test_UpdateBalance_ValidBalance(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := newInnerBalanceIgnoreError(account.Start(), 0, "USD")
	createdBalance, err := account.InsertBalance(db, newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := newInnerBalanceIgnoreError(account.Start().AddDate(0, 0, 1), 100, "EUR")
	updatedBalance, err := account.UpdateBalance(db, createdBalance, update)
	expectedError := error(nil)
	if err != expectedError {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err)
	}
	if createdBalance.ID != updatedBalance.ID {
		t.Errorf("Balance ID changed when updating Balance\n\tOriginal: %d\n\tFinal   : %d", createdBalance.ID, updatedBalance.ID)
	}
	expectedDate := update.Date().Truncate(time.Hour * 24)
	if !updatedBalance.Balance.Date().Equal(expectedDate) {
		t.Errorf("Unexpected Balance date.\n\tExpected: %s\n\tActual  : %s", update.Date(), updatedBalance.Balance.Date())
	}
	appliedAmount := update.Money()
	updatedAmount := updatedBalance.Money()
	equal, err := (&appliedAmount).Equal(updatedAmount)
	if err != nil {
		t.Errorf("Error comparing amounts. Err: %s", err)
	}
	if !equal {
		t.Errorf("Unexpected Balance Amount.\n\tExpected: %s\n\tActual  : %s", update.Money(), updatedBalance.Money())
	}
}

func Test_AccountBalanceAtDate(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	dbAccount, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf("Unable to create account for testing. Error: %s", err.Error())
	}

	// 1. no balances exist for account
	b, err := dbAccount.BalanceAtDate(db, time.Date(3000, 1, 2, 1, 1, 1, 1, time.UTC))
	expectedError := GOHMoneyDB.NoBalances
	if err != expectedError {
		t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s", expectedError, err)
	}
	expectedBalance := GOHMoneyDB.Balance{}
	assert.Equal(t, expectedBalance, b)

	balances := [5]balance.Balance{
		newInnerBalanceIgnoreError(time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC), 0, "GBP"),
		newInnerBalanceIgnoreError(time.Date(2000, 1, 3, 1, 1, 1, 1, time.UTC), 1, "GBP"),
		newInnerBalanceIgnoreError(time.Date(2000, 1, 5, 1, 1, 1, 1, time.UTC), 2, "GBP"),
		newInnerBalanceIgnoreError(time.Date(2000, 1, 5, 1, 1, 1, 1, time.UTC), 3, "GBP"),
		newInnerBalanceIgnoreError(time.Date(2000, 1, 7, 1, 1, 1, 1, time.UTC), 4, "GBP"),
	}

	var insertedBalances []GOHMoneyDB.Balance
	for _, balance := range balances {
		insertedBalance, err := dbAccount.InsertBalance(db, balance)
		if err != nil {
			var message bytes.Buffer
			fmt.Fprintf(&message, "Error when inserting b for testing. Error: %s", err.Error())
			fmt.Fprintf(&message, "\nBalance: %v", balance)
			fmt.Fprintf(&message, "\nAccount: %v", dbAccount)
			t.Fatal(message.String())
		}
		insertedBalances = append(insertedBalances, insertedBalance)
	}

	testSets := []struct {
		time.Time
		error
		expectedBalance GOHMoneyDB.Balance
	}{
		{
			// No balances exist before date
			balances[0].Date().AddDate(-1, 0, 0),
			GOHMoneyDB.NoBalances,
			GOHMoneyDB.Balance{},
		},
		{
			// Balance exists before date
			balances[0].Date().AddDate(0, 0, 1),
			nil,
			insertedBalances[0],
		},
		{
			// Balance is on date
			balances[1].Date(),
			nil,
			insertedBalances[1],
		},
		{
			// Multiple balances match date. Should return one with highest ID (latest inserted)
			balances[2].Date(),
			nil,
			insertedBalances[3],
		},
		{
			// Multiple have the same date that is before and is the closest to the given date
			balances[4].Date().AddDate(0, 0, -1),
			nil,
			insertedBalances[3],
		},
	}

	for _, testSet := range testSets {
		b, err = dbAccount.BalanceAtDate(db, testSet.Time)
		if err != testSet.error {
			message := fmt.Sprintf("Unexpected error.\nExpected: %s\nActual  : %s", testSet.error, err)
			message += fmt.Sprintf("\nFor time: %v", testSet.Time)
			t.Error(message)
		}
		assert.Equal(t, testSet.expectedBalance, b, "For time: %v", testSet.Time)
	}
}

func TestBalance_JSONLoop(t *testing.T) {
	a := GOHMoneyDB.Balance{ID: 3, Balance: newInnerBalanceIgnoreError(time.Now(), 7654, "USD")}
	jsonBytes, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("Error marshalling json for testing: %s", err)
	}

	var b = struct {
		ID    uint
		Date  time.Time
		Money money.Money
	}{}
	err = json.Unmarshal(jsonBytes, &b)
	common.FatalIfError(t, err, "Unmarshaling json for testing")
	assert.Equal(t, a.ID, b.ID, "JSON: %s", string(jsonBytes))
	assert.Equal(t, a.Date(), b.Date)
	assert.Equal(t, a.Money(), b.Money)

	var c GOHMoneyDB.Balance
	if err := json.Unmarshal(jsonBytes, &c); err != nil {
		t.Fatalf("Error unmarshaling bytes into Balance: %s\njson: %s", err, jsonBytes)
	}
	assert.Equal(t, a, c)
	if !a.Equal(c) {
		t.Fatalf("Expected %v, but got %v\njson: %s", a, c, jsonBytes)
	}
}
