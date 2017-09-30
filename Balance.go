package GOHMoneyDB

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/GlynOwenHanmer/GOHMoney/balance"
)

const (
	balanceInsertFields  string = "account_id, date, balance"
	balanceSelectFields  string = "id, date, balance"
	bBalanceSelectFields string = "b.id, b.date, b.balance"
)

// Balance holds logic for an Account item that is held within a GOHMoney database.
type Balance struct {
	balance.Balance
	ID uint `json:"id"`
}

// Balances holds multiple Balance items
type Balances []Balance

// Balances returns all Balances for a given Account and any errors that occur whilst attempting to retrieve the Balances.
// The Balances are sorted by chronological order then by the id of the Balance in the DB
func (a Account) Balances(db *sql.DB) (Balances, error) {
	return selectBalancesForAccount(db, a.ID)
}

// selectBalancesForAccount returns all Balance items, as a single Balances item, for a given account ID number in the given database, along with any errors that occur whilst attempting to retrieve the Balances.
// The Balances are sorted by chronological order then by the id of the Balance in the DB
func selectBalancesForAccount(db *sql.DB, accountID uint) (Balances, error) {
	rows, err := db.Query("SELECT "+balanceSelectFields+" FROM balances b WHERE account_id = $1 ORDER BY date ASC, ID ASC", accountID)
	if err != nil {
		return Balances{}, err
	}
	defer rows.Close()
	balances := Balances{}
	for rows.Next() {
		balance := Balance{}
		err := rows.Scan(&balance.ID, &balance.Date, &balance.Amount)
		if err != nil {
			return nil, err
		}
		balances = append(balances, balance)
	}
	return balances, rows.Err()
}

// InsertBalance adds a Balance entry to the given DB for the given account and returns the inserted Balance item with any errors that occured while attempting to insert the Balance.
func (a Account) InsertBalance(db *sql.DB, b balance.Balance) (Balance, error) {
	err := a.Account.ValidateBalance(b)
	if err != nil {
		return Balance{}, err
	}
	var query bytes.Buffer
	fmt.Fprintf(&query, `INSERT INTO balances (%s) VALUES ($1, $2, $3) `, balanceInsertFields)
	fmt.Fprintf(&query, `RETURNING %s;`, balanceSelectFields)
	row := db.QueryRow(query.String(), a.ID, b.Date, b.Amount)
	var insertedBalance Balance
	return insertedBalance, row.Scan(&insertedBalance.ID, &insertedBalance.Date, &insertedBalance.Amount)
}

// UpdateBalance updates a Balance entry in a given db for a given account and original Balance, returning any errors that are present with the validitiy of the Account, original Balance or update Balance.
func (a Account) UpdateBalance(db *sql.DB, original Balance, update balance.Balance) (Balance, error) {
	if err := a.ValidateBalance(db, original); err != nil {
		return Balance{}, err
	}
	if err := update.Validate(); err != nil {
		return Balance{}, errors.New(`Update Balance is not valid: ` + err.Error())
	}
	if err := a.Account.ValidateBalance(update); err != nil {
		return Balance{}, errors.New(`Update is not valid for account: ` + err.Error())
	}
	row := db.QueryRow(`UPDATE balances SET balance = $1, date = $2 WHERE id = $3 returning `+balanceSelectFields, update.Amount, update.Date, original.ID)
	balance, err := scanRowForBalance(row)
	return *balance, err
}

// BalanceAtDate returns a Balance item representing the Balance of an account at the given time for the given account with the given DB.
func (a Account) BalanceAtDate(db *sql.DB, time time.Time) (Balance, error) {
	var query bytes.Buffer
	fmt.Fprintf(&query, `SELECT %s`, balanceSelectFields)
	fmt.Fprint(&query, ` FROM balances `)
	fmt.Fprint(&query, `WHERE account_id = $1 AND date <= $2 `)
	fmt.Fprint(&query, `ORDER BY date DESC, id DESC LIMIT 1;`)
	row := db.QueryRow(query.String(), a.ID, time)
	balance, err := scanRowForBalance(row)
	return *balance, err
}

// scanRowForBalance scans a single sql.Row for a Balance object and returns any error occurring along the way.
func scanRowForBalance(row *sql.Row) (*Balance, error) {
	var balance Balance
	err := row.Scan(&balance.ID, &balance.Date, &balance.Amount)
	if err == sql.ErrNoRows {
		err = NoBalances
	}
	return &balance, err
}
