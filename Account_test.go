package GOHMoneyDB_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/GlynOwenHanmer/GOHMoney/account"
	"github.com/GlynOwenHanmer/GOHMoneyDB"
	gohtime "github.com/GlynOwenHanmer/go-time"
	"github.com/stretchr/testify/assert"
)

func Test_CreateAccount(t *testing.T) {
	now := time.Now()
	testSets := []struct {
		name                 string
		start, expectedStart time.Time
		end, expectedEnd     gohtime.NullTime
		error
	}{
		{
			name:  "TEST_ACCOUNT",
			start: now,
			end: gohtime.NullTime{
				Valid: true,
				Time:  now.AddDate(1, 0, 0),
			},
			error: nil,
		},
		{
			name:  "TEST_ACCOUNT",
			start: now,
			end:   gohtime.NullTime{Valid: false},
			error: nil,
		},
		{
			name:  "Account With'Apostrophe",
			start: now,
			end:   gohtime.NullTime{Valid: false},
			error: nil,
		},
	}
	db := prepareTestDB(t)
	defer close(t, db)
	for _, testSet := range testSets {
		newAccount, err := account.New(testSet.name, testSet.start, testSet.end)
		fatalIfError(t, err, "Error creating new account for testing")
		actualCreatedAccount, err := GOHMoneyDB.CreateAccount(db, newAccount)
		if testSet.error == nil && err != nil || testSet.error != nil && err == nil {
			t.Errorf("Unexpected error:\nExpected: %s\nActual  : %s", testSet.error, err)
		}
		if _, testSetErrIsNewAccountFieldError := testSet.error.(account.FieldError); testSetErrIsNewAccountFieldError {
			if _, actualErrorIsNewAccountFieldError := err.(account.FieldError); !actualErrorIsNewAccountFieldError {
				t.Errorf("Unexpected error:\nExpected: %s\nActual  : %s", testSet.error, err)
			}
		}
		expectedAccount, err := account.New(
			testSet.name,
			testSet.start.Truncate(24*time.Hour),
			gohtime.NullTime{
				Valid: testSet.end.Valid,
				Time:  testSet.end.Time.Truncate(24 * time.Hour),
			},
		)
		fatalIfError(t, err, "Error creating account for testing")
		if !actualCreatedAccount.Account.Equal(expectedAccount) {
			t.Errorf("Unexpected account:\nExpected: %+v\nActual  : %+v", expectedAccount, actualCreatedAccount)
		}
		//todo Check that id has incremented by one?
	}
}

func Test_SelectAccounts(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	accounts, err := GOHMoneyDB.SelectAccounts(db)
	if err != nil {
		if _, ok := err.(account.FieldError); !ok {
			t.Errorf("Unexpected error type when selecting accounts. Error: %s", err.Error())
		}
	}
	if accounts == nil {
		t.Fatalf("SelectAccounts returned nil Accounts object.\nError: %s", err)
	}
	if len(*accounts) == 0 {
		t.Fatalf("No accounts were returned.")
	}
	checkAccountsSortedByIdAscending(*accounts, t)
}

func Test_SelectAccountsOpen(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	openAccounts, err := GOHMoneyDB.SelectAccountsOpen(db)
	fatalIfError(t, err, "Error running SelectAccountsOpen method")
	if len(*openAccounts) == 0 {
		t.Fatalf("No accounts were returned.")
	}
	for _, account := range *openAccounts {
		if !account.IsOpen() {
			t.Errorf("SelectAccountsOpen returned closed account: %s", account)
		}
	}
	checkAccountsSortedByIdAscending(*openAccounts, t)
}

func Test_SelectAccountWithId(t *testing.T) {
	tests := []struct {
		id            uint
		expectedError error
		name          string
	}{
		{
			id:            0,
			expectedError: GOHMoneyDB.NoAccountWithIDError(0),
		},
		{
			// Max for postgres smallint value
			id:            32767,
			expectedError: GOHMoneyDB.NoAccountWithIDError(32767),
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
	db := prepareTestDB(t)
	defer close(t, db)
	for _, test := range tests {
		account, err := GOHMoneyDB.SelectAccountWithID(db, test.id)
		if test.expectedError != err {
			t.Errorf("Unexpected errors\nExpected: %v\nActual  : %v", test.expectedError, err)
		}
		if _, noAccount := err.(GOHMoneyDB.NoAccountWithIDError); noAccount {
			continue
		}
		if test.id != account.ID {
			t.Errorf("Unexpected Account ID\nExpected: %d\nActual  : %d", test.id, account.ID)
		}
		if test.name != account.Name {
			t.Errorf("Unexpected Account name\nExpected: %s\nActual  : %s", test.name, account.Name)
		}
	}
}

func TestAccount_SelectBalanceWithID_InvalidID(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	fatalIfError(t, err, "Error inserting account for testing")
	// Account with no Balances
	b, err := account.SelectBalanceWithID(db, 10)
	expectedErr := GOHMoneyDB.NoBalances
	if err != expectedErr {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedErr, err)
		t.Logf("Selected balance: %v", b)
	}

	innerBalance := newInnerBalanceIgnoreError(account.Start().AddDate(0, 0, 10), 10, "GBP")
	validBalance, err := account.InsertBalance(db, innerBalance)
	fatalIfError(t, err, "Error occurred whilst inserting Balance for testing")
	if validBalance.ID < 1 {
		t.Fatalf("Inserted balance returned balance of less than 1 so cannot be subtracted from to make invalid uint Balance ID")
	}
	invalidBalanceId := validBalance.ID - 1
	// Account with Balances
	_, err = account.SelectBalanceWithID(db, invalidBalanceId)
	if err != expectedErr {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedErr, err)
	}
}

func TestAccount_SelectBalanceWithID_ValidId(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	account := newTestDBAccount(t, db)
	var balances [3]GOHMoneyDB.Balance
	for i := 0; i < 3; i++ {
		var err error
		balances[i], err = account.InsertBalance(db, newInnerBalanceIgnoreError(account.Start().AddDate(0, 0, i), int64(i), "GBP"))
		fatalIfError(t, err, "Error inserting Balance for testing")
	}
	for _, balance := range balances {
		selectedBalance, err := account.SelectBalanceWithID(db, balance.ID)
		errorIfError(t, err, "Expected nil error but recieved error")
		errorIfError(t, err, fmt.Sprintf("Unexpected Balance returned.\n\tExpected: %s\n\tActual  : %s", balance, selectedBalance))
	}
}

func checkAccountsSortedByIdAscending(accounts GOHMoneyDB.Accounts, t *testing.T) {
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

func TestAccount_UpdateAccount(t *testing.T) {
	now := time.Now()
	original, err := account.New("TEST_ACCOUNT", now, gohtime.NullTime{})
	fatalIfError(t, err, "Error creating a for testing")
	updatedStart := now.AddDate(1, 0, 0)
	updatedEnd := gohtime.NullTime{Valid: true, Time: updatedStart.AddDate(2, 0, 0)}
	update, err := account.New("TEST_ACCOUNT_UPDATED", updatedStart, updatedEnd)
	fatalIfError(t, err, "Error creating a for testing")
	db := prepareTestDB(t)
	defer close(t, db)
	a, err := GOHMoneyDB.CreateAccount(db, original)
	fatalIfError(t, err, "Error creating Account")
	updated, err := a.Update(db, update)
	errorIfError(t, err, "Error updating account")
	expected, err := account.New(
		update.Name,
		update.Start().Truncate(24*time.Hour),
		gohtime.NullTime{
			Valid: update.End().Valid,
			Time:  update.End().Time.Truncate(24 * time.Hour),
		},
	)
	fatalIfError(t, err, "Error creating expected account for testing")
	if !updated.Account.Equal(expected) {
		t.Errorf("Updates not applied as expected.\nUpdated a: %s\nApplied updates: %s", updated, expected)
	}
}

func TestAccount_Delete(t *testing.T) {
	invalid := GOHMoneyDB.Account{}
	if invalid.Delete(nil) == nil {
		t.Errorf("Expected error but none was returned when attempting to delete an invalid account.")
	}
	db := prepareTestDB(t)
	defer close(t, db)
	account := newTestDBAccount(t, db)
	vErr := account.Validate(db)
	fatalIfError(t, vErr, "Invalid account returned for testing")
	err := account.Delete(db)
	fatalIfError(t, err, "Error occured whilst deleting account")
	valid := account.Validate(db)
	if valid == nil {
		t.Fatalf("Account still valid after deletion.")
	} else if valid != GOHMoneyDB.ErrAccountDeleted {
		t.Fatalf("Validity error not as expected. Expected %s, got %s.", GOHMoneyDB.ErrAccountDeleted, valid)
	}
}

func TestAccount_JsonLoop(t *testing.T) {
	innerAccount, err := account.New(
		"TEST",
		time.Now(),
		gohtime.NullTime{
			Valid: true,
			Time:  time.Now().AddDate(1, 0, 0),
		},
	)
	fatalIfError(t, err, "Error creating new account for testing")
	db := prepareTestDB(t)
	defer close(t, db)
	originalAccount, err := GOHMoneyDB.CreateAccount(db, innerAccount)
	fatalIfError(t, err, "Error creating DB account for testing")
	originalBytes, err := json.Marshal(originalAccount)
	fatalIfError(t, err, "Error marshalling account into json")
	var finalAccount GOHMoneyDB.Account
	err = json.Unmarshal(originalBytes, &finalAccount)
	fatalIfError(t, err, "Error unmarshalling account")
	logBytes := func(t *testing.T) { t.Log("Marshalled account: " + string(originalBytes)) }
	if finalAccount.ID != originalAccount.ID {
		t.Errorf("Unexpected account id.\n\tExpected: %d\n\tActuall  : %d", originalAccount.ID, finalAccount.ID)
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

func TestAccounts_JSONLoop(t *testing.T) {
	var innerAccounts account.Accounts
	numOfAccounts := 100
	for i := 0; i < numOfAccounts; i++ {
		innerAccount, err := account.New(
			"TEST",
			time.Now(),
			gohtime.NullTime{
				Valid: true,
				Time:  time.Now().AddDate(1, 0, 0),
			},
		)
		fatalIfError(t, err, "Error creating new account for testing. Error: %s")
		innerAccounts = append(innerAccounts, innerAccount)
	}
	db := prepareTestDB(t)
	defer close(t, db)
	var originalAccounts GOHMoneyDB.Accounts
	for i := 0; i < len(innerAccounts); i++ {
		originalAccount, err := GOHMoneyDB.CreateAccount(db, innerAccounts[i])
		fatalIfError(t, err, "Error creating DB account for testing")
		originalAccounts = append(originalAccounts, *originalAccount)
	}
	originalBytes, err := json.Marshal(originalAccounts)
	fatalIfError(t, err, "Error marshalling account into json")
	var finalAccounts GOHMoneyDB.Accounts
	err = json.Unmarshal(originalBytes, &finalAccounts)
	fatalIfError(t, err, "Error unmarshalling accounts json")
	logBytes := func(t *testing.T) { t.Log("Marshalled accounts: " + string(originalBytes)) }
	for i := 0; i < len(innerAccounts); i++ {
		final := finalAccounts[i]
		original := originalAccounts[i]
		equal, err := final.Equal(original)
		errorIfError(t, err, "Error comparing accounts")
		if !equal {
			t.Errorf("Unexpected account.\n\tExpected: %+v\n\tActuall  : %+v", original, final)
			logBytes(t)
			// FailNow here as logging the bytes for each loop iteration can cause an extremely long output.
			t.FailNow()
		}
	}
}

func TestAccount_Validate(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)
	invalid := GOHMoneyDB.Account{}
	err := invalid.Validate(db)
	if err == nil {
		t.Errorf("Expected expected but none returned.")
	}
	if expected := GOHMoneyDB.NoAccountWithIDError(0); err != expected {
		t.Errorf("Expected error %s, but got %s", expected, err)
	}
	invalid.ID = 5
	err = invalid.Validate(db)
	if expected := GOHMoneyDB.ErrAccountDifferentInDbAndRuntime; err != expected {
		t.Errorf("Expected error %s, but got %s", expected, err)
	}

	valid, err := GOHMoneyDB.SelectAccountWithID(db, 1)
	fatalIfError(t, err, "Error selecting valid account for testing")
	validErr := valid.Validate(db)
	errorIfError(t, validErr, "Expected nil error but got")
}

func TestAccount_Equal(t *testing.T) {
	db := prepareTestDB(t)
	defer close(t, db)

	a := newTestDBAccount(t, db)
	b := a
	assertFunc := func(expected bool) {
		equal, err := a.Equal(b)
		errorIfError(t, err, "Error comparing accounts")
		assert.Equal(t, expected, equal)
	}

	b.Name = "B"
	assertFunc(false)

	b.Name = a.Name
	assertFunc(true)

	b.ID++
	assertFunc(false)

	b.ID = a.ID
	assertFunc(true)

	fatalIfError(t, a.Delete(db), "Error deleting account")
	equal, err := a.Equal(b)
	assert.NotNil(t, err)
	assert.Equal(t, false, equal)
}

func newTestAccount() account.Account {
	account, err := account.New(
		"TEST_ACCOUNT",
		time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
		gohtime.NullTime{
			Valid: true,
			Time:  time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC),
		},
	)
	if err != nil {
		panic(err)
	}
	return account
}

func newTestDBAccount(t *testing.T, db *sql.DB) GOHMoneyDB.Account {
	account, err := GOHMoneyDB.CreateAccount(db, newTestAccount())
	fatalIfError(t, err, "Error creating account for testing")
	return *account
}

func fatalIfError(t *testing.T, err error, message string) {
	if err != nil {
		t.Fatalf("%s: %s", message, err)
	}
}

func errorIfError(t *testing.T, err error, message string) {
	if err != nil {
		t.Errorf("%s: %s", message, err)
	}
}
