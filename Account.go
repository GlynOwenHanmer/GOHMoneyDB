
package GOHMoneyDB

import (
	"database/sql"
	"github.com/GlynOwenHanmer/GOHMoney"
	_ "github.com/lib/pq"
	"fmt"
	"bytes"
	"github.com/lib/pq"
	"time"
	"encoding/json"
	"errors"
)

const (
	insertFields = "name, date_opened, date_closed"
	selectFields = "id, name, date_opened, date_closed"
)

// Account holds logic for an Account item that is held within a GOHMoney database.
type Account struct {
	Id         uint
	GOHMoney.Account
}

// accountJsonHelper is purely used as a helper struct to marshal and unmarshal Account objects to and from json bytes
type accountJsonHelper struct {
	Id uint
	Name string
	Start time.Time
	End pq.NullTime
}

// MarshalJSON Marshals an Account into json bytes and an error
func (account Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(&accountJsonHelper{
		Id: account.Id,
		Name: account.Name,
		Start:account.Start(),
		End:account.End(),
	})
}

// UnmarshalJSON attempts to unmarshal a json blob into an Account object and returns any errors with the unmarshalling or unmarshalled account.
func (account *Account) UnmarshalJSON(data []byte) error {
	var helper accountJsonHelper
	if err := json.Unmarshal(data, &helper); err != nil {
		return err
	}
	innerAccount, err := GOHMoney.NewAccount(helper.Name, helper.Start, helper.End)
	if err != nil {
		return err
	}
	account.Id = helper.Id
	account.Account = innerAccount
	var returnErr error
	if err := account.Validate(); err != nil {
		returnErr = err
	}
	return returnErr
}

// Accounts holds multiple Account items.
type Accounts []Account

// ValidateBalance validates a Balance against an Account and returns any errors that are encountered along the way.
// ValidateBalance will return any error that is present with the Balance itself, the Balance's Date in reference to the Account's TimeRange and also check that the Account is the valid owner of the Balance.
func (account Account) ValidateBalance(db *sql.DB, balance Balance) error {
	err := account.Account.ValidateBalance(balance.Balance)
	if err != nil {
		return err
	}
	balances, err := selectBalancesForAccount(db, account.Id)
	if err != nil {
		return err
	}
	for _, accountBalance := range balances {
		if accountBalance.Id == balance.Id {
			return nil
		}
	}
	return InvalidAccountBalanceError{
		AccountId:account.Id,
		BalanceId:balance.Id,
	}
}

// SelectAccounts returns an Accounts item holding all Account entries within the given database along with any errors occured whilst attempting to retrieve the Accounts.
func SelectAccounts(db *sql.DB) (*Accounts, error) {
	queryString := "SELECT " + selectFields + " FROM accounts ORDER BY id ASC;"
	rows, err := db.Query(queryString)
	if err != nil {
		return &Accounts{}, err
	}
	defer rows.Close()
	return scanRowsForAccounts(rows)
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
	if account == nil {
		account = &Account{}
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
	row := db.QueryRow(queryString.String(), newAccount.Name, newAccount.Start(), newAccount.End())
	return scanRowForAccount(row)
}

// SelectBalanceWithId returns a Balance from the database that has the given ID if the account is the correct one that it belongs to.
// Otherwise, SelectBalanceWithId returns an empty Balance object and an error.
func (account Account) SelectBalanceWithId(db *sql.DB, id uint) (*Balance, error) {
	var query bytes.Buffer
	fmt.Fprintf(&query, `SELECT %s FROM balances b `, balanceSelectFields)
	fmt.Fprint(&query, `WHERE b.account_id = $1 AND b.id = $2`)
	row := db.QueryRow(query.String(), account.Id, id)
	return scanRowForBalance(row)
}

// scanRowsForAccounts scans an sql.Rows object for GOHMoneyDB.Accounts objects and returns then along with any error that occurs whilst attempting to scan.
func scanRowsForAccounts(rows *sql.Rows) (*Accounts, error) {
	openAccounts := Accounts{}
	for rows.Next() {
		var id uint
		var name string
		var start time.Time
		var end pq.NullTime
		err := rows.Scan(&id, &name, &start, &end)
		if err != nil {
			return nil, err
		}
		innerAccount, err := GOHMoney.NewAccount(name, start, end)
		if err != nil {
			return nil, err
		}
		openAccounts = append(openAccounts, Account{Id: id,	Account:innerAccount})
	}
	return &openAccounts, rows.Err()
}

// scanRowForAccount scans a single sql.Row for a Account object and returns any error occurring along the way.
func scanRowForAccount(row *sql.Row) (*Account, error) {
	var id uint
	var name string
	var start time.Time
	var end pq.NullTime
	err := row.Scan(&id, &name, &start, &end)
	if err != nil {
		return nil, err
	}
	innerAccount, err := GOHMoney.NewAccount(name, start, end)
	if err != nil{
		return nil ,err
	}
	return &Account{Id:id, Account:innerAccount}, nil
}

// UpdateBalance updates an Account entry in a given db, returning any errors that are present with the validity of the original Account or update Account.
func (original Account) Update(db *sql.DB, update GOHMoney.Account) (Account, error) {
	if err := original.Validate(); err != nil {
		return Account{}, err
	}
	if err := update.Validate(); err != nil {
		return Account{}, errors.New(`Update Account is not valid: ` + err.Error())
	}
	row := db.QueryRow(`UPDATE accounts SET name = $1, date_opened = $2, date_closed = $3 WHERE id = $4 returning ` + selectFields, update.Name, update.Start(), update.End(), original.Id)
	account, err := scanRowForAccount(row)
	return *account, err
}