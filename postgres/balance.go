package postgres

import (
	"database/sql"
	"time"

	"github.com/glynternet/go-accounting-storage"
	"github.com/glynternet/go-accounting/balance"
)

const (
	//balanceInsertFields  = "account_id, date, balance, currency"
	balanceSelectFields = "id, date, balance"

//bBalanceSelectFields = "b.id, b.date, b.balance, currency"
)

//Balances returns all Balances for a given Account and any errors that occur whilst attempting to retrieve the Balances.
// The Balances are sorted by chronological order then by the id of the Balance in the DB
func (pg postgres) SelectAccountBalances(a storage.Account) (*storage.Balances, error) {
	return pg.selectBalancesForAccountID(a.ID)
}

//selectBalancesForAccount returns all Balance items, as a single Balances item, for a given account ID number in the given database, along with any errors that occur whilst attempting to retrieve the Balances.
//The Balances are sorted by chronological order then by the id of the Balance in the DB
func (pg postgres) selectBalancesForAccountID(accountID uint) (*storage.Balances, error) {
	rows, err := pg.db.Query("SELECT "+balanceSelectFields+" FROM balances b WHERE account_id = $1 ORDER BY date ASC, ID ASC", accountID)
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
