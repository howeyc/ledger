package main

import (
	"flag"
	"fmt"
	"os"
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

	var ledgerFileName string

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.StringVar(&startString, "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	flag.StringVar(&endString, "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
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

	ledgerFileReader, err := os.Open(ledgerFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer ledgerFileReader.Close()

	generalLedger, parseError := ledger.ParseLedger(ledgerFileReader)
	if parseError != nil {
		fmt.Printf("%s:%s\n", ledgerFileName, parseError.Error())
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

	containsFilterArray := args[1:]
	switch strings.ToLower(args[0]) {
	case "balance", "bal":
		PrintBalances(ledger.GetBalances(generalLedger, containsFilterArray), showEmptyAccounts, transactionDepth, columnWidth)
	case "print":
		PrintLedger(generalLedger, columnWidth)
	case "register", "reg":
		PrintRegister(generalLedger, containsFilterArray, columnWidth)
	case "stats":
		PrintStats(generalLedger)
	}
}
