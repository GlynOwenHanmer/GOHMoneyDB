package postgres

import (
	"database/sql"
	"io"
	"log"
	"time"

	"errors"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-money/currency"
	"github.com/lib/pq"
)

const (
	// insertFields = "name, date_opened, date_closed"
	selectFields = "id, name, date_opened, date_closed, deleted_at"
)

// SelectAccounts returns an Accounts item holding all Account entries within the given database along with any errors that occurred whilst attempting to retrieve the Accounts.
func (pg postgres) SelectAccounts() (*storage.Accounts, error) {
	queryString := "SELECT " + selectFields + " FROM accounts WHERE deleted_at IS NULL ORDER BY id ASC;"
	rows, err := pg.db.Query(queryString)
	if err != nil {
		return nil, err
	}
	defer nonReturningClose(rows)
	return scanRowsForAccounts(rows)
}

func (pg *postgres) InsertAccount(a account.Account) (*storage.Account, error) {
	return nil, errors.New("not implemented")
}

//type AccountFilter func(*postgres) error

//func ID(ID uint) AccountFilter {
//	return func(pg *postgres) error {
//
//	}
//}

//accountJSONHelper is purely used as a helper struct to marshal and unmarshal Account objects to and from json bytes
//type accountJSONHelper struct {
//	ID    uint
//	Name  string
//	Start time.Time
//	End   gohtime.NullTime
//}

// MarshalJSON Marshals an Account into json bytes and an error
//func (a Account) MarshalJSON() ([]byte, error) {
//	return json.Marshal(&accountJSONHelper{
//		ID:    a.ID,
//		Name:  a.Name,
//		Start: a.Start(),
//		End:   a.End(),
//	})
//}

// UnmarshalJSON attempts to unmarshal a json blob into an Account object and returns any errors with the unmarshalling or unmarshalled account.
//func (a *Account) UnmarshalJSON(data []byte) (err error) {
//	var helper accountJSONHelper
//	if err = json.Unmarshal(data, &helper); err != nil {
//		return err
//	}
//	innerAccount, err := account.New(helper.Name, helper.Start, helper.End)
//	if err != nil {
//		return err
//	}
//	a.ID = helper.ID
//	a.Account = innerAccount
//	if vErr := a.Account.validate(); vErr != nil {
//		err = vErr
//	}
//	return
//}

// ValidateBalance validates a Balance against an Account and returns any errors that are encountered along the way.
// ValidateBalance will return any error that is present with the Balance itself, the Balance's Date in reference to the Account's TimeRange and also check that the Account is the valid owner of the Balance.
//func (a Account) ValidateBalance(db *sql.DB, balance Balance) error {
//	if err := a.validate(db); err != nil {
//		return err
//	}
//	err := a.Account.ValidateBalance(balance.balance)
//	if err != nil {
//		return err
//	}
//	balances, err := selectBalancesForAccount(db, a.ID)
//	if err == NoBalances {
//		err = InvalidAccountBalanceError{
//			AccountID: a.ID,
//			BalanceID: balance.ID,
//		}
//	}
//	if err != nil {
//		return err
//	}
//	for _, accountBalance := range *balances {
//		if accountBalance.ID == balance.ID {
//			return nil
//		}
//	}
//	return InvalidAccountBalanceError{
//		AccountID: a.ID,
//		BalanceID: balance.ID,
//	}
//}
//
//
// SelectAccountsOpen returns an Accounts item holding all Account entries within the given database that are open along with any errors occured whilst attempting to retrieve the Accounts.
//func SelectAccountsOpen(db *sql.DB) (*Accounts, error) {
//	queryString := "SELECT " + selectFields + " FROM accounts WHERE date_closed IS NULL AND deleted_at IS NULL ORDER BY id ASC;"
//	rows, err := db.Query(queryString)
//	if err != nil {
//		return new(Accounts), err
//	}
//	defer deferredClose(rows)
//	return scanRowsForAccounts(rows)
//}
//
// SelectAccountWithID returns the Account from the DB with the given ID value along with any error that occurs whilst attempting to retrieve the Account.
//func SelectAccountWithID(db *sql.DB, id uint) (Account, error) {
//	if db == nil {
//		return Account{}, errors.New("nil db pointer")
//	}
//	row := db.QueryRow("SELECT "+selectFields+" FROM accounts WHERE id = $1;", id)
//	account, err := scanRowForAccount(row)
//	if err == sql.ErrNoRows {
//		err = NoAccountWithIDError(id)
//	}
//	if account == nil {
//		account = new(Account)
//	}
//	return *account, err
//}
//
// CreateAccount created an Account entry within the DB and returns it, if successful, along with any errors that occur whilst attempting to create the Account.
//func CreateAccount(db *sql.DB, newAccount account.Account) (*Account, error) {
//	var queryString bytes.Buffer
//	fmt.Fprintf(&queryString, `INSERT INTO accounts (%s) `, insertFields)
//	fmt.Fprint(&queryString, `VALUES ($1, $2, $3) `)
//	fmt.Fprintf(&queryString, `returning %s`, selectFields)
//	row := db.QueryRow(queryString.String(), newAccount.Name, newAccount.Start(), pq.NullTime(newAccount.End()))
//	return scanRowForAccount(row)
//}

//SelectBalanceWithID returns a Balance from the database that has the given ID if the account is the correct one that it belongs to.
//Otherwise, SelectBalanceWithID returns an empty Balance object and an error.
//func (a Account) SelectBalanceWithID(db *sql.DB, id uint) (*Balance, error) {
//	if err := a.validate(db); err != nil {
//		return new(Balance), err
//	}
//	var query bytes.Buffer
//	fmt.Fprintf(&query, `SELECT %s FROM balances b JOIN accounts a ON b.account_id = a.id `, bBalanceSelectFields)
//	fmt.Fprint(&query, `WHERE a.deleted_at IS NULL AND b.account_id = $1 AND b.id = $2`)
//	row := db.QueryRow(query.String(), a.ID, id)
//	return scanRowForBalance(row)
//}

// scanRowsForAccounts scans an sql.Rows object for go-moneypostgres.Accounts objects and returns then along with any error that occurs whilst attempting to scan.
func scanRowsForAccounts(rows *sql.Rows) (*storage.Accounts, error) {
	var openAccounts storage.Accounts
	for rows.Next() {
		var id uint
		var name string
		var start time.Time
		var end, deletedAt pq.NullTime
		err := rows.Scan(&id, &name, &start, &end, &deletedAt)
		if err != nil {
			return nil, err
		}
		c, err := currency.NewCode("GBP")
		if err != nil {
			return nil, err
		}
		innerAccount, err := account.New(name, *c, start)
		if err != nil {
			return nil, err
		}
		if end.Valid {
			err = account.CloseTime(end.Time)(innerAccount)
			if err != nil {
				return nil, err
			}
		}
		a := &storage.Account{ID: id, Account: *innerAccount}
		if deletedAt.Valid {
			err := storage.DeletedAt(deletedAt.Time)(a)
			if err != nil {
				return nil, err
			}
		}
		openAccounts = append(openAccounts, *a)
	}
	return &openAccounts, rows.Err()
}

// scanRowForAccount scans a single sql.Row for a Account object and returns any error occurring along the way.
// If the account exists but has been marked as deleted, an ErrAccountDeleted error will be returned along with the account.
//func scanRowForAccount(row *sql.Row) (*Account, error) {
//	var id uint
//	var name string
//	var start time.Time
//	var end, deletedAt pq.NullTime
//	if err := row.Scan(&id, &name, &start, &end, &deletedAt); err != nil {
//		return nil, err
//	}
//	innerAccount, err := account.New(name, start, gohtime.NullTime(end))
//	if err != nil {
//		return nil, err
//	}
//	if deletedAt.Valid {
//		err = ErrAccountDeleted
//	}
//	return &Account{ID: id, Account: innerAccount, deletedAt: gohtime.NullTime(deletedAt)}, err
//}

//Update updates an Account entry in a given db, returning any errors that are present with the validity of the original Account or update Account.
//func (a Account) Update(db *sql.DB, update account.Account) (Account, error) {
//	if err := a.validate(db); err != nil {
//		return Account{}, err
//	}
//	if err := update.Validate(); err != nil {
//		return Account{}, errors.New(`Update Account is not valid: ` + err.Error())
//	}
//	balances, err := a.Balances(db)
//	if err != nil && err != NoBalances {
//		return Account{}, errors.New("Error selecting balances for validation: " + err.Error())
//	}
//	for _, b := range *balances {
//		if err = update.ValidateBalance(b.balance); err != nil {
//			return Account{}, fmt.Errorf("Update would make at least one account balance (id: %d) invalid. Error: %s", b.ID, err)
//		}
//	}
//	account, err := scanRowForAccount(
//		db.QueryRow(`UPDATE accounts SET name = $1, date_opened = $2, date_closed = $3 WHERE id = $4 returning `+selectFields, update.Name, update.Start(), pq.NullTime(update.End()), a.ID),
//	)
//	return *account, err
//}
//
// Delete marks an Account as deleted in the DB, returning any errors that occur whilst attempting the deletion.
//func (a *Account) Delete(db *sql.DB) error {
//	if err := a.validate(db); err != nil {
//		return errors.New("Account is not valid. " + err.Error())
//	}
//	deletedAt := pq.NullTime{Valid: true, Time: time.Now()}
//	_, err := scanRowForAccount(
//		db.QueryRow(`UPDATE accounts SET deleted_at = $1 WHERE id = $2 returning `+selectFields, deletedAt, a.ID),
//	)
//	if err == ErrAccountDeleted {
//		a.deletedAt = gohtime.NullTime(deletedAt)
//		return nil
//	}
//	if err != nil {
//		return err
//	}
//	Should not be reached.
//return errors.New("internal error when deleting account")
//}
//
// validate returns any errors that are present with an Account object
//func (a Account) validate(db *sql.DB) error {
//	b, err := SelectAccountWithID(db, a.ID)
//	if err != nil {
//		return err
//	}
//	if a.deletedAt.Valid && b.deletedAt.Valid && !a.deletedAt.Time.Equal(b.deletedAt.Time) {
//		return ErrAccountDifferentInDbAndRuntime
//	}
//	if !a.Account.Equal(b.Account) {
//		return ErrAccountDifferentInDbAndRuntime
//	}
//	if err := a.Account.Validate(); err != nil {
//		return nil
//	}
//	return err
//}

func nonReturningClose(c io.Closer) {
	if c == nil {
		log.Printf("Attempted to close Closer but it was nil.")
		return
	}
	if err := c.Close(); err != nil {
		log.Printf("Error closing postgres: %v", err)
	}
}
