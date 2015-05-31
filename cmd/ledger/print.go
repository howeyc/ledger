package main

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/howeyc/ledger/pkg/ledger"
)

// Prints out account balances formated to a windows of a width of columns.
// Only shows accounts with names less than or equal to the given depth.
func PrintBalances(accountList []*ledger.Account, printZeroBalances bool, depth, columns int) {
	overallBalance := new(big.Rat)
	for _, account := range accountList {
		accDepth := len(strings.Split(account.Name, ":"))
		if accDepth == 1 {
			overallBalance.Add(overallBalance, account.Balance)
		}
		if (printZeroBalances || account.Balance.Sign() != 0) && (depth < 0 || accDepth <= depth) {
			outBalanceString := account.Balance.FloatString(ledger.DisplayPrecision)
			spaceCount := columns - len(account.Name) - len(outBalanceString)
			fmt.Printf("%s%s%s\n", account.Name, strings.Repeat(" ", spaceCount), outBalanceString)
		}
	}
	fmt.Println(strings.Repeat("-", columns))
	outBalanceString := overallBalance.FloatString(ledger.DisplayPrecision)
	spaceCount := columns - len(outBalanceString)
	fmt.Printf("%s%s\n", strings.Repeat(" ", spaceCount), outBalanceString)
}

// Prints a transaction formatted to fit in specified column width.
func PrintTransaction(trans *ledger.Transaction, columns int) {
	fmt.Printf("%s %s\n", trans.Date.Format(ledger.TransactionDateFormat), trans.Payee)
	for _, accChange := range trans.AccountChanges {
		outBalanceString := accChange.Balance.FloatString(ledger.DisplayPrecision)
		spaceCount := columns - 4 - len(accChange.Name) - len(outBalanceString)
		fmt.Printf("    %s%s%s\n", accChange.Name, strings.Repeat(" ", spaceCount), outBalanceString)
	}
	fmt.Println("")
}

// Prints all transactions as a formatted ledger file.
func PrintLedger(generalLedger []*ledger.Transaction, columns int) {
	for _, trans := range generalLedger {
		PrintTransaction(trans, columns)
	}
}

// Prints each transaction that matches the given filters.
func PrintRegister(generalLedger []*ledger.Transaction, filterArr []string, columns int) {
	runningBalance := new(big.Rat)
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
			if inFilter {
				runningBalance.Add(runningBalance, accChange.Balance)
				writtenBytes, _ := fmt.Printf("%s %s", trans.Date.Format(ledger.TransactionDateFormat), trans.Payee)
				outBalanceString := accChange.Balance.FloatString(ledger.DisplayPrecision)
				outRunningBalanceString := runningBalance.FloatString(ledger.DisplayPrecision)
				spaceCount := columns - writtenBytes - 2 - len(outBalanceString) - len(outRunningBalanceString)
				if spaceCount < 0 {
					spaceCount = 0
				}
				fmt.Printf("%s%s %s", strings.Repeat(" ", spaceCount), outBalanceString, outRunningBalanceString)
				fmt.Println("")
			}
		}
	}
}
