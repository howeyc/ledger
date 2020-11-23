package ledger

import (
	"math/big"
	"time"
)

// Account holds the name and balance
type Account struct {
	Name    string
	Balance *big.Rat
}

// Transaction is the basis of a ledger. The ledger holds a list of transactions.
// A Transaction has a Payee, Date (with no time, or to put another way, with
// hours,minutes,seconds values that probably doesn't make sense), and a list of
// Account values that hold the value of the transaction for each account.
type Transaction struct {
	Payee          string
	Date           time.Time
	AccountChanges []Account
	Comments       []string
}
