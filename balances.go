package ledger

import (
	"slices"
	"strings"

	"github.com/howeyc/ledger/decimal"
)

// GetBalances provided a list of transactions and filter strings, returns account balances of
// all accounts that have any filter as a substring of the account name. Also
// returns balances for each account level depth as a separate record.
//
// Accounts are sorted by name.
func GetBalances(generalLedger []*Transaction, filterArr []string) []*Account {
	var accList []*Account
	balances := make(map[string]map[string]*Account)

	// at every depth, for each account, track the parent account
	depthMap := make(map[int]map[string]string)
	var maxDepth int

	incAccount := func(accName string, currency string, val decimal.Decimal) {
		// track parent
		accDepth := strings.Count(accName, ":") + 1
		pmap, pmapfound := depthMap[accDepth]
		if !pmapfound {
			pmap = make(map[string]string)
			depthMap[accDepth] = pmap
		}
		if _, foundparent := pmap[accName]; !foundparent && accDepth > 1 {
			colIdx := strings.LastIndex(accName, ":")
			pmap[accName] = accName[:colIdx]
			maxDepth = max(maxDepth, accDepth)
		}

		// add to balance
		if _, ok := balances[accName]; !ok {
			balances[accName] = make(map[string]*Account)
		}

		if acc, ok := balances[accName][currency]; !ok {
			acc := &Account{Name: accName, Currency: currency, Balance: val}
			accList = append(accList, acc)
			balances[accName][currency] = acc
		} else {
			acc.Balance = acc.Balance.Add(val)
		}
	}

	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for i := 0; i < len(filterArr) && !inFilter; i++ {
				if strings.Contains(accChange.Name, filterArr[i]) {
					inFilter = true
				}
			}
			if inFilter {
				incAccount(accChange.Name, accChange.Currency, accChange.Balance)
			}
		}
	}

	// roll-up balances
	for curDepth := maxDepth; curDepth > 1; curDepth-- {
		for accName, parentName := range depthMap[curDepth] {
			for currency, acc := range balances[accName] {
				incAccount(parentName, currency, acc.Balance)
			}
		}
	}

	slices.SortFunc(accList, func(a, b *Account) int {
		return strings.Compare(a.Name, b.Name)
	})
	return accList
}
