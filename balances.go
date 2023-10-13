package ledger

import (
	"slices"
	"strings"
)

// GetBalances provided a list of transactions and filter strings, returns account balances of
// all accounts that have any filter as a substring of the account name. Also
// returns balances for each account level depth as a separate record.
//
// Accounts are sorted by name.
func GetBalances(generalLedger []*Transaction, filterArr []string) []*Account {
	var accList []*Account
	balances := make(map[string]*Account)
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for i := 0; i < len(filterArr) && !inFilter; i++ {
				if strings.Contains(accChange.Name, filterArr[i]) {
					inFilter = true
				}
			}
			if inFilter {
				accHier := strings.Split(accChange.Name, ":")
				accDepth := len(accHier)
				for currDepth := accDepth; currDepth > 0; currDepth-- {
					currAccName := strings.Join(accHier[:currDepth], ":")
					if acc, ok := balances[currAccName]; !ok {
						acc := &Account{Name: currAccName, Balance: accChange.Balance}
						accList = append(accList, acc)
						balances[currAccName] = acc
					} else {
						acc.Balance = acc.Balance.Add(accChange.Balance)
					}
				}
			}
		}
	}

	slices.SortFunc(accList, func(a, b *Account) int {
		return strings.Compare(a.Name, b.Name)
	})
	return accList
}
