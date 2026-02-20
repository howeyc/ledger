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
	"github.com/howeyc/ledger/ledger/cmd/internal/import/camt"
	"github.com/howeyc/ledger/ledger/cmd/internal/import/qfx"
	"github.com/howeyc/ledger/ledger/cmd/internal/import/qif"
	"github.com/jbrukh/bayesian"
	"github.com/shopspring/decimal"
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
var overrideCurrency string

type Importer struct {
	filename        string
	reader          *os.File
	decScale        decimal.Decimal
	matchingAccount string
	generalLedger   []*ledger.Transaction
	classifier      *bayesian.Classifier
}

func NewImporter(accountSubstring, filename string) *Importer {
	imp := Importer{
		filename: filename,
		decScale: decimal.NewFromFloat(scaleFactor),
	}

	fileReader, err := os.Open(filename)
	if err != nil {
		fmt.Println("CSV: ", err)
		return nil
	}
	imp.reader = fileReader

	// If a ledger file path is provided, load it and train the classifier.
	// Otherwise, skip loading and prediction will fall back to "unknown:unknown".
	if ledgerFilePath != "" {
		generalLedger, parseError := ledger.ParseLedgerFile(ledgerFilePath)
		if parseError != nil {
			fmt.Printf("%s:%s\n", ledgerFilePath, parseError.Error())
			return nil
		}
		imp.generalLedger = generalLedger

		matchingAccount, err := imp.findMatchingAccount(accountSubstring)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		imp.matchingAccount = matchingAccount

		imp.classifier = imp.trainClassifier(imp.matchingAccount)
	} else {
		imp.matchingAccount = accountSubstring
	}

	return &imp
}

func (imp *Importer) Close() {
	imp.reader.Close()
}

func (imp *Importer) trainClassifier(matchingAccount string) *bayesian.Classifier {
	allAccounts := ledger.GetBalances(imp.generalLedger, []string{})
	uniqueAccounts := make(map[string]bool)
	for _, acc := range allAccounts {
		if ok, _ := uniqueAccounts[acc.Name]; !ok {
			uniqueAccounts[acc.Name] = true
		}
	}

	classes := []bayesian.Class{}
	for name := range uniqueAccounts {
		classes = append(classes, bayesian.Class(name))
	}

	classifier := bayesian.NewClassifier(classes...)
	for _, tran := range imp.generalLedger {
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

func (imp *Importer) predictAccount(inputPayeeWords []string) string {
	if imp.classifier == nil {
		return "unknown:unknown"
	}

	// Classify into expense account

	// Find the highest and second highest scores
	highScore1 := math.Inf(-1)
	highScore2 := math.Inf(-1)
	matchIdx := 0
	scores, _, _ := imp.classifier.LogScores(inputPayeeWords)
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
		return string(imp.classifier.Classes[matchIdx])
	} else {
		return "unknown:unknown"
	}
}

func (imp *Importer) findMatchingAccount(accountSubstring string) (string, error) {
	var matchingAccount string
	matchingAccounts := ledger.GetBalances(imp.generalLedger, []string{accountSubstring})
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

func (imp *Importer) importCSV() {
	csvReader := csv.NewReader(imp.reader)
	csvReader.Comma, _ = utf8.DecodeRuneInString(fieldDelimiter)
	csvRecords, cerr := csvReader.ReadAll()
	if cerr != nil {
		fmt.Println("CSV parse error:", cerr.Error())
		return
	}

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
	csvAccount := ledger.Account{Name: imp.matchingAccount, Balance: decimal.Zero}
	for _, record := range csvRecords[1:] {
		inputPayeeWords := strings.Fields(record[payeeColumn])
		csvDate, _ := time.Parse(csvDateFormat, record[dateColumn])
		if allowMatching || !imp.existingTransaction(csvDate, record[payeeColumn]) {
			expenseAccount.Name = imp.predictAccount(inputPayeeWords)

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
			expenseAccount.Balance = expenseAccount.Balance.Mul(imp.decScale)

			// Csv amount is the negative of the expense amount
			csvAccount.Balance = expenseAccount.Balance.Neg()

			trans := &ledger.Transaction{Date: csvDate, Payee: record[payeeColumn]}
			trans.AccountChanges = []ledger.Account{csvAccount, expenseAccount}

			if overrideCurrency != "" {
				for i := range trans.AccountChanges {
					trans.AccountChanges[i].Currency = overrideCurrency
				}
			}
			if commentColumn >= 0 && record[commentColumn] != "" {
				trans.Comments = []string{";" + record[commentColumn]}
			}
			WriteTransaction(os.Stdout, trans, 80)
		}
	}
}

func (imp *Importer) importCamt() {
	entries, err := camt.ParseCamt(imp.reader)
	if err != nil {
		fmt.Println("CAMT parse error:", err.Error())
		return
	}

	expenseAccount := ledger.Account{Name: "unknown:unknown", Balance: decimal.Zero}
	camtAccount := ledger.Account{Name: imp.matchingAccount, Balance: decimal.Zero}
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

		expenseAccount.Name = imp.predictAccount(inputPayeeWords)
		expenseAccount.Balance = amount

		// Determine if debit
		isDebit := entry.CdtDbtInd == "DBIT"
		if !isDebit {
			expenseAccount.Balance = expenseAccount.Balance.Neg()
		}

		// Apply scale
		expenseAccount.Balance = expenseAccount.Balance.Mul(imp.decScale)

		// Csv amount is the negative of the expense amount
		camtAccount.Balance = expenseAccount.Balance.Neg()

		trans := &ledger.Transaction{Date: dateTime, Payee: payee}
		trans.AccountChanges = []ledger.Account{camtAccount, expenseAccount}
		if overrideCurrency != "" {
			for i := range trans.AccountChanges {
				trans.AccountChanges[i].Currency = overrideCurrency
			}
		} else if entry.Amt.Ccy != "" {
			for i := range trans.AccountChanges {
				trans.AccountChanges[i].Currency = entry.Amt.Ccy
			}
		}
		if reference != "" {
			trans.Comments = []string{";" + reference}
		}
		WriteTransaction(os.Stdout, trans, 80)
	}
}

func (imp *Importer) importQIF() {
	entries, err := qif.ParseQIF(imp.reader)
	if err != nil {
		fmt.Println("QIF parse error:", err.Error())
		return
	}

	expenseAccount := ledger.Account{Name: "unknown:unknown", Balance: decimal.Zero}
	qifAccount := ledger.Account{Name: imp.matchingAccount, Balance: decimal.Zero}
	for _, entry := range entries {
		// Parse date (QIF dates are often locale-specific; assume mm/dd/yyyy here)
		dateTime, err := time.Parse("01/02/2006", entry.Date)
		if err != nil {
			// Try an alternate common QIF date format (dd/mm/yyyy)
			dateTime, err = time.Parse("02/01/2006", entry.Date)
			if err != nil {
				fmt.Println("QIF date parse error:", err.Error())
				continue
			}
		}

		// Parse amount
		amount, err := decimal.NewFromString(entry.Amount)
		if err != nil {
			fmt.Println("QIF amount parse error:", err.Error())
			continue
		}

		payee := entry.Payee
		inputPayeeWords := strings.Fields(payee)

		expenseAccount.Name = imp.predictAccount(inputPayeeWords)
		expenseAccount.Balance = amount

		// Apply scale
		expenseAccount.Balance = expenseAccount.Balance.Mul(imp.decScale)

		// Account side is the opposite of expense
		qifAccount.Balance = expenseAccount.Balance.Neg()

		trans := &ledger.Transaction{Date: dateTime, Payee: payee}
		trans.AccountChanges = []ledger.Account{qifAccount, expenseAccount}
		if overrideCurrency != "" {
			for i := range trans.AccountChanges {
				trans.AccountChanges[i].Currency = overrideCurrency
			}
		}
		if len(entry.RawLines) > 0 {
			// Join all raw lines except header/type line
			comment := strings.Join(entry.RawLines, " ")
			trans.Comments = []string{";" + comment}
		}
		WriteTransaction(os.Stdout, trans, 80)
	}
}

func (imp *Importer) importQFX() {
	entries, err := qfx.ParseQFX(imp.reader)
	if err != nil {
		fmt.Println("QFX parse error:", err.Error())
		return
	}

	expenseAccount := ledger.Account{Name: "unknown:unknown", Balance: decimal.Zero}
	qfxAccount := ledger.Account{Name: imp.matchingAccount, Balance: decimal.Zero}
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

		expenseAccount.Name = imp.predictAccount(inputPayeeWords)
		expenseAccount.Balance = amount

		// Apply scale
		expenseAccount.Balance = expenseAccount.Balance.Mul(imp.decScale)

		// Account side is the opposite of expense
		qfxAccount.Balance = expenseAccount.Balance.Neg()

		trans := &ledger.Transaction{Date: dateTime, Payee: payee}
		trans.AccountChanges = []ledger.Account{qfxAccount, expenseAccount}
		if overrideCurrency != "" {
			for i := range trans.AccountChanges {
				trans.AccountChanges[i].Currency = overrideCurrency
			}
		}
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

		imp := NewImporter(accountSubstring, fileName)
		defer imp.Close()

		lower := strings.ToLower(fileName)
		if strings.HasSuffix(lower, ".xml") {
			imp.importCamt()
		} else if strings.HasSuffix(lower, ".qfx") || strings.HasSuffix(lower, ".ofx") {
			imp.importQFX()
		} else if strings.HasSuffix(lower, ".qif") {
			imp.importQIF()
		} else {
			imp.importCSV()
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
	importCmd.Flags().StringVar(&overrideCurrency, "override-currency", "", "Override detected currency for imported transactions.")
}

func (imp *Importer) existingTransaction(transDate time.Time, payee string) bool {
	for _, trans := range imp.generalLedger {
		if trans.Date == transDate && strings.TrimSpace(trans.Payee) == strings.TrimSpace(payee) {
			return true
		}
	}
	return false
}
