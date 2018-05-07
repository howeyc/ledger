package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/howeyc/ledger"
)

const (
	transactionDateFormat = "2006/01/02"
	displayPrecision      = 2
)

func main() {
	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(time.Hour * 24)
	var startString, endString string
	var columnWidth, transactionDepth int
	var showEmptyAccounts bool
	var columnWide bool
	var period string
	var payeeFilter string

	var ledgerFileName string

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.StringVar(&startString, "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	flag.StringVar(&endString, "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
	flag.StringVar(&period, "period", "", "Split output into periods (Monthly,Quarterly,SemiYearly,Yearly).")
	flag.StringVar(&payeeFilter, "payee", "", "Filter output to payees that contain this string.")
	flag.BoolVar(&showEmptyAccounts, "empty", false, "Show empty (zero balance) accounts.")
	flag.IntVar(&transactionDepth, "depth", -1, "Depth of transaction output (balance).")
	flag.IntVar(&columnWidth, "columns", 80, "Set a column width for output.")
	flag.BoolVar(&columnWide, "wide", false, "Wide output (same as --columns=132).")
	flag.Parse()

	if columnWidth == 80 && columnWide {
		columnWidth = 132
	}

	if len(ledgerFileName) == 0 {
		flag.Usage()
		return
	}

	parsedStartDate, tstartErr := time.Parse(transactionDateFormat, startString)
	parsedEndDate, tendErr := time.Parse(transactionDateFormat, endString)

	if tstartErr != nil || tendErr != nil {
		fmt.Println("Unable to parse start or end date string argument.")
		fmt.Println("Expected format: YYYY/MM/dd")
		return
	}

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Specify a command.")
		fmt.Println("Valid commands are:")
		fmt.Println(" bal/balance: summarize account balances")
		fmt.Println(" print: print ledger")
		fmt.Println(" reg/register: print filtered register")
		fmt.Println(" stats: ledger summary")
		return
	}

	ledgerFileReader, err := ledger.NewLedgerReader(ledgerFileName)
	if err != nil {
		fmt.Println(err)
		return
	}

	generalLedger, parseError := ledger.ParseLedger(ledgerFileReader)
	if parseError != nil {
		fmt.Printf("%s\n", parseError.Error())
		return
	}

	timeStartIndex, timeEndIndex := 0, 0
	for idx := 0; idx < len(generalLedger); idx++ {
		if generalLedger[idx].Date.After(parsedStartDate) {
			timeStartIndex = idx
			break
		}
	}
	for idx := len(generalLedger) - 1; idx >= 0; idx-- {
		if generalLedger[idx].Date.Before(parsedEndDate) {
			timeEndIndex = idx
			break
		}
	}
	generalLedger = generalLedger[timeStartIndex : timeEndIndex+1]

	origLedger := generalLedger
	generalLedger = make([]*ledger.Transaction, 0)
	for _, trans := range origLedger {
		if strings.Contains(trans.Payee, payeeFilter) {
			generalLedger = append(generalLedger, trans)
		}
	}

	containsFilterArray := args[1:]
	switch strings.ToLower(args[0]) {
	case "balance", "bal":
		if period == "" {
			PrintBalances(ledger.GetBalances(generalLedger, containsFilterArray), showEmptyAccounts, transactionDepth, columnWidth)
		} else {
			lperiod := ledger.Period(period)
			rbalances := ledger.BalancesByPeriod(generalLedger, lperiod, ledger.RangePartition)
			for rIdx, rb := range rbalances {
				if rIdx > 0 {
					fmt.Println("")
					fmt.Println(strings.Repeat("=", columnWidth))
				}
				fmt.Println(rb.Start.Format(transactionDateFormat), "-", rb.End.Format(transactionDateFormat))
				fmt.Println(strings.Repeat("=", columnWidth))
				PrintBalances(rb.Balances, showEmptyAccounts, transactionDepth, columnWidth)
			}
		}
	case "print":
		PrintLedger(generalLedger, columnWidth)
	case "register", "reg":
		if period == "" {
			PrintRegister(generalLedger, containsFilterArray, columnWidth)
		} else {
			lperiod := ledger.Period(period)
			rtrans := ledger.TransactionsByPeriod(generalLedger, lperiod)
			for rIdx, rt := range rtrans {
				if rIdx > 0 {
					fmt.Println(strings.Repeat("=", columnWidth))
				}
				fmt.Println(rt.Start.Format(transactionDateFormat), "-", rt.End.Format(transactionDateFormat))
				fmt.Println(strings.Repeat("=", columnWidth))
				PrintRegister(rt.Transactions, containsFilterArray, columnWidth)
			}
		}
	case "stats":
		PrintStats(generalLedger)
	}
}
