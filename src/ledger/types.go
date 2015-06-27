package ledger

import (
	"math/big"
	"time"
)

type Account struct {
	Name    string
	Balance *big.Rat
}

type sortAccounts []*Account

func (s sortAccounts) Len() int      { return len(s) }
func (s sortAccounts) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortAccountsByName struct{ sortAccounts }

func (s sortAccountsByName) Less(i, j int) bool {
	return s.sortAccounts[i].Name < s.sortAccounts[j].Name
}

type Transaction struct {
	Payee          string
	Date           time.Time
	AccountChanges []Account
}

type sortTransactions []*Transaction

func (s sortTransactions) Len() int      { return len(s) }
func (s sortTransactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortTransactionsByDate struct{ sortTransactions }

func (s sortTransactionsByDate) Less(i, j int) bool {
	return s.sortTransactions[i].Date.Before(s.sortTransactions[j].Date)
}
