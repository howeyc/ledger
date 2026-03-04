package cmd

import (
	"encoding/csv"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/howeyc/ledger"
	"github.com/howeyc/ledger/decimal"
	"github.com/howeyc/ledger/ledger/camt"
	"github.com/howeyc/ledger/ledger/qfx"
	"github.com/jbrukh/bayesian"
	"github.com/spf13/cobra"
)

var (
	ErrNoMatchingAccount = errors.New("Unable to find matching account.")
)

var csvDateFormat string
var negateAmount bool
var allowMatching bool
var fieldDelimiter string
var scaleFactor float64

func trainClassifier(generalLedger []*ledger.Transaction, matchingAccount string) *bayesian.Classifier {
	allAccounts := ledger.GetBalances(generalLedger, []string{})
	classes := make([]bayesian.Class, len(allAccounts))
	for i, bal := range allAccounts {
		classes[i] = bayesian.Class(bal.Name)
	}
	classifier := bayesian.NewClassifier(classes...)
	for _, tran := range generalLedger {
		payeeWords := strings.Fields(tran.Payee)
		// learn accounts names (except matchingAccount) for transactions where matchingAccount is present
		learnName := false
		for _, accChange := range tran.AccountChanges {
			if accChange.Name == matchingAccount {
				learnName = true
				break
			}
		}
		if learnName {
			for _, accChange := range tran.AccountChanges {
				if accChange.Name != matchingAccount {
					classifier.Learn(payeeWords, bayesian.Class(accChange.Name))
				}
			}
		}
	}

	return classifier
}

func predictAccount(classifier *bayesian.Classifier, inputPayeeWords []string) string {
	// Classify into expense account

	// Find the highest and second highest scores
	highScore1 := math.Inf(-1)
	highScore2 := math.Inf(-1)
	matchIdx := 0
	scores, _, _ := classifier.LogScores(inputPayeeWords)
	for j, score := range scores {
		if score > highScore1 {
			highScore2 = highScore1
			highScore1 = score
			matchIdx = j
		}
	}
	// If the difference between the highest and second highest scores is greater than 10
	// then it indicates that highscore is a high confidence match
	if highScore1-highScore2 > 10 {
		return string(classifier.Classes[matchIdx])
	} else {
		return "unknown:unknown"
	}
}

func findMatchingAccount(generalLedger []*ledger.Transaction, accountSubstring string) (string, error) {
	var matchingAccount string
	matchingAccounts := ledger.GetBalances(generalLedger, []string{accountSubstring})
	if len(matchingAccounts) < 1 {
		return "", ErrNoMatchingAccount
	}
	for _, m := range matchingAccounts {
		if strings.EqualFold(m.Name, accountSubstring) {
			matchingAccount = m.Name
			break
		}
	}
	if matchingAccount == "" {
		matchingAccount = matchingAccounts[len(matchingAccounts)-1].Name
	}

	return matchingAccount, nil
}

func importCSV(accountSubstring, csvFileName string) {
	decScale := decimal.NewFromFloat(scaleFactor)

	csvFileReader, err := os.Open(csvFileName)
	if err != nil {
		fmt.Println("CSV: ", err)
		return
	}
	defer csvFileReader.Close()

	generalLedger, parseError := ledger.ParseLedgerFile(ledgerFilePath)
	if parseError != nil {
		fmt.Printf("%s:%s\n", ledgerFilePath, parseError.Error())
		return
	}

	matchingAccount, err := findMatchingAccount(generalLedger, accountSubstring)
	if err != nil {
		fmt.Println(err)
		return
	}

	csvReader := csv.NewReader(csvFileReader)
	csvReader.Comma, _ = utf8.DecodeRuneInString(fieldDelimiter)
	csvRecords, cerr := csvReader.ReadAll()
	if cerr != nil {
		fmt.Println("CSV parse error:", cerr.Error())
		return
	}

	classifier := trainClassifier(generalLedger, matchingAccount)

	// Find columns from header
	var dateColumn, payeeColumn, amountColumn, commentColumn int
	dateColumn, payeeColumn, amountColumn, commentColumn = -1, -1, -1, -1
	for fieldIndex, fieldName := range csvRecords[0] {
		fieldName = strings.ToLower(fieldName)
		if strings.Contains(fieldName, "date") {
			dateColumn = fieldIndex
		} else if strings.Contains(fieldName, "description") {
			payeeColumn = fieldIndex
		} else if strings.Contains(fieldName, "payee") {
			payeeColumn = fieldIndex
		} else if strings.Contains(fieldName, "amount") {
			amountColumn = fieldIndex
		} else if strings.Contains(fieldName, "expense") {
			amountColumn = fieldIndex
		} else if strings.Contains(fieldName, "note") {
			commentColumn = fieldIndex
		} else if strings.Contains(fieldName, "comment") {
			commentColumn = fieldIndex
		}
	}

	if dateColumn < 0 || payeeColumn < 0 || amountColumn < 0 {
		fmt.Println("Unable to find columns required from header field names.")
		return
	}

	expenseAccount := ledger.Account{Name: "unknown:unknown", Balance: decimal.Zero}
	csvAccount := ledger.Account{Name: matchingAccount, Balance: decimal.Zero}
	for _, record := range csvRecords[1:] {
		inputPayeeWords := strings.Fields(record[payeeColumn])
		csvDate, _ := time.Parse(csvDateFormat, record[dateColumn])
		if allowMatching || !existingTransaction(generalLedger, csvDate, record[payeeColumn]) {
			expenseAccount.Name = predictAccount(classifier, inputPayeeWords)

			// Parse error, set to zero
			if dec, derr := decimal.NewFromString(record[amountColumn]); derr != nil {
				expenseAccount.Balance = decimal.Zero
			} else {
				expenseAccount.Balance = dec
			}

			// Negate amount if required
			if negateAmount {
				expenseAccount.Balance = expenseAccount.Balance.Neg()
			}

			// Apply scale
			expenseAccount.Balance = expenseAccount.Balance.Mul(decScale)

			// Csv amount is the negative of the expense amount
			csvAccount.Balance = expenseAccount.Balance.Neg()

			// Create valid transaction for print in ledger format
			trans := &ledger.Transaction{Date: csvDate, Payee: record[payeeColumn]}
			trans.AccountChanges = []ledger.Account{csvAccount, expenseAccount}

			// Comment
			if commentColumn >= 0 && record[commentColumn] != "" {
				trans.Comments = []string{";" + record[commentColumn]}
			}
			WriteTransaction(os.Stdout, trans, 80)
		}
	}
}

func importCamt(accountSubstring, camtFileName string) {
	decScale := decimal.NewFromFloat(scaleFactor)

	fileReader, err := os.Open(camtFileName)
	if err != nil {
		fmt.Println("CAMT: ", err, camtFileName)
		return
	}
	defer fileReader.Close()

	generalLedger, parseError := ledger.ParseLedgerFile(ledgerFilePath)
	if parseError != nil {
		fmt.Printf("%s:%s\n", ledgerFilePath, parseError.Error())
		return
	}

	matchingAccount, err := findMatchingAccount(generalLedger, accountSubstring)
	if err != nil {
		fmt.Println(err)
		return
	}

	classifier := trainClassifier(generalLedger, matchingAccount)

	entries, err := camt.ParseCamt(fileReader)
	if err != nil {
		fmt.Println("CAMT parse error:", err.Error())
		return
	}

	expenseAccount := ledger.Account{Name: "unknown:unknown", Balance: decimal.Zero}
	camtAccount := ledger.Account{Name: matchingAccount, Balance: decimal.Zero}
	for _, entry := range entries {
		dateTime, err := time.Parse(time.RFC3339, entry.BookgDt.DtTm)
		if err != nil {
			// Try another format if RFC3339 fails
			dateTime, err = time.Parse("2006-01-02T15:04:05.999999-07:00", entry.BookgDt.DtTm)
			if err != nil {
				fmt.Println("CAMT parse error:", err.Error())
			}
		}

		// Parse amount
		amount, err := decimal.NewFromString(entry.Amt.Value)
		if err != nil {
			fmt.Println("CAMT parse error:", err.Error())
		}

		// Get reference and payee
		reference := entry.BkTxCd.Prtry.Cd
		payee := ""

		// Extract payee from entry details if available
		if entry.NtryDtls != nil && entry.NtryDtls.TxDtls.RltdPties.Cdtr != nil {
			payee = entry.NtryDtls.TxDtls.RltdPties.Cdtr.Pty.Nm
		} else {
			// Use additional entry info as fallback
			payee = entry.AddtlNtryInf
		}
		inputPayeeWords := strings.Fields(payee)

		expenseAccount.Name = predictAccount(classifier, inputPayeeWords)
		expenseAccount.Balance = amount

		// Determine if debit
		isDebit := entry.CdtDbtInd == "DBIT"
		if !isDebit {
			expenseAccount.Balance = expenseAccount.Balance.Neg()
		}

		// Apply scale
		expenseAccount.Balance = expenseAccount.Balance.Mul(decScale)

		// Csv amount is the negative of the expense amount
		camtAccount.Balance = expenseAccount.Balance.Neg()

		// Create valid transaction for print in ledger format
		trans := &ledger.Transaction{Date: dateTime, Payee: payee}
		trans.AccountChanges = []ledger.Account{camtAccount, expenseAccount}

		// Comment
		if reference != "" {
			trans.Comments = []string{";" + reference}
		}
		WriteTransaction(os.Stdout, trans, 80)
	}
}

func importQFX(accountSubstring, qfxFileName string) {
	decScale := decimal.NewFromFloat(scaleFactor)

	fileReader, err := os.Open(qfxFileName)
	if err != nil {
		fmt.Println("QFX: ", err, qfxFileName)
		return
	}
	defer fileReader.Close()

	generalLedger, parseError := ledger.ParseLedgerFile(ledgerFilePath)
	if parseError != nil {
		fmt.Printf("%s:%s\n", ledgerFilePath, parseError.Error())
		return
	}

	matchingAccount, err := findMatchingAccount(generalLedger, accountSubstring)
	if err != nil {
		fmt.Println(err)
		return
	}

	classifier := trainClassifier(generalLedger, matchingAccount)

	entries, err := qfx.ParseQFX(fileReader)
	if err != nil {
		fmt.Println("QFX parse error:", err.Error())
		return
	}

	expenseAccount := ledger.Account{Name: "unknown:unknown", Balance: decimal.Zero}
	qfxAccount := ledger.Account{Name: matchingAccount, Balance: decimal.Zero}
	for _, entry := range entries {
		// QFX DTPOSTED is typically YYYYMMDDHHMMSS.XXX; we only care about the date.
		// Take the first 8 characters as YYYYMMDD.
		dateStr := entry.DtPosted
		if len(dateStr) >= 8 {
			dateStr = dateStr[:8]
		}
		dateTime, err := time.Parse("20060102", dateStr)
		if err != nil {
			fmt.Println("QFX date parse error:", err.Error())
			continue
		}

		// Parse amount
		amount, err := decimal.NewFromString(entry.TrnAmt)
		if err != nil {
			fmt.Println("QFX amount parse error:", err.Error())
			continue
		}

		payee := entry.Memo
		inputPayeeWords := strings.Fields(payee)

		expenseAccount.Name = predictAccount(classifier, inputPayeeWords)
		expenseAccount.Balance = amount

		// Apply scale
		expenseAccount.Balance = expenseAccount.Balance.Mul(decScale)

		// Account side is the opposite of expense
		qfxAccount.Balance = expenseAccount.Balance.Neg()

		// Create valid transaction for print in ledger format
		trans := &ledger.Transaction{Date: dateTime, Payee: payee}
		trans.AccountChanges = []ledger.Account{qfxAccount, expenseAccount}

		// Comment with FITID if present
		if entry.FitID != "" {
			trans.Comments = []string{";" + entry.FitID}
		}
		WriteTransaction(os.Stdout, trans, 80)
	}
}

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <account-substring> <csv-file>",
	Args:  cobra.ExactArgs(2),
	Short: "Import transactions from csv to ledger format",
	Run: func(_ *cobra.Command, args []string) {
		accountSubstring := args[0]
		fileName := args[1]

		lower := strings.ToLower(fileName)
		if strings.HasSuffix(lower, ".xml") {
			importCamt(accountSubstring, fileName)
		} else if strings.HasSuffix(lower, ".qfx") || strings.HasSuffix(lower, ".ofx") {
			importQFX(accountSubstring, fileName)
		} else {
			importCSV(accountSubstring, fileName)
		}

	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	importCmd.Flags().BoolVar(&negateAmount, "neg", false, "Negate amount column value.")
	importCmd.Flags().BoolVar(&allowMatching, "allow-matching", false, "Have output include imported transactions that\nmatch existing ledger transactions.")
	importCmd.Flags().Float64Var(&scaleFactor, "scale", 1.0, "Scale factor to multiply against every imported amount.")
	importCmd.Flags().StringVar(&csvDateFormat, "date-format", "01/02/2006", "Date format.")
	importCmd.Flags().StringVar(&fieldDelimiter, "delimiter", ",", "Field delimiter.")
}

func existingTransaction(generalLedger []*ledger.Transaction, transDate time.Time, payee string) bool {
	for _, trans := range generalLedger {
		if trans.Date == transDate && strings.TrimSpace(trans.Payee) == strings.TrimSpace(payee) {
			return true
		}
	}
	return false
}
