package GOHMoneyDB

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/GlynOwenHanmer/GOHMoney"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

const (
	insertFields = "name, date_opened, date_closed"
	selectFields = "id, name, date_opened, date_closed, deleted_at"
)

// Account holds logic for an Account item that is held within a GOHMoney database.
type Account struct {
	Id uint
	GOHMoney.Account
	deletedAt GOHMoney.NullTime
}

// accountJsonHelper is purely used as a helper struct to marshal and unmarshal Account objects to and from json bytes
type accountJsonHelper struct {
	Id    uint
	Name  string
	Start time.Time
	End   GOHMoney.NullTime
}

// MarshalJSON Marshals an Account into json bytes and an error
func (account Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(&accountJsonHelper{
		Id:    account.Id,
		Name:  account.Name,
		Start: account.Start(),
		End:   account.End(),
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
	if err := account.Account.Validate(); err != nil {
		return err
	}
	return nil
}

// Accounts holds multiple Account items.
type Accounts []Account

// ValidateBalance validates a Balance against an Account and returns any errors that are encountered along the way.
// ValidateBalance will return any error that is present with the Balance itself, the Balance's Date in reference to the Account's TimeRange and also check that the Account is the valid owner of the Balance.
func (account Account) ValidateBalance(db *sql.DB, balance Balance) error {
	if err := account.Validate(db); err != nil {
		return err
	}
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
		AccountId: account.Id,
		BalanceId: balance.Id,
	}
}

// SelectAccounts returns an Accounts item holding all Account entries within the given database along with any errors occured whilst attempting to retrieve the Accounts.
func SelectAccounts(db *sql.DB) (*Accounts, error) {
	queryString := "SELECT " + selectFields + " FROM accounts WHERE deleted_at IS NULL ORDER BY id ASC;"
	rows, err := db.Query(queryString)
	if err != nil {
		return &Accounts{}, err
	}
	defer rows.Close()
	return scanRowsForAccounts(rows)
}

// SelectAccountsOpen returns an Accounts item holding all Account entries within the given database that are open along with any errors occured whilst attempting to retrieve the Accounts.
func SelectAccountsOpen(db *sql.DB) (Accounts, error) {
	queryString := "SELECT " + selectFields + " FROM accounts WHERE date_closed IS NULL AND deleted_at IS NULL ORDER BY id ASC;"
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
	if db == nil {
		return Account{}, errors.New("nil db pointer")
	}
	row := db.QueryRow("SELECT "+selectFields+" FROM accounts WHERE id = $1;", id)
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
	row := db.QueryRow(queryString.String(), newAccount.Name, newAccount.Start(), pq.NullTime(newAccount.End()))
	return scanRowForAccount(row)
}

// SelectBalanceWithId returns a Balance from the database that has the given ID if the account is the correct one that it belongs to.
// Otherwise, SelectBalanceWithId returns an empty Balance object and an error.
func (a Account) SelectBalanceWithId(db *sql.DB, id uint) (*Balance, error) {
	if err := a.Validate(db); err != nil {
		return &Balance{}, err
	}
	var query bytes.Buffer
	fmt.Fprintf(&query, `SELECT %s FROM balances b JOIN accounts a ON b.account_id = a.id `, bBalanceSelectFields)
	fmt.Fprint(&query, `WHERE a.deleted_at IS NULL AND b.account_id = $1 AND b.id = $2`)
	row := db.QueryRow(query.String(), a.Id, id)
	return scanRowForBalance(row)
}

// scanRowsForAccounts scans an sql.Rows object for GOHMoneyDB.Accounts objects and returns then along with any error that occurs whilst attempting to scan.
func scanRowsForAccounts(rows *sql.Rows) (*Accounts, error) {
	openAccounts := Accounts{}
	for rows.Next() {
		var id uint
		var name string
		var start time.Time
		var end, deletedAt pq.NullTime
		err := rows.Scan(&id, &name, &start, &end, &deletedAt)
		if err != nil {
			return nil, err
		}
		innerAccount, err := GOHMoney.NewAccount(name, start, GOHMoney.NullTime(end))
		if err != nil {
			return nil, err
		}
		openAccounts = append(openAccounts, Account{Id: id, Account: innerAccount, deletedAt: GOHMoney.NullTime(deletedAt)})
	}
	return &openAccounts, rows.Err()
}

// scanRowForAccount scans a single sql.Row for a Account object and returns any error occurring along the way.
// If the account exists but has been marked as deleted, an AccountDeleted error will be returned along with the account.
func scanRowForAccount(row *sql.Row) (*Account, error) {
	var id uint
	var name string
	var start time.Time
	var end, deletedAt pq.NullTime
	if err := row.Scan(&id, &name, &start, &end, &deletedAt); err != nil {
		return nil, err
	}
	innerAccount, err := GOHMoney.NewAccount(name, start, GOHMoney.NullTime(end))
	if err != nil {
		return nil, err
	}
	if deletedAt.Valid {
		err = AccountDeleted
	}
	return &Account{Id: id, Account: innerAccount, deletedAt: GOHMoney.NullTime(deletedAt)}, err
}

// Update updates an Account entry in a given db, returning any errors that are present with the validity of the original Account or update Account.
func (original Account) Update(db *sql.DB, update GOHMoney.Account) (Account, error) {
	if err := original.Validate(db); err != nil {
		return Account{}, err
	}
	if err := update.Validate(); err != nil {
		return Account{}, errors.New(`Update Account is not valid: ` + err.Error())
	}
	balances, err := original.Balances(db)
	if err != nil {
		return Account{}, errors.New("Error selecting balances for validation: " + err.Error())
	}
	for _, b := range balances {
		if err := update.ValidateBalance(b.Balance); err != nil {
			return Account{}, errors.New(fmt.Sprintf("Update would make at least one account balance (id: %d) invalid. Error: %s", b.Id, err))
		}
	}
	account, err := scanRowForAccount(
		db.QueryRow(`UPDATE accounts SET name = $1, date_opened = $2, date_closed = $3 WHERE id = $4 returning `+selectFields, update.Name, update.Start(), pq.NullTime(update.End()), original.Id),
	)
	return *account, err
}

// Delete marks an Account as deleted in the DB, returning any errors that occur whilst attempting the deletion.
func (a *Account) Delete(db *sql.DB) error {
	if err := a.Validate(db); err != nil {
		return errors.New("Account is not valid. " + err.Error())
	}
	deletedAt := pq.NullTime{Valid: true, Time: time.Now()}
	_, err := scanRowForAccount(
		db.QueryRow(`UPDATE accounts SET deleted_at = $1 WHERE id = $2 returning `+selectFields, deletedAt, a.Id),
	)
	if err == AccountDeleted {
		a.deletedAt = GOHMoney.NullTime(deletedAt)
		return nil
	}
	if err != nil {
		return err
	}
	//Should not be reached.
	return errors.New("internal error when deleting account")
}

// Validate returns any errors that are present with an Account object
func (a Account) Validate(db *sql.DB) error {
	b, err := SelectAccountWithID(db, a.Id)
	if err != nil {
		return err
	}
	if a.deletedAt.Valid && b.deletedAt.Valid && !a.deletedAt.Time.Equal(b.deletedAt.Time) {
		return AccountDifferentInDbAndRuntime
	}
	if !a.Account.Equal(&b.Account) {
		return AccountDifferentInDbAndRuntime
	}
	if err := a.Account.Validate(); err != nil {
		return nil
	}
	return err
}
