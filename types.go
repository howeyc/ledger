package ledger

import (
	"time"

	"github.com/howeyc/ledger/decimal"
)

// Account holds the name and balance
type Account struct {
	Name string
	// Default "" for no currency/token displayed
	Currency string
	Balance  decimal.Decimal
	Comment  string

	// Balance converted using @@ notation
	Converted *decimal.Decimal
	// Conversion factor using @ notation
	ConversionFactor *decimal.Decimal
}

// Transaction is the basis of a ledger. The ledger holds a list of transactions.
// A Transaction has a Payee, Date (with no time, or to put another way, with
// hours,minutes,seconds values that probably doesn't make sense), and a list of
// Account values that hold the value of the transaction for each account.
type Transaction struct {
	Date           time.Time
	Payee          string
	PayeeComment   string
	AccountChanges []Account
	Comments       []string
}
