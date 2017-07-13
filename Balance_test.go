package GOHMoneyDB

import (
	"testing"
	"github.com/GlynOwenHanmer/GOHMoney"
	"time"
	"bytes"
	"fmt"
)

func Test_BalancesForInvalidAccountId(t *testing.T) {
	invalidIds := []uint{0, 99999}
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection. Error: %s", err)
	}
	for _, invalidId := range invalidIds {
		balances, err := selectBalancesForAccount(db, invalidId)
		if err != nil {
			t.Errorf("account Id: %d, expected nil error but got: %s", invalidId, err.Error())
		}
		if len(balances) != 0 {
			t.Errorf("account Id: %d, expected no balances but got: %s", invalidId, balances)
		}
	}
}

func Test_BalancesForValidAccountId(t *testing.T) {
	validId := uint(1)
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection. Error: %s", err)
	}
	balances, err := selectBalancesForAccount(db, validId)
	if err != nil {
		t.Errorf("Id: %d, expected nil error but got: %s", validId, err.Error())
	}
	minBalances := 91
	if len(balances) < minBalances {
		t.Errorf("account Id: %d, expected at least %d balances but got: %d", validId, minBalances, len(balances))
		return
	}
	expectedAmount := float32(636.42)
	actualAmount := balances[0].Amount
	if actualAmount != expectedAmount {
		t.Errorf("account Id: %d, first balance, expected balance amount of %f but got %f", validId, expectedAmount, actualAmount)
	}
	expectedDate, err := ParseDateString("2016-06-17")
	if err != nil {
		t.Fatalf("Error parsing date string for use in tests. Error: %s", err.Error())
	}
	actualDate := balances[0].Date
	if !expectedDate.Equal(actualDate) {
		t.Errorf("account Id: %d, first balance, expected date of %s but got %s", validId, formatDateString(expectedDate), formatDateString(actualDate))
	}
}

func Test_BalanceInsert(t *testing.T) {
	accountId :=uint(1)
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection. Error: %s", err)
	}
	account, err := SelectAccountWithID(db, accountId)
	if err != nil {
		t.Errorf("Error selecting account with Id %d for testing: %s", accountId, err.Error())
	}
	dbAccount := Account(account)
	startingBalances, err := dbAccount.Balances(db)
	if err != nil {
		t.Fatalf("Unable to get balances for testing for account: %s\nError: %s", dbAccount, err)
	}
	invalidBalance := GOHMoney.Balance{}
	insertedBalance, err := dbAccount.InsertBalance(db, invalidBalance)
	if err != GOHMoney.BalanceZeroDate {
		t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s", GOHMoney.BalanceZeroDate, err)
	}
	if !insertedBalance.Date.IsZero() {
		t.Errorf("Inserted balance date should be zero but is: %s", insertedBalance.Date.String())
	}
	if insertedBalance.Amount != 0 {
		t.Errorf("Inserted balance amount should be %f but is %f", 0, insertedBalance.Amount)
	}
	if insertedBalance.Id != 0 {
		t.Errorf("Inserted balance Id should be %d but is %d", 0, insertedBalance.Id)
	}
	balancesAfterTest, err := dbAccount.Balances(db)
	if err != nil {
		t.Fatalf("Unable to get balances for testing for account: %s", dbAccount)
	}
	if len(startingBalances) != len(balancesAfterTest) {
		t.Errorf("Number of balances changed during test.\nBefore: %d\nAfter : %d", len(startingBalances), len(balancesAfterTest))
	}
	validDate := time.Date(3000, 6, 1, 1, 1, 1, 1,time.UTC)
	validBalance := GOHMoney.Balance{Date:validDate, Amount:1234.56}
	startingBalances, err = dbAccount.Balances(db)
	if err != nil {
		t.Fatalf("Unable to get balances for testing for account: %s", dbAccount)
	}
	insertedBalance, err = dbAccount.InsertBalance(db, validBalance)
	if err != nil {
		t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s", nil, err)
	}
	if !insertedBalance.Date.Equal(validDate.Truncate(time.Hour * 24)) {
		t.Errorf("Inserted balance date should be %s but is %s", validDate, insertedBalance.Date.String())
	}
	if insertedBalance.Amount != validBalance.Amount {
		t.Errorf("Inserted balance amount should be %f but is %f", validBalance.Amount, insertedBalance.Amount)
	}
	balancesAfterTest, err = dbAccount.Balances(db)
	if err != nil {
		t.Fatalf("Unable to get balances for testing for account: %s", dbAccount)
	}
	expectedDiff := 1
	balancesCountDiff := len(balancesAfterTest) - len(startingBalances)
	if balancesCountDiff != expectedDiff {
		t.Errorf("Number of balances should changed by %d but changed by %d", expectedDiff, balancesCountDiff)
	}
}

func Test_UpdateBalance(t *testing.T) {
	timeStart := time.Now()
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error when prepping test DB. Error: %s", err.Error())
	}
	account, err := CreateAccount(db,newTestAccount())
	newBalance := GOHMoney.Balance{
		Date:timeStart.AddDate(500, 1, 2),
		Amount:0,
	}
	createdBalance, err := account.InsertBalance(db,newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := GOHMoney.Balance{
		Date:timeStart,
		Amount:100,
	}
	updatedBalance, err := account.UpdateBalance(db,createdBalance,update)
	_ = updatedBalance
	t.Fail() // WIP so fail to disallow builds
}

func Test_AccountBalanceAtDate(t *testing.T) {
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection. Error: %s", err)
	}
	dbAccount, err := CreateAccount(db, newTestAccount())
	if err != nil {
		t.Fatalf("Unable to create account for testing. Error: %s", err.Error())
	}

	// 1. no balances exist for account
	balance, err := dbAccount.BalanceAtDate(db, time.Date(3000, 1, 2, 1, 1, 1, 1, time.UTC),)
	expectedError := NoBalances
	if err != expectedError {
		t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s", expectedError, err)
	}
	expectedBalance := Balance{}
	if balance != expectedBalance {
		t.Errorf("Unexpected balance.\nExpected: %s\nActual  : %s", expectedBalance, balance)
	}

	balances := [5]GOHMoney.Balance{
		{
			time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
			0,
		},
		{
			time.Date(2000, 1, 3, 1, 1, 1, 1, time.UTC),
			1,
		},
		{
			time.Date(2000, 1, 5, 1, 1, 1, 1, time.UTC),
			2,
		},
		{
			time.Date(2000, 1, 5, 1, 1, 1, 1, time.UTC),
			3,
		},
		{
			time.Date(2000, 1, 7, 1, 1, 1, 1, time.UTC),
			4,
		},
	}

	var insertedBalances []Balance
	for _, balance := range balances {
		insertedBalance, err := dbAccount.InsertBalance(db, balance)
		if err != nil {
			var message bytes.Buffer
			fmt.Fprintf(&message, "Error when inserting balance for testing. Error: %s", err.Error())
			fmt.Fprintf(&message, "\nBalance: %v", balance)
			fmt.Fprintf(&message, "\nAccount: %v", dbAccount)
			t.Fatal(message.String())
		}
		insertedBalances = append(insertedBalances, insertedBalance)
	}

	testSets := []struct{
		time.Time
		error
		expectedBalance Balance
	}{
		{
			// No balances exist before date
			balances[0].Date.AddDate(-1,0,0),
			NoBalances,
			Balance{},
		},
		{
			// Balance exists before date
			balances[0].Date.AddDate(0,0,1),
			nil,
			insertedBalances[0],
		},
		{
			// Balance is on date
			balances[1].Date,
			nil,
			insertedBalances[1],
		},
		{
			// Multiple balances match date. Should return one with highest Id (latest inserted)
			balances[2].Date,
			nil,
			insertedBalances[3],
		},
		{
			// Multiple have the same date that is before and is the closest to the given date
			balances[4].Date.AddDate(0,0,-1),
			nil,
			insertedBalances[3],
		},
	}

	for _, testSet := range testSets {
		balance, err = dbAccount.BalanceAtDate(db, testSet.Time)
		if err != testSet.error {
			message := fmt.Sprintf("Unexpected error.\nExpected: %s\nActual  : %s", testSet.error, err)
			message += fmt.Sprintf("\nFor time: %v", testSet.Time)
			t.Error(message)
		}
		if balance != testSet.expectedBalance {
			message := fmt.Sprintf("Unexpected balance.\nExpected: %s\nActual  : %s", expectedBalance, balance)
			message += fmt.Sprintf("\nFor time: %v", testSet.Time)
			t.Error(message)
		}
	}
}


