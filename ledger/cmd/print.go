package cmd

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/howeyc/ledger"
	"github.com/howeyc/ledger/decimal"
	"github.com/howeyc/ledger/ledger/internal/fastcolor"
	date "github.com/joyt/godate"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	transactionDateFormat = "2006/01/02"
	newLine               = "\n"
)

var startString, endString string
var columnWidth, transactionDepth int
var showEmptyAccounts bool
var columnWide bool
var period string
var payeeFilter string
var spaceStr string

func cliTransactions() ([]*ledger.Transaction, error) {
	if columnWidth == 80 && columnWide {
		columnWidth = 132
		fd := int(os.Stdout.Fd())
		if term.IsTerminal(fd) {
			tw, _, err := term.GetSize(fd)
			if err == nil {
				columnWidth = tw
			}
		}
	}
	spaceStr = strings.Repeat(" ", columnWidth)

	parsedStartDate, tstartErr := date.Parse(startString)
	parsedEndDate, tendErr := date.Parse(endString)

	if tstartErr != nil || tendErr != nil {
		return nil, errors.New("unable to parse start or end date string argument")
	}

	// include end dates' transactions too
	parsedEndDate = parsedEndDate.Add(time.Second)

	var generalLedger []*ledger.Transaction
	var parseError error
	if ledgerFilePath == "-" {
		generalLedger, parseError = ledger.ParseLedger(os.Stdin)
	} else {
		generalLedger, parseError = ledger.ParseLedgerFile(ledgerFilePath)
	}
	if parseError != nil {
		return nil, parseError
	}

	slices.SortStableFunc(generalLedger, func(a, b *ledger.Transaction) int {
		return a.Date.Compare(b.Date)
	})

	generalLedger = ledger.TransactionsInDateRange(generalLedger, parsedStartDate, parsedEndDate)

	origLedger := generalLedger
	generalLedger = make([]*ledger.Transaction, 0)
	for _, trans := range origLedger {
		if strings.Contains(trans.Payee, payeeFilter) {
			generalLedger = append(generalLedger, trans)
		}
	}

	return generalLedger, nil
}

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print [account-substring-filter]...",
	Short: "Print transactions in ledger file format",
	Run: func(_ *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}

		PrintLedger(generalLedger, args, columnWidth)
	},
}

func init() {
	rootCmd.AddCommand(printCmd)

	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(1<<63 - 1)
	printCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	printCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
	printCmd.Flags().StringVar(&payeeFilter, "payee", "", "Filter output to payees that contain this string.")
	printCmd.Flags().IntVar(&columnWidth, "columns", 80, "Set a column width for output.")
	printCmd.Flags().BoolVar(&columnWide, "wide", false, "Wide output (use terminal width).")
}

// PrintBalances prints out account balances formatted to a window set to a width of columns.
// Only shows accounts with names less than or equal to the given depth.
func PrintBalances(accountList []*ledger.Account, printZeroBalances bool, depth, columns int) {
	// Calculate widths: 10 columns for balance, rest for accountname
	if columns < 12 {
		columns = 12
		fmt.Fprintf(os.Stderr, "warning: `columns` too small, setting to %d\n", columns)
	}
	accWidth := columns - 11

	colorNeg := fastcolor.FgRed
	colorAccount := fastcolor.FgBlue
	colorReset := fastcolor.Reset

	buf := bufio.NewWriter(os.Stdout)
	overallBalance := decimal.Zero
	for _, account := range accountList {
		accDepth := strings.Count(account.Name, ":") + 1
		if accDepth == 1 {
			overallBalance = overallBalance.Add(account.Balance)
		}
		if (printZeroBalances || account.Balance.Sign() != 0) && (depth < 0 || accDepth <= depth) {
			outBalanceString := account.Balance.StringFixedBank()
			amtColor := colorReset
			if account.Balance.Sign() < 0 {
				amtColor = colorNeg
			}
			colorAccount.WriteStringFixed(buf, account.Name, accWidth, false)
			buf.WriteString(" ")
			amtColor.WriteStringFixed(buf, outBalanceString, 10, true)
			buf.WriteString(newLine)
		}
	}
	fmt.Fprintln(buf, strings.Repeat("-", columns))
	outBalanceString := overallBalance.StringFixedBank()
	amtColor := colorReset
	if overallBalance.Sign() < 0 {
		amtColor = colorNeg
	}
	colorAccount.WriteStringFixed(buf, "", accWidth, false)
	buf.WriteString(" ")
	amtColor.WriteStringFixed(buf, outBalanceString, 10, true)
	buf.WriteString(newLine)
	buf.Flush()
}

// WriteTransaction writes a transaction formatted to fit in specified column width.
func WriteTransaction(w io.StringWriter, trans *ledger.Transaction, columns int) {
	for _, c := range trans.Comments {
		w.WriteString(c)
		w.WriteString(newLine)
	}

	// Print accounts sorted by name
	slices.SortFunc(trans.AccountChanges, func(a, b ledger.Account) int {
		return strings.Compare(a.Name, b.Name)
	})

	w.WriteString(trans.Date.Format(transactionDateFormat))
	w.WriteString(spaceStr[:1])
	w.WriteString(trans.Payee)
	if len(trans.PayeeComment) > 0 {
		spaceCount := columns - 10 - utf8.RuneCountInString(trans.Payee)
		if spaceCount < 1 {
			spaceCount = 1
		}
		w.WriteString(spaceStr[:spaceCount])
		w.WriteString(trans.PayeeComment)
	}
	w.WriteString(newLine)
	for _, accChange := range trans.AccountChanges {
		outBalanceString := accChange.Balance.StringFixedBank()
		spaceCount := columns - 4 - utf8.RuneCountInString(accChange.Name) - utf8.RuneCountInString(outBalanceString)
		if spaceCount < 1 {
			spaceCount = 1
		}
		w.WriteString(spaceStr[:4])
		w.WriteString(accChange.Name)
		w.WriteString(spaceStr[:spaceCount])
		w.WriteString(outBalanceString)
		if len(accChange.Comment) > 0 {
			w.WriteString(spaceStr[:1])
			w.WriteString(accChange.Comment)
		}
		w.WriteString(newLine)
	}
	w.WriteString(newLine)
}

// PrintLedger prints all transactions as a formatted ledger file.
func PrintLedger(generalLedger []*ledger.Transaction, filterArr []string, columns int) {
	// Print transactions by date
	if len(generalLedger) > 1 {
		slices.SortStableFunc(generalLedger, func(a, b *ledger.Transaction) int {
			return a.Date.Compare(b.Date)
		})
	}

	buf := bufio.NewWriter(os.Stdout)
	for _, trans := range generalLedger {
		inFilter := len(filterArr) == 0
		for _, accChange := range trans.AccountChanges {
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
		}
		if inFilter {
			WriteTransaction(buf, trans, columns)
		}
	}
	buf.Flush()
}

// PrintRegister prints each transaction that matches the given filters.
func PrintRegister(generalLedger []*ledger.Transaction, filterArr []string, columns int) {
	// Calculate widths for variable-length part of output
	// 3 10-width columns (date, account-change, running-total)
	// 4 spaces
	if columns < 35 {
		columns = 35
		fmt.Fprintf(os.Stderr, "warning: `columns` too small, setting to %d\n", columns)
	}
	remainingWidth := columns - (10 * 3) - (4 * 1)
	col1width := remainingWidth / 3
	col2width := remainingWidth - col1width

	colorNeg := fastcolor.FgRed
	colorPayee := fastcolor.Bold
	colorAccount := fastcolor.FgBlue
	colorReset := fastcolor.Reset

	buf := bufio.NewWriter(os.Stdout)
	runningBalance := decimal.Zero
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
			if inFilter {
				runningBalance = runningBalance.Add(accChange.Balance)
				outBalanceString := accChange.Balance.StringFixedBank()
				outRunningBalanceString := runningBalance.StringFixedBank()

				balamtColor := colorReset
				if accChange.Balance.Sign() < 0 {
					balamtColor = colorNeg
				}
				runamtColor := colorReset
				if runningBalance.Sign() < 0 {
					runamtColor = colorNeg
				}

				buf.WriteString(trans.Date.Format(transactionDateFormat))
				buf.WriteString(" ")
				colorPayee.WriteStringFixed(buf, trans.Payee, col1width, false)
				buf.WriteString(" ")
				colorAccount.WriteStringFixed(buf, accChange.Name, col2width, false)
				buf.WriteString(" ")
				balamtColor.WriteStringFixed(buf, outBalanceString, 10, true)
				buf.WriteString(" ")
				runamtColor.WriteStringFixed(buf, outRunningBalanceString, 10, true)
				buf.WriteString(newLine)
			}
		}
	}
	buf.Flush()
}

// PrintCSV prints each transaction that matches the given filters in CSV format
func PrintCSV(generalLedger []*ledger.Transaction, filterArr []string) {
	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma, _ = utf8.DecodeRuneInString(fieldDelimiter)

	runningBalance := decimal.Zero
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
			if inFilter {
				runningBalance = runningBalance.Add(accChange.Balance)
				outBalanceString := accChange.Balance.StringFixedBank()
				record := []string{trans.Date.Format(transactionDateFormat),
					trans.Payee,
					accChange.Name,
					outBalanceString,
				}
				if err := csvWriter.Write(record); err != nil {
					fmt.Fprintf(os.Stderr, "error writing record to CSV: %s", err)
					return
				}
			}
		}
	}

	// Write any buffered data to the underlying writer (standard output).
	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		fmt.Fprintf(os.Stderr, "error flushing CSV buffer: %s", err)
		return
	}
}
