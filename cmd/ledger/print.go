package main

import (
	"fmt"
	"math"
	"math/big"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/howeyc/ledger"
)

// PrintStats prints out statistics of the ledger
func PrintStats(generalLedger []*ledger.Transaction) {
	if len(generalLedger) < 1 {
		fmt.Println("Empty ledger.")
		return
	}
	startDate := generalLedger[0].Date
	endDate := generalLedger[len(generalLedger)-1].Date

	payees := make(map[string]struct{})
	accounts := make(map[string]struct{})

	for _, trans := range generalLedger {
		payees[trans.Payee] = struct{}{}
		for _, account := range trans.AccountChanges {
			accounts[account.Name] = struct{}{}
		}
	}

	days := math.Floor(endDate.Sub(startDate).Hours() / 24)

	fmt.Printf("%-25s : %s to %s (%s)\n", "Transactions span", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), DurationInWords(endDate.Sub(startDate)))
	fmt.Printf("%-25s : %s\n", "Since last post", DurationInWords(time.Since(endDate)))
	fmt.Printf("%-25s : %d, (%.1f per day)\n", "Transactions", len(generalLedger), float64(len(generalLedger))/days)
	fmt.Printf("%-25s : %d\n", "Payees", len(payees))
	fmt.Printf("%-25s : %d\n", "Referenced Accounts", len(accounts))
}

// PrintBalances prints out account balances formatted to a window set to a width of columns.
// Only shows accounts with names less than or equal to the given depth.
func PrintBalances(accountList []*ledger.Account, printZeroBalances bool, depth, columns int) {
	overallBalance := new(big.Rat)
	for _, account := range accountList {
		accDepth := len(strings.Split(account.Name, ":"))
		if accDepth == 1 {
			overallBalance.Add(overallBalance, account.Balance)
		}
		if (printZeroBalances || account.Balance.Sign() != 0) && (depth < 0 || accDepth <= depth) {
			outBalanceString := account.Balance.FloatString(displayPrecision)
			spaceCount := columns - utf8.RuneCountInString(account.Name) - utf8.RuneCountInString(outBalanceString)
			fmt.Printf("%s%s%s\n", account.Name, strings.Repeat(" ", spaceCount), outBalanceString)
		}
	}
	fmt.Println(strings.Repeat("-", columns))
	outBalanceString := overallBalance.FloatString(displayPrecision)
	spaceCount := columns - utf8.RuneCountInString(outBalanceString)
	fmt.Printf("%s%s\n", strings.Repeat(" ", spaceCount), outBalanceString)
}

// PrintTransaction prints a transaction formatted to fit in specified column width.
func PrintTransaction(trans *ledger.Transaction, columns int) {
	fmt.Printf("%s %s\n", trans.Date.Format(transactionDateFormat), trans.Payee)
	for _, accChange := range trans.AccountChanges {
		outBalanceString := accChange.Balance.FloatString(displayPrecision)
		spaceCount := columns - 4 - utf8.RuneCountInString(accChange.Name) - utf8.RuneCountInString(outBalanceString)
		fmt.Printf("    %s%s%s\n", accChange.Name, strings.Repeat(" ", spaceCount), outBalanceString)
	}
	fmt.Println("")
}

// PrintLedger prints all transactions as a formatted ledger file.
func PrintLedger(generalLedger []*ledger.Transaction, columns int) {
	for _, trans := range generalLedger {
		PrintTransaction(trans, columns)
	}
}

// PrintRegister prints each transaction that matches the given filters.
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
				writtenBytes, _ := fmt.Printf("%s %s", trans.Date.Format(transactionDateFormat), trans.Payee)
				outBalanceString := accChange.Balance.FloatString(displayPrecision)
				outRunningBalanceString := runningBalance.FloatString(displayPrecision)
				spaceCount := columns - writtenBytes - 2 - utf8.RuneCountInString(outBalanceString) - utf8.RuneCountInString(outRunningBalanceString)
				if spaceCount < 0 {
					spaceCount = 0
				}
				fmt.Printf("%s%s %s", strings.Repeat(" ", spaceCount), outBalanceString, outRunningBalanceString)
				fmt.Println("")
			}
		}
	}
}
