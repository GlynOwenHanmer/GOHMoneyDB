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
	balancesFieldAmount    = "amount"
	balancesFieldID        = "id"
	balancesFieldTime      = "time"
	balancesTable          = "balances"
)

var (
	balancesSelectFields = fmt.Sprintf(
		"%s, %s, %s",
		balancesFieldID, balancesFieldTime, balancesFieldAmount)
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
		balancesFieldAccountID, balancesFieldTime, balancesFieldAmount)
	balancesInsertBalance = fmt.Sprintf(
		`INSERT INTO %s (%s) VALUES ($1, $2, $3) returning %s;`,
		balancesTable,
		balancesInsertFields,
		balancesFieldID)
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

func (pg postgres) selectBalanceByID(id uint) (*storage.Balance, error) {
	return queryBalance(pg.db, balancesSelectBalanceByID, id)
}

func (pg postgres) InsertBalance(a storage.Account, b balance.Balance) (*storage.Balance, error) {
	err := a.ValidateBalance(b)
	if err != nil {
		return nil, err
	}
	id, err := queryUint(pg, balancesInsertBalance, a.ID, b.Date, b.Amount)
	if err != nil {
		return nil, err
	}
	return &storage.Balance{
		ID:      *id,
		Balance: b,
	}, nil
}

func queryBalance(db *sql.DB, queryString string, values ...interface{}) (*storage.Balance, error) {
	bs, err := queryBalances(db, queryString, values...)
	if err != nil {
		return nil, err
	}
	var b *storage.Balance
	if len(*bs) > 1 {
		err = errors.New("query returned more than 1 result")
	} else if bs != nil {
		*b = (*bs)[0]
	}
	return b, err
}

func queryBalances(db *sql.DB, queryString string, values ...interface{}) (*storage.Balances, error) {
	rows, err := db.Query(queryString, values...)
	if err != nil {
		return nil, err
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
