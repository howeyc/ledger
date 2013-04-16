package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/howeyc/cli-ledger/pkg/ledger"
	"github.com/jbrukh/bayesian"
)

func main() {
	var ledgerFileName string
	var dateColumn, payeeColumn, amountColumn int
	var accountSubstring, csvFileName, csvDateFormat string
	var negateAmount bool

	flag.BoolVar(&negateAmount, "neg", false, "Negate amount column value.")
	flag.IntVar(&dateColumn, "dc", -1, "Date column number.")
	flag.IntVar(&payeeColumn, "pc", -1, "Payee column number.")
	flag.IntVar(&amountColumn, "ac", -1, "Amount column number.")
	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.StringVar(&csvDateFormat, "date-format", "01/02/2006", "Date format.")
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		fmt.Println("Specify account and import csv file.")
	} else {
		accountSubstring = args[0]
		csvFileName = args[1]
	}

	csvFileReader, err := os.Open(csvFileName)
	if err != nil {
		fmt.Println("CSV: ", err)
		return
	}
	defer csvFileReader.Close()

	ledgerFileReader, err := os.Open(ledgerFileName)
	if err != nil {
		fmt.Println("Ledger: ", err)
		return
	}
	defer ledgerFileReader.Close()

	generalLedger, parseError := ledger.ParseLedger(ledgerFileReader)
	if parseError != nil {
		fmt.Println(parseError)
		return
	}

	var matchingAccount string
	matchingAccounts := ledger.GetBalances(generalLedger, []string{accountSubstring})
	if len(matchingAccounts) < 1 {
		fmt.Println("Unable to find matching account.")
		return
	} else {
		matchingAccount = matchingAccounts[len(matchingAccounts)-1].Name
	}

	allAccounts := ledger.GetBalances(generalLedger, []string{})

	csvReader := csv.NewReader(csvFileReader)
	csvRecords, _ := csvReader.ReadAll()

	classes := make([]bayesian.Class, len(allAccounts))
	for i, bal := range allAccounts {
		classes[i] = bayesian.Class(bal.Name)
	}
	classifier := bayesian.NewClassifier(classes...)
	for _, tran := range generalLedger {
		payeeWords := strings.Split(tran.Payee, " ")
		for _, accChange := range tran.AccountChanges {
			if strings.Contains(accChange.Name, "Expense") {
				classifier.Learn(payeeWords, bayesian.Class(accChange.Name))
			}
		}
	}

	for _, record := range csvRecords[1:] {
		inputPayeeWords := strings.Split(record[payeeColumn], " ")
		csvDate, _ := time.Parse(csvDateFormat, record[dateColumn])
		fmt.Printf("%s %s\n", csvDate.Format(ledger.TransactionDateFormat), record[payeeColumn])
		fmt.Printf("   %s\n", matchingAccount)
		_, likely, _ := classifier.LogScores(inputPayeeWords)
		fmt.Printf("   %s     %s\n", classifier.Classes[likely], record[amountColumn])
	}
}
