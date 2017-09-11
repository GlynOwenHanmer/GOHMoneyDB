package GOHMoneyDB

import (
	"testing"
	"github.com/GlynOwenHanmer/GOHMoney"
	"time"
	"bytes"
	"fmt"
	"database/sql"
	"errors"
	"github.com/lib/pq"
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
	expectedId := uint(1)
	actualId := balances[0].Id
	if expectedId != actualId {
		t.Errorf(`Unexpected Balance Id.\nExpected: %d\nActual:  %d`, expectedId, actualId)
	}
	expectedAmount := float32(636.42)
	actualAmount := balances[0].Amount
	if actualAmount != expectedAmount {
		t.Errorf("account Id: %d, first balance, expected balance amount of %f but got %f", validId, expectedAmount, actualAmount)
	}
	expectedDate, err := parseDateString("2016-06-17")
	if err != nil {
		t.Fatalf("Error parsing date string for use in tests. Error: %s", err.Error())
	}
	actualDate := balances[0].Date
	if !expectedDate.Equal(actualDate) {
		t.Errorf("account Id: %d, first balance, expected date of %s but got %s", validId, expectedDate, actualDate)
	}
}

func Test_BalanceInsert(t *testing.T) {
	accountId :=uint(1)
	db, err := prepareTestDB()
	defer db.Close()
	if err != nil {
		t.Fatalf("Unable to open DB connection. Error: %s", err)
	}
	initialLastId := getHighestBalanceId(db, t)
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
	if insertedBalance.Id != 0 {
		t.Errorf("Expected uninitialised Balance id of 0, got %d", insertedBalance.Id)
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
	validDate := time.Date(3000, 6, 1, 1, 1, 1, 1, time.UTC)
	validBalance := GOHMoney.Balance{Date: validDate, Amount: 1234.56}
	startingBalances, err = dbAccount.Balances(db)
	if err != nil {
		t.Fatalf("Unable to get balances for testing for account: %s", dbAccount)
	}
	insertedBalance, err = dbAccount.InsertBalance(db, validBalance)
	if err != nil {
		t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s", nil, err)
	}
	if insertedBalance.Id != initialLastId+1 {
		t.Errorf("Expected Id to incremement by 1.\nInitial last Id: %d\nInserted Balance Id: %d", initialLastId, insertedBalance.Id)
	}
	if !insertedBalance.Date.Equal(validDate.Truncate(time.Hour * 24)) {
		t.Errorf("Inserted balance date should be %s but is %s", validDate, insertedBalance.Date.String())
	}
	if insertedBalance.Amount != validBalance.Amount {
		t.Errorf("Inserted balance amount should be %f but is %f", validBalance.Amount, insertedBalance.Amount)
	}
	err = dbAccount.ValidateBalance(db, insertedBalance)
	if err != nil {
		t.Errorf("Expected inserted balance to be valid against account.\nError: %s\nAccount: %s\nBalance: %s", err, dbAccount, insertedBalance)
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

func getHighestBalanceId(db *sql.DB, t *testing.T) uint {
	initialLastId := uint(0)
	allAccounts, err := SelectAccounts(db)
	if err != nil {
		t.Fatalf("Error selecting all Balances for testing: %s", err.Error())
	}
	for _, account := range *allAccounts {
		balances, err := selectBalancesForAccount(db, account.Id)
		if err != nil {
			t.Fatalf("Error selecting balances for testing: %s", err.Error())
		}
		for _, balance := range balances {
			if balance.Id > initialLastId {
				initialLastId = balance.Id
			}
		}
	}
	return initialLastId
}

func TestAccount_ValidateBalance(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error when prepping test DB. Error: %s", err.Error())
	}
	account, err := CreateAccount(db,newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	validBalance, err := account.InsertBalance(db, GOHMoney.Balance{Date:account.Start().AddDate(0,0,1)})
	if err != nil {
		t.Fatalf(`Error inserting new balance for testing. Error :%s`, err)
	}
	outOfDateRange := Balance{
		Id:account.Id,
		Balance:GOHMoney.Balance{
			Date:time.Date(1,1,1,1,1,1,1,time.UTC),
		},
	}
	balanceWithWrongOwner := Balance{
		Id:account.Id-1,
		Balance:validBalance.Balance,
	}
	testSets := []struct{
		account *Account
		balance Balance
		error
	}{
		{
			account:account,
			balance:validBalance,
			error:nil,
		},
		{
			account:account,
			balance:outOfDateRange,
			error:account.Account.ValidateBalance(outOfDateRange.Balance),
		},
		{
			account:account,
			balance:balanceWithWrongOwner,
			error:InvalidAccountBalanceError{AccountId:account.Id, BalanceId:balanceWithWrongOwner.Id},
		},
	}
	for _, testSet := range testSets {
		err := testSet.account.ValidateBalance(db, testSet.balance)
		if err != testSet.error {
			t.Fatalf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", testSet.error, err)
		}
	}
}

func Test_UpdateBalance_WrongAccount(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error when prepping test DB. Error: %s", err.Error())
	}
	account0, err := CreateAccount(db,newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	account1, err := CreateAccount(db,newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := GOHMoney.Balance{
		Date:account0.Start(),
		Amount:0,
	}
	createdBalance0, err := account0.InsertBalance(db,newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := GOHMoney.Balance{
		Date:time.Now(),
		Amount:100,
	}
	updatedBalance, err := account1.UpdateBalance(db, createdBalance0,update)
	expectedError := InvalidAccountBalanceError{AccountId:account1.Id, BalanceId:createdBalance0.Id}
	if err != expectedError {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err)
	}
	expectedBalance := Balance{}
	if updatedBalance != expectedBalance {
		t.Errorf("Unexpected Balance.\n\tExpected: %s\n\tActual  : %s", expectedBalance, updatedBalance)
	}
}

func Test_UpdateBalance_InvalidUpdate(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error when prepping test DB. Error: %s", err.Error())
	}
	account, err := CreateAccount(db,newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := GOHMoney.Balance{ Date: account.Start() }
	createdBalance, err := account.InsertBalance(db,newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := GOHMoney.Balance{}
	updatedBalance, err := account.UpdateBalance(db, createdBalance,update)
	expectedError := errors.New(`Update Balance is not valid: ` + GOHMoney.BalanceZeroDate.Error())
	if err.Error() != expectedError.Error() {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err)
	}
	expectedBalance := Balance{}
	if updatedBalance != expectedBalance {
		t.Errorf("Unexpected Balance.\n\tExpected: %s\n\tActual  : %s", expectedBalance, updatedBalance)
	}
}

// Test for when the update Balance that is trying to be applied is not valid in the context of the account. For example, where the Data of the update is outside of the TimeRange of the account.
func Test_UpdateBalance_InvalidUpdateForAccount(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error when prepping test DB. Error: %s", err.Error())
	}
	account, err := CreateAccount(db,newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := GOHMoney.Balance{ Date: account.Start() }
	createdBalance, err := account.InsertBalance(db,newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := GOHMoney.Balance{Date:account.Start().AddDate(-1,0,0)}
	_, err = account.UpdateBalance(db, createdBalance,update)
	expectedError := `Update is not valid for account: ` + GOHMoney.BalanceDateOutOfAccountTimeRange{}.Error()
	if err == nil {
		t.Errorf("Expected error but got nil.")
	} else if err.Error() != expectedError {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err.Error())
	}
}

func Test_UpdateBalance_ValidBalance(t *testing.T) {
	db, err := prepareTestDB()
	if err != nil {
		t.Fatalf("Error when prepping test DB. Error: %s", err.Error())
	}
	account, err := CreateAccount(db,newTestAccount())
	if err != nil {
		t.Fatalf(`Error creating new account for testing. Error: %s`, err)
	}
	newBalance := GOHMoney.Balance{	Date: account.Start() }
	createdBalance, err := account.InsertBalance(db,newBalance)
	if err != nil {
		t.Fatalf(`Error creating inserting new Balance into DB for testing. Error: %s`, err.Error())
	}
	update := GOHMoney.Balance{
		Date:account.Start().AddDate(0,0,1),
		Amount:100,
	}
	updatedBalance, err := account.UpdateBalance(db, createdBalance,update)
	expectedError := error(nil)
	if err != expectedError {
		t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", expectedError, err)
	}
	if updatedBalance.Id != updatedBalance.Id {
		t.Errorf("Balance Id changed when updating Balance\n\tOriginal: %d\n\tFinal   : %d", createdBalance.Id, updatedBalance.Id)
	}
	expectedDate := update.Date.Truncate(time.Hour * 24)
	if !updatedBalance.Balance.Date.Equal(expectedDate) {
		t.Errorf("Unexpected Balance date.\n\tExpected: %s\n\tActual  : %s", update.Date, updatedBalance.Balance.Date)
	}
	if updatedBalance.Balance.Amount != update.Amount {
		t.Errorf("Unexpected Balance Amount.\n\tExpected: %s\n\tActual  : %s", update.Amount, updatedBalance.Amount)
	}
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
			message := fmt.Sprintf("Unexpected balance.\nExpected: %s\nActual  : %s", testSet.expectedBalance, balance)
			message += fmt.Sprintf("\nFor time: %v", testSet.Time)
			t.Error(message)
		}
	}
}

func newTestAccount() GOHMoney.Account {
	account, err := GOHMoney.NewAccount(
		"TEST_ACCOUNT",
		time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
		pq.NullTime{
			Valid: true,
			Time: time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC),
		},
	)
	if err != nil {
		panic(err)
	}
	return account
}
