package postgres2

import (
	"database/sql"
	"time"

	"fmt"

	"errors"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/balance"
)

const (
	balancesFieldAccountID = "account_id"
	balancesFieldBalance   = "balance"
	balancesFieldID        = "id"
	balancesFieldDate      = "date"
	balancesTable          = "balances"
)

var (
	balancesSelectFields                    = fmt.Sprintf("%s, %s, %s", balancesFieldID, balancesFieldDate, balancesFieldBalance)
	balancesQuerySelectBalancesForAccountId = fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s = $1 ORDER BY %s ASC, ID ASC",
		balancesSelectFields,
		balancesTable,
		balancesFieldAccountID,
		balancesFieldDate,
	)
)

//Balances returns all Balances for a given Account and any errors that occur whilst attempting to retrieve the Balances.
// The Balances are sorted by chronological order then by the id of the Balance in the DB
func (pg postgres) SelectAccountBalances(a storage.Account) (*storage.Balances, error) {
	return pg.selectBalancesForAccountID(a.ID)
}

//selectBalancesForAccount returns all Balance items, as a single Balances item, for a given account ID number in the given database, along with any errors that occur whilst attempting to retrieve the Balances.
//The Balances are sorted by chronological order then by the id of the Balance in the DB
func (pg postgres) selectBalancesForAccountID(accountID uint) (*storage.Balances, error) {
	return queryBalances(pg.db, balancesQuerySelectBalancesForAccountId, accountID)
}

func queryBalances(db *sql.DB, queryString string, values ...interface{}) (*storage.Balances, error) {
	rows, err := db.Query(queryString, values...)
	if err != nil {
		return nil, err
	}
	defer nonReturningClose(rows)
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
			return nil, err
		}
		var innerB *balance.Balance
		innerB, err = balance.New(date, balance.Amount(int(amount)))
		if err != nil {
			return nil, err
		}
		*bs = append(*bs, storage.Balance{ID: ID, Balance: *innerB})
	}
	if err == nil {
		err = rows.Err()
	}
	return
}

func (pg postgres) InsertBalance(a storage.Account, b balance.Balance) (*storage.Balance, error) {
	return nil, errors.New("not supported")
}
