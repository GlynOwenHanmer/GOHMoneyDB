package GOHMoneyDB

import (
	"github.com/glynternet/GOHMoney/money"
)

type dbMoney struct {
	innerMoney money.Money
}
//
//func (m *dbMoney) Scan(src interface{}) error {
//	f, ok := src.(float64)
//	if !ok {
//		return errors.New("source is not a float64 value and can't be scanned")
//	}
//	*m = dbMoney{money.GBP(int64(f * 100))}
//	return nil
//}
//
//func (m dbMoney) Value() (driver.Value, error) {
//	return float64(m.innerMoney.Amount()) / 100., nil
//}
