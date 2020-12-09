package main

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/howeyc/ledger"
)

// PrintTransaction prints a transaction formatted to fit in specified column width.
func PrintTransaction(trans *ledger.Transaction, columns int) {
	for _, c := range trans.Comments {
		fmt.Println(c)
	}
	fmt.Printf("%s %s\n", trans.Date.Format(transactionDateFormat), trans.Payee)
	for _, accChange := range trans.AccountChanges {
		outBalanceString := accChange.Balance.FloatString(displayPrecision)
		spaceCount := columns - 4 - utf8.RuneCountInString(accChange.Name) - utf8.RuneCountInString(outBalanceString)
		if spaceCount < 1 {
			spaceCount = 1
		}
		fmt.Printf("    %s%s%s\n", accChange.Name, strings.Repeat(" ", spaceCount), outBalanceString)
	}
	fmt.Println("")
}
