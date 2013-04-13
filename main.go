package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

const TransactionDateFormat = "2006/01/02"

func main() {
	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now()
	var startString, endString string

	var sortTransactions bool
	var ledgerFileName string

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name.")
	flag.StringVar(&startString, "s", startDate.Format(TransactionDateFormat), "Start date of transaction processing.")
	flag.StringVar(&endString, "e", endDate.Format(TransactionDateFormat), "Start date of transaction processing.")
	flag.BoolVar(&sortTransactions, "sort", true, "Sort transactions on output.")
	flag.Parse()

	// TODO(chris): Handle parse of arg errors
	startDate, _ = time.Parse(TransactionDateFormat, startString)
	endDate, _ = time.Parse(TransactionDateFormat, endString)

	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Specify a command.")
		return
	}

	ledgerFileReader, err := os.Open(ledgerFileName)
	defer ledgerFileReader.Close()
	if err != nil {
		fmt.Println(err)
		return
	}

	generalLedger, parseError := parseLedger(ledgerFileReader)
	if parseError != nil {
		fmt.Println(parseError)
		return
	}

	switch strings.ToLower(args[0]) {
	case "balance":
	case "print":
		printLedger(os.Stdout, generalLedger)
	}
}
