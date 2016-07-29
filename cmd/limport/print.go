package main

import (
	"fmt"
	"strings"

	"github.com/howeyc/ledger"
)

// PrintTransaction prints a transaction formatted to fit in specified column width.
func PrintTransaction(trans *ledger.Transaction, columns int) {
	fmt.Printf("%s %s\n", trans.Date.Format(transactionDateFormat), trans.Payee)
	for _, accChange := range trans.AccountChanges {
		outBalanceString := accChange.Balance.FloatString(displayPrecision)
		spaceCount := columns - 4 - len(accChange.Name) - len(outBalanceString)
		fmt.Printf("    %s%s%s\n", accChange.Name, strings.Repeat(" ", spaceCount), outBalanceString)
	}
	fmt.Println("")
}
