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
			fmt.Fprintf(&message, "Accounts not returned sorted by Id. Id %d appears before %d.\n", account.Id, nextAccount.Id)
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
			t.Errorf("Unexpected Account Id\nExpected: %d\nActual  : %d", testSet.expectedAccount.Id, account.Id)
		}
		if testSet.expectedAccount.Name != account.Name {
			t.Errorf("Unexpected Account name\nExpected: %s\nActual  : %s", testSet.expectedAccount.Name, account.Name)
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
			error:          GOHMoney.AccountFieldError{},
		},
		{
			insertedAccount: newTestAccount(),
			createdAccount: newTestAccount(),
			error:          nil,
		},
		{
			insertedAccount: GOHMoney.Account{
				Name:"TEST_ACCOUNT",
				TimeRange:GOHMoney.TimeRange{
					Start:pq.NullTime{
						Valid:true,
						Time:time.Now(),
					},
					End:pq.NullTime{
						Valid:false,
					},
				},
			},
			createdAccount: GOHMoney.Account{
				Name:"TEST_ACCOUNT",
				TimeRange:GOHMoney.TimeRange{
					Start:pq.NullTime{
						Valid:true,
						Time:time.Now(),
					},
					End:pq.NullTime{
						Valid:false,
					},
				},
			},
			error:          nil,
		},
		{
			insertedAccount: GOHMoney.Account{
				Name:"Account With'Apostrophe",
				TimeRange:GOHMoney.TimeRange{
					Start:pq.NullTime{
						Valid:true,
						Time:time.Now(),
					},
					End:pq.NullTime{
						Valid:false,
					},
				},
			},
			createdAccount: GOHMoney.Account{
				Name:"Account With'Apostrophe",
				TimeRange:GOHMoney.TimeRange{
					Start:pq.NullTime{
						Valid:true,
						Time:time.Now(),
					},
					End:pq.NullTime{
						Valid:false,
					},
				},
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
		if _, testSetErrIsNewAccountFieldError := testSet.error.(GOHMoney.AccountFieldError); testSetErrIsNewAccountFieldError {
			if _, actualErrorIsNewAccountFieldError := err.(GOHMoney.AccountFieldError); !actualErrorIsNewAccountFieldError {
				t.Errorf("Unexpected error:\nExpected: %s\nActual  : %s", testSet.error, err)
			}
		}

		if testSet.createdAccount.Name != actualCreatedAccount.Name {
			t.Errorf("Unexpected created account name:\nExpected: %s\nActual  : %s", testSet.createdAccount.Name, actualCreatedAccount.Name)
		}
	}
}

func TestAccount_SelectBalanceWithID_InvalidID(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Unable to prepare DB for testings. Error: %s", err.Error())
	}
	account, err := CreateAccount(db, newTestAccount())
	// Account with no Balances
	_, err = account.SelectBalanceWithId(db, 10)
	expectedErr := NoBalances
	if err != expectedErr {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedErr, err)
	}

	validBalance, err := account.InsertBalance(db,
		GOHMoney.Balance{
			Date:   account.Start.Time.AddDate(0, 0, 10),
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
	account, err := CreateAccount(db, newTestAccount())
	var balances [3]Balance
	for i := 0; i < 3; i++ {
		balances[i], err =  account.InsertBalance(db,
			GOHMoney.Balance{
				Date:account.Start.Time.AddDate(0,0,i),
				Amount:float32(i),
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
			t.Errorf("Unexpected Balance returned.\n\tExpected: %s\n\tActual  : %s",balance, selectedBalance)
		}
	}
}

func newTestAccount() GOHMoney.Account {
	return GOHMoney.Account{
		Name:       "TEST_ACCOUNT",
		TimeRange:GOHMoney.TimeRange{
			Start:pq.NullTime{
				Valid: true,
				Time: time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			End:pq.NullTime{
				Valid: true,
				Time: time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC),
			},
		},
	}
}

