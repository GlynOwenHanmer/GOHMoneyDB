
package GOHMoneyDB

import (
	"database/sql"
	"github.com/GlynOwenHanmer/GOHMoney"
	_ "github.com/lib/pq"
	"fmt"
	"bytes"
	"strings"
)

const (
	insertFields = "name, date_opened, date_closed"
	selectFields = "id, name, date_opened, date_closed"
)

// Account holds logic for an Account item that is held within a GOHMoney database.
type Account struct {
	Id         uint		`json:"id"`
	GOHMoney.Account
}

// Accounts holds multiple Account items.
type Accounts []Account

// SelectAccounts returns an Accounts item holding all Account entries within the given database along with any errors occured whilst attempting to retrieve the Accounts.
func SelectAccounts(db *sql.DB) (Accounts, error) {
	queryString := "SELECT " + selectFields + " FROM accounts ORDER BY id ASC;"
	rows, err := db.Query(queryString)
	if err != nil {
		return Accounts{}, err
	}
	defer rows.Close()
	accounts := Accounts{}
	for rows.Next() {
		account := Account{}
		err = rows.Scan(&account.Id, &account.Name, &account.DateOpened, &account.DateClosed)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	err = rows.Err()
	return accounts, err
}

// SelectAccountsOpen returns an Accounts item holding all Account entries within the given database that are open along with any errors occured whilst attempting to retrieve the Accounts.
func SelectAccountsOpen(db *sql.DB) (Accounts, error) {
	queryString := "SELECT " + selectFields + " FROM accounts WHERE date_closed IS NULL ORDER BY id ASC;"
	rows, err := db.Query(queryString)
	if err != nil {
		return Accounts{}, err
	}
	defer rows.Close()
	openAccounts := Accounts{}
	for rows.Next() {
		account := Account{}
		err = rows.Scan(&account.Id, &account.Name, &account.DateOpened, &account.DateClosed)
		if err != nil {
			return nil, err
		}
		openAccounts = append(openAccounts, account)
	}
	err = rows.Err()
	return openAccounts, err
}

// SelectAccountWithId returns the Account from the DB with the given id value along with any error that occurs whilst attempting to retrieve the Account.
func SelectAccountWithID(db *sql.DB, id uint) (Account, error) {
	queryString := fmt.Sprintf("SELECT " + selectFields + " FROM accounts WHERE id = %d;", id)
	row := db.QueryRow(queryString)
	account := Account{}
	err := row.Scan(&account.Id, &account.Name, &account.DateOpened, &account.DateClosed)
	if err == sql.ErrNoRows {
		err = NoAccountWithIdError(id)
	}
	return account, err
}

// CreateAccount created an Account entry within the DB and returns it, if successful, along with any errors that occur whilst attempting to create the Account.
func CreateAccount(db *sql.DB, newAccount GOHMoney.Account) (Account, error) {
	newAccountFieldErrors := checkNewAccount(newAccount)
	if newAccountFieldErrors != nil {
		return Account{}, newAccountFieldErrors
	}
	var queryString bytes.Buffer
	fmt.Fprintf(&queryString, `INSERT INTO accounts (%s) `, insertFields)
	fmt.Fprint(&queryString, `VALUES ($1, $2, $3) `)
	fmt.Fprintf(&queryString, `returning %s`, selectFields)
	row := db.QueryRow(queryString.String(), newAccount.Name, newAccount.DateOpened, newAccount.DateClosed)
	createdAccount := Account{}
	err := row.Scan(&createdAccount.Id, &createdAccount.Name, &createdAccount.DateOpened, &createdAccount.DateClosed)
	return createdAccount, err
}

// checkNewAccount checks the state of an account to see if it is eligible to create a new account from. checkNewAccount returns a set of errors representing errors with different fields of te new account.
func checkNewAccount(newAccount GOHMoney.Account) NewAccountFieldError {
	var fieldErrorDescriptions []string
	if len(strings.TrimSpace(newAccount.Name)) == 0 {
		fieldErrorDescriptions = append(fieldErrorDescriptions, EmptyNameError)
	}
	if newAccount.DateOpened.IsZero() {
		fieldErrorDescriptions = append(fieldErrorDescriptions, ZeroDateOpenedError)
	}
	if newAccount.DateClosed.Valid {
		if newAccount.DateClosed.Time.IsZero() {
			fieldErrorDescriptions = append(fieldErrorDescriptions, ZeroValidDateClosedError)
		} else if newAccount.DateClosed.Time.Before(newAccount.DateOpened) {
			fieldErrorDescriptions = append(fieldErrorDescriptions, DateClosedBeforeDateOpenedError)
		}
	}
	return NewAccountFieldError(fieldErrorDescriptions)
}