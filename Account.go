
package GOHMoneyDB

import (
	"database/sql"
	"github.com/GlynOwenHanmer/GOHMoney"
	_ "github.com/lib/pq"
	"fmt"
	"bytes"
	"github.com/lib/pq"
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
	accounts, err := scanRowsForAccounts(rows)
	return *accounts, err
}

// SelectAccountsOpen returns an Accounts item holding all Account entries within the given database that are open along with any errors occured whilst attempting to retrieve the Accounts.
func SelectAccountsOpen(db *sql.DB) (Accounts, error) {
	queryString := "SELECT " + selectFields + " FROM accounts WHERE date_closed IS NULL ORDER BY id ASC;"
	rows, err := db.Query(queryString)
	if err != nil {
		return Accounts{}, err
	}
	defer rows.Close()
	openAccounts, err := scanRowsForAccounts(rows)
	return *openAccounts, err
}

// SelectAccountWithId returns the Account from the DB with the given Id value along with any error that occurs whilst attempting to retrieve the Account.
func SelectAccountWithID(db *sql.DB, id uint) (Account, error) {
	queryString := fmt.Sprintf("SELECT " + selectFields + " FROM accounts WHERE id = %d;", id)
	row := db.QueryRow(queryString)
	account, err := scanRowForAccount(row)
	if err == sql.ErrNoRows {
		err = NoAccountWithIdError(id)
	}
	return *account, err
}

// CreateAccount created an Account entry within the DB and returns it, if successful, along with any errors that occur whilst attempting to create the Account.
func CreateAccount(db *sql.DB, newAccount GOHMoney.Account) (*Account, error) {
	newAccountFieldErrors := newAccount.Validate()
	if newAccountFieldErrors != nil {
		return &Account{}, newAccountFieldErrors
	}
	var queryString bytes.Buffer
	fmt.Fprintf(&queryString, `INSERT INTO accounts (%s) `, insertFields)
	fmt.Fprint(&queryString, `VALUES ($1, $2, $3) `)
	fmt.Fprintf(&queryString, `returning %s`, selectFields)
	row := db.QueryRow(queryString.String(), newAccount.Name, newAccount.TimeRange.Start.Time, newAccount.TimeRange.End)
	return scanRowForAccount(row)
}

// scanRowsForAccounts scans an sql.Rows object for GOHMoneyDB.Accounts objects and returns then along with any error that occurs whilst attempting to scan.
func scanRowsForAccounts(rows *sql.Rows) (*Accounts, error) {
	openAccounts := Accounts{}
	for rows.Next() {
		account := Account{
			Account:GOHMoney.Account{
				TimeRange:GOHMoney.TimeRange{
					Start:pq.NullTime{
						Valid:true,
					},
				},
			},
		}
		err := rows.Scan(&account.Id, &account.Name, &account.TimeRange.Start.Time, &account.TimeRange.End)
		if err != nil {
			return nil, err
		}
		openAccounts = append(openAccounts, account)
	}
	return &openAccounts, rows.Err()
}

// scanRowForAccount scans a single sql.Row for a GOHMoneyDB.Account obect and returns any error occuring along the way.
func scanRowForAccount(row *sql.Row) (*Account, error) {
	account := Account{
		Account:GOHMoney.Account{
			TimeRange:GOHMoney.TimeRange{
				Start:pq.NullTime{
					Valid:true,
				},
			},
		},
	}
	return &account, row.Scan(&account.Id, &account.Name, &account.TimeRange.Start.Time, &account.TimeRange.End)
}