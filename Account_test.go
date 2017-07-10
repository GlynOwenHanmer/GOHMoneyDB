package GOHMoneyDB

import (
	"testing"
	"github.com/GlynOwenHanmer/GOHMoney"
	"github.com/lib/pq"
	"time"
	"fmt"
	"bytes"
)

func Test_SelectAccounts(t *testing.T) {
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	accounts, err := SelectAccounts(db)
	if err != nil {
		t.Fatalf("Error running SelectAccounts method: %s", err)
	}
	checkAccountsSortedByIdAscending(accounts, t)
}

func Test_SelectAccountsOpen(t *testing.T) {
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	openAccounts, err := SelectAccountsOpen(db)
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

func checkAccountsSortedByIdAscending(accounts Accounts, t *testing.T) {
	for i := 0; i+1 < len(accounts); i++ {
		account := accounts[i]
		nextAccount := accounts[i+1]
		switch {
		case account.Id > nextAccount.Id:
			var message bytes.Buffer
			fmt.Fprintf(&message, "Accounts not returned sorted by id. Id %d appears before %d.\n", account.Id, nextAccount.Id)
			fmt.Fprintf(&message, "accounts[%d]: %s", i, account)
			fmt.Fprintf(&message, "accounts[%d]: %s", i+1, nextAccount)
			t.Errorf(message.String())
		}
	}
}

func Test_SelectAccountWithId(t *testing.T) {
	testSets := []struct{
		id uint
		expectedError error
		expectedAccount Account
	}{
		{
			id:0,
			expectedError:NoAccountWithIdError(0),
			expectedAccount:Account{},
		},
		{
			id:9999999,
			expectedError:NoAccountWithIdError(9999999),
			expectedAccount:Account{},
		},
		{
			id:10,
			expectedError:nil,
			expectedAccount:Account{
				Id:10,
				Account:GOHMoney.Account{Name:"Ikaros"}},
		},
		{
			id:20,
			expectedError:nil,
			expectedAccount:Account{
				Id: 20,
				Account:GOHMoney.Account{Name:"Amsterdam"}},
		},
	}
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	for _, testSet := range testSets {
		if err != nil {
			t.Fatalf("Unable to open DB connection.\n%s", err)
		}
		account, err := SelectAccountWithID(db, testSet.id)
		if testSet.expectedError != err {
			t.Errorf("Unexpected errors\nExpected: %v\nActual  : %v", testSet.expectedError, err)
		}
		if testSet.expectedAccount.Id != account.Id {
			t.Errorf("Unexpected Account id\nExpected: %d\nActual  : %d", testSet.expectedAccount.Id, account.Id)
		}
		if testSet.expectedAccount.Name != account.Name {
			t.Errorf("Unexpected Account name\nExpected: %s\nActual  : %s", testSet.expectedAccount.Name, account.Name)
		}
	}
}

func Test_checkNewAccount(t *testing.T) {
	testSets := []struct{
		insertedAccount GOHMoney.Account
		NewAccountFieldError
	}{
		{
			insertedAccount: GOHMoney.Account{},
			NewAccountFieldError: NewAccountFieldError{EmptyNameError, ZeroDateOpenedError},
		},
		{
			insertedAccount: GOHMoney.Account{},
			NewAccountFieldError: NewAccountFieldError{EmptyNameError, ZeroDateOpenedError},
		},
		{
			insertedAccount: GOHMoney.Account{
				Name:"TEST_ACCOUNT",
			},
			NewAccountFieldError: NewAccountFieldError{ZeroDateOpenedError},
		},
		{
			insertedAccount: GOHMoney.Account{
				DateOpened: time.Date(2000,1,1,1,1,1,1,time.UTC),
			},
			NewAccountFieldError: NewAccountFieldError{EmptyNameError},
		},
		{
			insertedAccount: GOHMoney.Account{
				Name:"TEST_ACCOUNT",
				DateClosed: pq.NullTime{Valid:true},
			},
			NewAccountFieldError: NewAccountFieldError{ZeroDateOpenedError,ZeroValidDateClosedError},
		},
		{
			insertedAccount: GOHMoney.Account{
				Name:"TEST_ACCOUNT",
				DateOpened: time.Date(2000,1,1,1,1,1,1,time.UTC),
				DateClosed: pq.NullTime{Valid:true, Time:time.Date(1999,1,1,1,1,1,1,time.UTC)},
			},
			NewAccountFieldError: NewAccountFieldError{DateClosedBeforeDateOpenedError},
		},
		{
			insertedAccount: newTestAccount(),
			NewAccountFieldError: nil,
		},
	}
	for _, testSet := range testSets {
		actual := checkNewAccount(testSet.insertedAccount)
		expected := testSet.NewAccountFieldError
		if !stringSlicesMatch(expected, actual) {
			t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s\nInserted Account: %s", expected, actual, testSet.insertedAccount)
		}
	}
}

func Test_CreateAccount(t *testing.T) {
	testSets := []struct{
		insertedAccount, createdAccount GOHMoney.Account
		error
	}{
		{
			insertedAccount: GOHMoney.Account{},
			createdAccount: GOHMoney.Account{},
			error:          NewAccountFieldError{},
		},
		{
			insertedAccount: newTestAccount(),
			createdAccount: newTestAccount(),
			error:          nil,
		},
		{
			insertedAccount: GOHMoney.Account{
				Name:"TEST_ACCOUNT",
				DateOpened:time.Now(),
				DateClosed:pq.NullTime{Valid:false},
			},
			createdAccount: GOHMoney.Account{
				Name:"TEST_ACCOUNT",
				DateOpened:time.Now(),
				DateClosed:pq.NullTime{Valid:false},
			},
			error:          nil,
		},
		{
			insertedAccount: GOHMoney.Account{
				Name:"Account With'Apostrophe",
				DateOpened:time.Now(),
				DateClosed:pq.NullTime{Valid:false},
			},
			createdAccount: GOHMoney.Account{
				Name:"Account With'Apostrophe",
				DateOpened:time.Now(),
				DateClosed:pq.NullTime{Valid:false},
			},
			error:          nil,
		},
	}
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection.: %s", err)
	}
	for _, testSet := range testSets {
		newAccount := testSet.insertedAccount
		actualCreatedAccount, err := CreateAccount(db, newAccount)
		if testSet.error == nil && err != nil || testSet.error != nil && err == nil {
			t.Errorf("Unexpected error:\nExpected: %s\nActual  : %s", testSet.error, err)
		}
		if _, testSetErrIsNewAccountFieldError := testSet.error.(NewAccountFieldError); testSetErrIsNewAccountFieldError {
			if _, actualErrorIsNewAccountFieldError := err.(NewAccountFieldError); !actualErrorIsNewAccountFieldError {
				t.Errorf("Unexpected error:\nExpected: %s\nActual  : %s", testSet.error, err)
			}
		}

		if testSet.createdAccount.Name != actualCreatedAccount.Name {
			t.Errorf("Unexpected created account name:\nExpected: %s\nActual  : %s", testSet.createdAccount.Name, actualCreatedAccount.Name)
		}
	}
}

func stringSlicesMatch(array1, array2 []string) bool {
	if len(array1) != len(array2) {
		return false
	}
	for i := 0; i < len(array1); i++ {
		if array1[i] != array2[i] {
			return false
		}
	}
	return true
}

func newTestAccount() GOHMoney.Account {
	return GOHMoney.Account{
		Name:       "TEST_ACCOUNT",
		DateOpened: time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
		DateClosed: pq.NullTime{Valid: true, Time: time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC)},
	}
}

