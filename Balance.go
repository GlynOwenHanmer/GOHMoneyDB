package GOHMoneyDB

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"encoding/json"

	"github.com/GlynOwenHanmer/GOHMoney/balance"
	"github.com/GlynOwenHanmer/GOHMoney/money"
)

const (
	balanceInsertFields  string = "account_id, date, balance, currency"
	balanceSelectFields  string = "id, date, balance, currency"
	bBalanceSelectFields string = "b.id, b.date, b.balance, currency"
)

// Balance holds logic for an Account item that is held within a GOHMoney database.
type Balance struct {
	balance.Balance
	ID uint
}

// Equal returns true if two Balance items are logically identical
func (b Balance) Equal(ob Balance) bool {
	if b.ID != ob.ID || !b.Balance.Equal(ob.Balance) {
		return false
	}
	return true
}

// Balances holds multiple Balance items
type Balances []Balance

// Balances returns all Balances for a given Account and any errors that occur whilst attempting to retrieve the Balances.
// The Balances are sorted by chronological order then by the id of the Balance in the DB
func (a Account) Balances(db *sql.DB) (*Balances, error) {
	return selectBalancesForAccount(db, a.ID)
}

// selectBalancesForAccount returns all Balance items, as a single Balances item, for a given account ID number in the given database, along with any errors that occur whilst attempting to retrieve the Balances.
// The Balances are sorted by chronological order then by the id of the Balance in the DB
func selectBalancesForAccount(db *sql.DB, accountID uint) (*Balances, error) {
	rows, err := db.Query("SELECT "+balanceSelectFields+" FROM balances b WHERE account_id = $1 ORDER BY date ASC, ID ASC", accountID)
	if err != nil {
		return &Balances{}, err
	}
	defer rows.Close()
	return scanRowsForBalances(rows)
}

// InsertBalance adds a Balance entry to the given DB for the given account and returns the inserted Balance item with any errors that occured while attempting to insert the Balance.
func (a Account) InsertBalance(db *sql.DB, b balance.Balance) (Balance, error) {
	err := a.Account.ValidateBalance(b)
	if err != nil {
		dbb, _ := newBalance(0, time.Time{}, 0, "")
		return *dbb, err
	}
	var query bytes.Buffer
	fmt.Fprintf(&query, `INSERT INTO balances (%s) VALUES ($1, $2, $3, $4) `, balanceInsertFields)
	fmt.Fprintf(&query, `RETURNING %s;`, balanceSelectFields)
	amount := b.Money()
	floatAmount := float64((&amount).Amount()) / 100.
	code := "non"
	if cur, err := b.Money().Currency(); err == nil {
		code = cur.Code
	}
	row := db.QueryRow(query.String(), a.ID, b.Date(), floatAmount, code)
	balance, err := scanRowForBalance(row)
	return *balance, err
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
	amount := update.Money()
	floatAmount := float64((&amount).Amount()) / 100.
	row := db.QueryRow(`UPDATE balances SET balance = $1, date = $2 WHERE id = $3 returning `+balanceSelectFields, floatAmount, update.Date(), original.ID)
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

type jsonHelper struct {
	ID    uint
	Date  time.Time
	Money money.Money
}

// MarshalJSON is a custom JSON marshalling method to avoid the custom JSON marshalling method of the Balance's inner Balance method being called instead.
func (b Balance) MarshalJSON() ([]byte, error) {
	return json.Marshal(jsonHelper{
		ID:    b.ID,
		Date:  b.Date(),
		Money: b.Money(),
	})
}

// UnmarshalJSON is a custom JSON unmarshalling method to avoid the custom JSON unmarshalling method of the Balance's inner Balance method being called instead.
func (b *Balance) UnmarshalJSON(data []byte) (err error) {
	var aux jsonHelper
	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}
	b.ID = aux.ID
	b.Balance, err = balance.New(aux.Date, aux.Money)
	return
}

// scanRowForBalance scans a single sql.Row for a Balance object and returns any error occurring along the way.
func scanRowForBalance(row *sql.Row) (*Balance, error) {
	b := new(Balance)
	var ID uint
	var date time.Time
	var amount float64
	var currency string
	err := row.Scan(&ID, &date, &amount, &currency)
	b, _ = newBalance(ID, date, amount, "GBP")
	if err == sql.ErrNoRows {
		err = NoBalances
	}
	if err != nil {
		return b, err
	}
	return b, b.Validate()
}

// scanRowsForBalance scans a sql.Rows for a Balances object and returns any error occurring along the way.
func scanRowsForBalances(rows *sql.Rows) (bs *Balances, err error) {
	bs = new(Balances)
	for rows.Next() {
		var ID uint
		var date time.Time
		var amount float64
		var cur string
		err = rows.Scan(&ID, &date, &amount, &cur)
		if err != nil {
			return nil, err
		}
		var innerB balance.Balance
		m, err := moneyIntFromFloat(amount, cur)
		if err != nil {
			return nil, err
		}
		innerB, err = balance.New(date, *m)
		if err != nil {
			return nil, err
		}
		*bs = append(*bs, Balance{ID: ID, Balance: innerB})
	}
	if err == nil {
		err = rows.Err()
	}
	if err == nil && len(*bs) == 0 {
		err = NoBalances
	}
	return
}

func newBalance(ID uint, d time.Time, a float64, cur string) (*Balance, error) {
	mon, err := moneyIntFromFloat(a, cur)
	innerB := new(balance.Balance)
	*innerB, err = balance.New(d, *mon)
	return &Balance{ID: ID, Balance: *innerB}, err
}

func moneyIntFromFloat(f float64, cur string) (*money.Money, error) {
	return money.New(int64(f * 100), cur)
}
