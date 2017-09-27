package GOHMoneyDB_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/GlynOwenHanmer/GOHMoney"
	"github.com/GlynOwenHanmer/GOHMoneyDB"
	"github.com/GlynOwenHanmer/GOHMoney/account"
	"github.com/GlynOwenHanmer/GOHMoney/balance"
)

func Test_CreateAccount(t *testing.T) {
	now := time.Now()
	testSets := []struct {
		name                 string
		start, expectedStart time.Time
		end, expectedEnd     GOHMoney.NullTime
		error
	}{
		{
			name:          "TEST_ACCOUNT",
			start:         now,
			expectedStart: now,
			end: GOHMoney.NullTime{
				Valid: true,
				Time:  now.AddDate(1, 0, 0),
			},
			expectedEnd: GOHMoney.NullTime{
				Valid: true,
				Time:  now.AddDate(1, 0, 0),
			},
			error: nil,
		},
		{
			name:          "TEST_ACCOUNT",
			start:         now,
			expectedStart: now,
			end:           GOHMoney.NullTime{Valid: false},
			expectedEnd:   GOHMoney.NullTime{Valid: false},
			error:         nil,
		},
		{
			name:          "Account With'Apostrophe",
			start:         now,
			expectedStart: now,
			end:           GOHMoney.NullTime{Valid: false},
			expectedEnd:   GOHMoney.NullTime{Valid: false},
			error:         nil,
		},
	}
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	for _, testSet := range testSets {
		newAccount, err := account.New(testSet.name, testSet.start, testSet.end)
		if err != nil {
			t.Fatalf("Error creating new account for testing. Error: %s", err.Error())
		}
		actualCreatedAccount, err := GOHMoneyDB.CreateAccount(db, newAccount)
		if testSet.error == nil && err != nil || testSet.error != nil && err == nil {
			t.Errorf("Unexpected error:\nExpected: %s\nActual  : %s", testSet.error, err)
		}
		if _, testSetErrIsNewAccountFieldError := testSet.error.(GOHMoney.AccountFieldError); testSetErrIsNewAccountFieldError {
			if _, actualErrorIsNewAccountFieldError := err.(GOHMoney.AccountFieldError); !actualErrorIsNewAccountFieldError {
				t.Errorf("Unexpected error:\nExpected: %s\nActual  : %s", testSet.error, err)
			}
		}
		if testSet.name != actualCreatedAccount.Name {
			t.Errorf("Unexpected created account name:\nExpected: %s\nActual  : %s", testSet.name, actualCreatedAccount.Name)
		}
		if !testSet.expectedStart.Truncate(24 * time.Hour).Equal(actualCreatedAccount.Start()) {
			t.Errorf("Unexpected account start.\nExpected: %s\nActual  : %s", testSet.expectedStart, actualCreatedAccount.Start())
		}
		testSet.expectedEnd.Time = testSet.expectedEnd.Time.Truncate(24 * time.Hour)
		if testSet.expectedEnd.Valid != actualCreatedAccount.End().Valid || !testSet.expectedEnd.Time.Equal(actualCreatedAccount.End().Time) {
			t.Errorf("Unexpected account end.\nExpected: %s\nActual  : %s", testSet.expectedEnd, actualCreatedAccount.End())
		}
	}
}

func Test_SelectAccounts(t *testing.T) {
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	accounts, err := GOHMoneyDB.SelectAccounts(db)
	if err != nil {
		if _, ok := err.(GOHMoney.AccountFieldError); !ok {
			t.Errorf("Unexpected error type when selecting accounts. Error: %s", err.Error())
		}
	}
	if accounts == nil {
		t.Fatalf("SelectAccounts returned nil Accounts object.\nError: %s", err)
	}
	checkAccountsSortedByIdAscending(*accounts, t)
}

func Test_SelectAccountsOpen(t *testing.T) {
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	openAccounts, err := GOHMoneyDB.SelectAccountsOpen(db)
	if err != nil {
		t.Fatalf("Error running SelectAccountsOpen method: %s", err)
	}
	for _, account := range openAccounts {
		if !account.IsOpen() {
			t.Errorf("SelectAccountsOpen returned closed account: %s", account)
		}
	}
	checkAccountsSortedByIdAscending(openAccounts, t)
}

func Test_SelectAccountWithId(t *testing.T) {
	tests := []struct {
		id            uint
		expectedError error
		name          string
	}{
		{
			id:            0,
			expectedError: GOHMoneyDB.NoAccountWithIdError(0),
		},
		{
			// Max for postgres smallint value
			id:            32767,
			expectedError: GOHMoneyDB.NoAccountWithIdError(32767),
		},
		{
			id:            10,
			expectedError: nil,
			name:          "Ikaros",
		},
		{
			id:            20,
			expectedError: nil,
			name:          "Amsterdam",
		},
	}
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	for _, test := range tests {
		if err != nil {
			t.Fatalf("Unable to open DB connection.\n%s", err)
		}
		account, err := GOHMoneyDB.SelectAccountWithID(db, test.id)
		if test.expectedError != err {
			t.Errorf("Unexpected errors\nExpected: %v\nActual  : %v", test.expectedError, err)
		}
		if _, noAccount := err.(GOHMoneyDB.NoAccountWithIdError); noAccount {
			continue
		}
		if test.id != account.Id {
			t.Errorf("Unexpected Account Id\nExpected: %d\nActual  : %d", test.id, account.Id)
		}
		if test.name != account.Name {
			t.Errorf("Unexpected Account name\nExpected: %s\nActual  : %s", test.name, account.Name)
		}
	}
}

func TestAccount_SelectBalanceWithID_InvalidID(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Unable to prepare DB for testings. Error: %s", err.Error())
	}
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	// Account with no Balances
	_, err = account.SelectBalanceWithId(db, 10)
	expectedErr := GOHMoneyDB.NoBalances
	if err != expectedErr {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedErr, err)
	}

	validBalance, err := account.InsertBalance(db,
		balance.Balance{
			Date:   account.Start().AddDate(0, 0, 10),
			Amount: float32(10),
		},
	)
	if err != nil {
		t.Fatalf("Error occurred whilst inserting Balance for testing. Error: %s", err)
	}
	if validBalance.Id < 1 {
		t.Fatalf("Inserted balance returned balance of less than 1 so cannot be subtracted from to make invalid uint Balance Id")
	}
	invalidBalanceId := validBalance.Id - 1
	// Account with Balances
	_, err = account.SelectBalanceWithId(db, invalidBalanceId)
	if err != expectedErr {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedErr, err)
	}
}

func TestAccount_SelectBalanceWithID_ValidId(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Unable to prepare DB for testings. Error: %s", err.Error())
	}
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	var balances [3]GOHMoneyDB.Balance
	for i := 0; i < 3; i++ {
		balances[i], err = account.InsertBalance(db,
			balance.Balance{
				Date:   account.Start().AddDate(0, 0, i),
				Amount: float32(i),
			},
		)
	}
	for _, balance := range balances {
		selectedBalance, err := account.SelectBalanceWithId(db, balance.Id)
		if err != nil {
			t.Errorf("Expected nil error but recieved error: %s", err)
		}
		switch {
		case selectedBalance.Id != balance.Id,
			selectedBalance.Amount != balance.Amount,
			!selectedBalance.Date.Equal(balance.Date):
			t.Errorf("Unexpected Balance returned.\n\tExpected: %s\n\tActual  : %s", balance, selectedBalance)
		}
	}
}

func checkAccountsSortedByIdAscending(accounts GOHMoneyDB.Accounts, t *testing.T) {
	for i := 0; i+1 < len(accounts); i++ {
		account := accounts[i]
		nextAccount := accounts[i+1]
		switch {
		case account.Id > nextAccount.Id:
			var message bytes.Buffer
			fmt.Fprintf(&message, "Accounts not returned sorted by Id. Id %d appears before %d.\n", account.Id, nextAccount.Id)
			fmt.Fprintf(&message, "accounts[%d]: %s", i, account)
			fmt.Fprintf(&message, "accounts[%d]: %s", i+1, nextAccount)
			t.Errorf(message.String())
		}
	}
}

func newTestAccount() account.Account {
	account, err := account.New(
		"TEST_ACCOUNT",
		time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
		GOHMoney.NullTime{
			Valid: true,
			Time:  time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC),
		},
	)
	if err != nil {
		panic(err)
	}
	return account
}

func newTestDBAccount(db *sql.DB) GOHMoneyDB.Account {
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	if err != nil {
		panic(err)
	}
	return *account
}

func TestAccount_UpdateAccount(t *testing.T) {
	now := time.Now()
	original, err := account.New("TEST_ACCOUNT", now, GOHMoney.NullTime{})
	if err != nil {
		t.Fatalf("Error creating a for testing: %s", err)
	}
	updatedStart := now.AddDate(1, 0, 0)
	updatedEnd := GOHMoney.NullTime{Valid: true, Time: updatedStart.AddDate(2, 0, 0)}
	update, err := account.New("TEST_ACCOUNT_UPDATED", updatedStart, updatedEnd)
	if err != nil {
		t.Fatalf("Error creating a for testing: %s", err)
	}
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error preparing test DB: %s", err)
	}
	a, err := GOHMoneyDB.CreateAccount(db, original)
	if err != nil {
		t.Fatalf("Error creating a: %s", err)
	}
	updated, err := a.Update(db, update)
	if err != nil {
		t.Errorf("Error updating a: %s", err)
	}
	expected, err := account.New(
		update.Name,
		update.Start().Truncate(24*time.Hour),
		GOHMoney.NullTime{
			Valid: update.End().Valid,
			Time:  update.End().Time.Truncate(24 * time.Hour),
		},
	)
	if !updated.Account.Equal(expected) {
		t.Errorf("Updates not applied as expected.\nUpdated a: %s\nApplied updates: %s", updated, expected)
	}
}

func TestAccount_Delete(t *testing.T) {
	invalid := GOHMoneyDB.Account{}
	if invalid.Delete(nil) == nil {
		t.Errorf("Expected error but none was returned when attempting to delete an invalid account.")
	}
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error preparing DB for testing.")
	}
	account := newTestDBAccount(db)
	if account.Validate(db) != nil {
		t.Fatalf("Invalid account returned for testing. Details: %s", err)
	}
	if err := account.Delete(db); err != nil {
		t.Fatalf("Error occured whilst deleting account. Error: %s", err)
	}
	valid := account.Validate(db)
	if valid == nil {
		t.Fatalf("Account still valid after deletion.")
	} else if valid != GOHMoneyDB.AccountDeleted {
		t.Fatalf("Validity error not as expected. Expected %s, got %s.", GOHMoneyDB.AccountDeleted, valid)
	}
}

func TestAccount_JsonLoop(t *testing.T) {
	innerAccount, err := account.New(
		"TEST",
		time.Now(),
		GOHMoney.NullTime{
			Valid: true,
			Time:  time.Now().AddDate(1, 0, 0),
		},
	)
	if err != nil {
		t.Fatalf("Error creating new account for testing. Error: %s", err.Error())
	}
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error preparing db for testing: %s", err)
	}
	originalAccount, err := GOHMoneyDB.CreateAccount(db, innerAccount)
	if err != nil {
		t.Fatalf("Error creating DB account for testing: %s", err)
	}
	originalBytes, err := json.Marshal(originalAccount)
	if err != nil {
		t.Fatalf("Error marshalling account into json. Error: %s", err.Error())
	}
	var finalAccount GOHMoneyDB.Account
	json.Unmarshal(originalBytes, &finalAccount)
	logBytes := func(t *testing.T) { t.Log("Marshalled account: " + string(originalBytes)) }
	if finalAccount.Id != originalAccount.Id {
		t.Errorf("Unexpected account id.\n\tExpected: %d\n\tActuall  : %d", originalAccount.Id, finalAccount.Id)
		logBytes(t)
	}
	if !originalAccount.Start().Equal(finalAccount.Start()) {
		t.Errorf("Unexpected account Start.\n\tExpected: %s\n\tActual  : %s", originalAccount.Start(), finalAccount.Start())
		logBytes(t)
	}
	if originalAccount.End().Valid != finalAccount.End().Valid || !originalAccount.End().Time.Equal(finalAccount.End().Time) {
		t.Errorf("Unexpected account End. \n\tExpected: %s\n\tActual  : %s", originalAccount.End(), finalAccount.End())
		logBytes(t)
	}
}

func TestAccount_Validate(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error preparing DB for testing: %s", err)
	}
	defer db.Close()
	invalid := GOHMoneyDB.Account{}
	err = invalid.Validate(db)
	if err == nil {
		t.Errorf("Expected expected but none returned.")
	}
	if expected := GOHMoneyDB.NoAccountWithIdError(0); err != expected {
		t.Errorf("Expected error %s, but got %s", expected, err)
	}
	invalid.Id = 5
	err = invalid.Validate(db)
	if expected := GOHMoneyDB.AccountDifferentInDbAndRuntime; err != expected {
		t.Errorf("Expected error %s, but got %s", expected, err)
	}

	valid, err := GOHMoneyDB.SelectAccountWithID(db, 1)
	if err != nil {
		t.Fatalf("Error selecting valid account for testing: %s", err)
	}
	if validErr := valid.Validate(db); validErr != nil {
		t.Errorf("Expected nil error but got %s", validErr)
	}
}
