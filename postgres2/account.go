package postgres2

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-money/currency"
	"github.com/lib/pq"
)

const (
	fieldID       = "id"
	fieldName     = "name"
	fieldOpened   = "opened"
	fieldClosed   = "closed"
	fieldCurrency = "currency"
	fieldDeleted  = "deleted"
	table         = "accounts"
)

var (
	fieldsInsert        = fmt.Sprintf("%s, %s, %s, %s", fieldName, fieldOpened, fieldClosed, fieldCurrency)
	fieldsSelect        = fmt.Sprintf("%s, %s, %s, %s, %s, %s", fieldID, fieldName, fieldOpened, fieldClosed, fieldCurrency, fieldDeleted)
	querySelectAccounts = fmt.Sprintf("SELECT %s FROM %s WHERE %s IS NULL ORDER BY %s ASC;", fieldsSelect, table, fieldDeleted, fieldID)
	queryInsertAccount  = fmt.Sprintf(`INSERT INTO accounts (%s) VALUES ($1, $2, $3, $4) returning %s`, fieldsInsert, fieldID)
)

// SelectAccounts returns an Accounts item holding all Account entries within the given database along with any errors that occurred whilst attempting to retrieve the Accounts.
func (pg postgres) SelectAccounts() (*storage.Accounts, error) {
	return queryAccounts(pg.db, querySelectAccounts)
}

func (pg postgres) InsertAccount(a account.Account) (*storage.Account, error) {
	id, err := queryUint(pg, queryInsertAccount, a.Name(), a.Opened(), pq.NullTime(a.Closed()), a.CurrencyCode())
	if err != nil {
		return nil, err
	}
	return &storage.Account{ID: *id, Account: a}, err
}

// scanRowForAccount scans a single sql.Row for a Account object and returns any error occurring along the way.
// If the account exists but has been marked as deleted, an ErrAccountDeleted error will be returned along with the account.
func scanRowForAccount(row *sql.Row) (*storage.Account, error) {
	var id uint
	var name, currencyCode string
	var opened time.Time
	var closed, deleted pq.NullTime
	if err := row.Scan(&id, &name, &opened, &closed, &currencyCode, &deleted); err != nil {
		return nil, err
	}
	c, err := currency.NewCode(currencyCode)
	if err != nil {
		return nil, err
	}
	innerAccount, err := account.New(name, *c, opened)
	if err != nil {
		return nil, err
	}
	if closed.Valid {
		err := account.CloseTime(closed.Time)(innerAccount)
		if err != nil {
			return nil, err
		}
	}
	return &storage.Account{ID: id, Account: *innerAccount}, nil
}

func queryAccounts(db *sql.DB, queryString string, values ...interface{}) (*storage.Accounts, error) {
	rows, err := db.Query(queryString, values...)
	if err != nil {
		return nil, err
	}
	defer nonReturningCloseRows(rows)
	return scanRowsForAccounts(rows)
}

// scanRowsForAccounts scans an sql.Rows object for go-moneypostgres.Accounts objects and returns then along with any error that occurs whilst attempting to scan.
func scanRowsForAccounts(rows *sql.Rows) (*storage.Accounts, error) {
	var openAccounts storage.Accounts
	for rows.Next() {
		var id uint
		var name, code string
		var opened time.Time
		var closed, deleted pq.NullTime
		// 	fieldID, fieldName, fieldOpened, fieldClosed, fieldCurrency, fieldDeleted)
		err := rows.Scan(&id, &name, &opened, &closed, &code, &deleted)
		if err != nil {
			return nil, err
		}
		c, err := currency.NewCode(code)
		if err != nil {
			return nil, err
		}
		innerAccount, err := account.New(name, *c, opened)
		if err != nil {
			return nil, err
		}
		if closed.Valid {
			err = account.CloseTime(closed.Time)(innerAccount)
			if err != nil {
				return nil, err
			}
		}
		a := &storage.Account{ID: id, Account: *innerAccount}
		if deleted.Valid {
			err := storage.DeletedAt(deleted.Time)(a)
			if err != nil {
				return nil, err
			}
		}
		openAccounts = append(openAccounts, *a)
	}
	return &openAccounts, rows.Err()
}
