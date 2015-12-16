package ledger

import (
	"math/big"
	"sort"
	"strings"
)

// GetBalances provided a list of transactions and filter strings, returns account balances of
// all accounts that have any filter as a substring of the account name. Also
// returns balances for each account level depth as a separate record.
//
// Accounts are sorted by name.
func GetBalances(generalLedger []*Transaction, filterArr []string) []*Account {
	balances := make(map[string]*big.Rat)
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
			if inFilter {
				accHier := strings.Split(accChange.Name, ":")
				accDepth := len(accHier)
				for currDepth := accDepth; currDepth > 0; currDepth-- {
					currAccName := strings.Join(accHier[:currDepth], ":")
					if ratNum, ok := balances[currAccName]; !ok {
						ratNum = new(big.Rat)
						ratNum.SetString(accChange.Balance.RatString())
						balances[currAccName] = ratNum
					} else {
						ratNum.Add(ratNum, accChange.Balance)
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

	sort.Sort(sortAccountsByName{accList})
	return accList
}
