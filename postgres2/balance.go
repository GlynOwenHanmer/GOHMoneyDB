package postgres2

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/balance"
	"github.com/pkg/errors"
)

const (
	balancesFieldAccountID = "account_id"
	balancesFieldAmount    = "amount"
	balancesFieldID        = "id"
	balancesFieldTime      = "time"
	balancesTable          = "balances"
)

var (
	balancesSelectFields = fmt.Sprintf(
		"%s, %s, %s",
		balancesFieldID,
		balancesFieldTime,
		balancesFieldAmount)

	balancesSelectPrefix = fmt.Sprintf(
		`SELECT %s FROM %s WHERE `,
		balancesSelectFields,
		balancesTable)

	balancesSelectBalanceByID = fmt.Sprintf(
		`%s%s = $1;`,
		balancesSelectPrefix,
		balancesFieldID)

	balancesSelectBalancesForAccountId = fmt.Sprintf(
		"%s%s = $1 ORDER BY %s ASC, %s ASC;",
		balancesSelectPrefix,
		balancesFieldAccountID,
		balancesFieldTime,
		balancesFieldAccountID)

	balancesInsertFields = fmt.Sprintf(
		"%s, %s, %s",
		balancesFieldAccountID,
		balancesFieldTime,
		balancesFieldAmount)

	balancesInsertBalance = fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES ($1, $2, $3) returning %s;`,
		balancesTable,
		balancesInsertFields,
		balancesSelectFields)
)

//Balances returns all Balances for a given Account and any errors that occur whilst attempting to retrieve the Balances.
// The Balances are sorted by chronological order then by the id of the Balance in the DB
func (pg postgres) SelectAccountBalances(a storage.Account) (*storage.Balances, error) {
	return pg.selectBalancesForAccountID(a.ID)
}

//selectBalancesForAccount returns all Balance items, as a single Balances item, for a given account ID number in the given database, along with any errors that occur whilst attempting to retrieve the Balances.
//The Balances are sorted by chronological order then by the id of the Balance in the DB
func (pg postgres) selectBalancesForAccountID(accountID uint) (*storage.Balances, error) {
	return queryBalances(pg.db, balancesSelectBalancesForAccountId, accountID)
}

// SelectBalanceByAccountAndID selects a balance with a given ID within a given account.
// An error will be returned if no balance can be found with the ID for the given account.
func (pg postgres) SelectBalanceByAccountAndID(a storage.Account, balanceID uint) (*storage.Balance, error) {
	bs, err := pg.SelectAccountBalances(a)
	if err != nil {
		return nil, errors.Wrap(err, "selecting account balances for account %+v")
	}
	for _, b := range *bs {
		if b.ID == balanceID {
			return &b, nil
		}
	}
	return nil, fmt.Errorf("no balance with id %d for account", balanceID)
}

func (pg postgres) selectBalanceByID(id uint) (*storage.Balance, error) {
	return queryBalance(pg.db, balancesSelectBalanceByID, id)
}

func (pg postgres) InsertBalance(a storage.Account, b balance.Balance) (*storage.Balance, error) {
	err := a.ValidateBalance(b)
	if err != nil {
		return nil, errors.Wrap(err, "validating balance")
	}
	dbb, err := queryBalance(pg.db, balancesInsertBalance, a.ID, b.Date, b.Amount)
	return dbb, errors.Wrap(err, "querying Balance")
}

// TODO: check behaviour of queryBalance when 0 and 1 results are returned. Maybe it should return an error if there are non present but queryBalances should not do?
// queryBalance returns an error if more than one result is returned from the query
// queryBalance may or may not return an error if zero results are returned.
func queryBalance(db *sql.DB, queryString string, values ...interface{}) (*storage.Balance, error) {
	bs, err := queryBalances(db, queryString, values...)
	if err != nil {
		return nil, errors.Wrap(err, "querying balances")
	}
	var b *storage.Balance
	if len(*bs) > 1 {
		err = errors.New("query returned more than 1 result")
	} else if bs != nil {
		b = &(*bs)[0]
	}
	return b, err
}

func queryBalances(db *sql.DB, queryString string, values ...interface{}) (*storage.Balances, error) {
	rows, err := db.Query(queryString, values...)
	if err != nil {
		return nil, errors.Wrap(err, "querying db")
	}
	defer nonReturningCloseRows(rows)
	return scanRowsForBalances(rows)
}

//scanRowsForBalance scans a sql.Rows for a Balances object and returns any error occurring along the way.
func scanRowsForBalances(rows *sql.Rows) (bs *storage.Balances, err error) {
	bs = new(storage.Balances)
	for rows.Next() {
		var ID uint
		var date time.Time
		var amount float64
		err = rows.Scan(&ID, &date, &amount)
		if err != nil {
			return nil, errors.Wrap(err, "scanning rows")
		}
		var innerB *balance.Balance
		innerB, err = balance.New(date, balance.Amount(int(amount)))
		if err != nil {
			return nil, errors.Wrap(err, "creating new balance from scan results")
		}
		*bs = append(*bs, storage.Balance{ID: ID, Balance: *innerB})
	}
	if err == nil {
		err = errors.Wrap(rows.Err(), "rows error: ")
	}
	return
}
