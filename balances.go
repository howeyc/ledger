package ledger

import (
	"sort"
	"strings"

	"github.com/howeyc/ledger/decimal"
)

// GetBalances provided a list of transactions and filter strings, returns account balances of
// all accounts that have any filter as a substring of the account name. Also
// returns balances for each account level depth as a separate record.
//
// Accounts are sorted by name.
func GetBalances(generalLedger []*Transaction, filterArr []string) []*Account {
	balances := make(map[string]decimal.Decimal)
	filters := len(filterArr) > 0
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := false
			if filters {
				for i := 0; i < len(filterArr) && !inFilter; i++ {
					if strings.Contains(accChange.Name, filterArr[i]) {
						inFilter = true
					}
				}
			} else {
				inFilter = true
			}
			if inFilter {
				accHier := strings.Split(accChange.Name, ":")
				accDepth := len(accHier)
				for currDepth := accDepth; currDepth > 0; currDepth-- {
					currAccName := strings.Join(accHier[:currDepth], ":")
					if ratNum, ok := balances[currAccName]; !ok {
						balances[currAccName] = accChange.Balance
					} else {
						balances[currAccName] = ratNum.Add(accChange.Balance)
					}
				}
			}
		}
	}

	accList := make([]*Account, len(balances))
	count := 0
	for accName, accBalance := range balances {
		account := &Account{Name: accName, Balance: accBalance}
		accList[count] = account
		count++
	}

	sort.Slice(accList, func(i, j int) bool {
		return accList[i].Name < accList[j].Name
	})
	return accList
}
