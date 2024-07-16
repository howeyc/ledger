package cmd

import (
	"bytes"
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

	"github.com/fatih/color"
	"github.com/howeyc/ledger"
	"github.com/howeyc/ledger/decimal"
	date "github.com/joyt/godate"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

const (
	transactionDateFormat = "2006/01/02"
)

var startString, endString string
var columnWidth, transactionDepth int
var showEmptyAccounts bool
var columnWide bool
var period string
var payeeFilter string

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
	formatAcc := fmt.Sprintf("%%-%[1]d.%[1]ds", accWidth)
	formatAmt := "%10.10s"

	colorNeg := color.New(color.FgRed)
	colorAccount := color.New(color.FgBlue)
	colorReset := color.New(color.Reset)

	var buf bytes.Buffer
	overallBalance := decimal.Zero
	for _, account := range accountList {
		accDepth := len(strings.Split(account.Name, ":"))
		if accDepth == 1 {
			overallBalance = overallBalance.Add(account.Balance)
		}
		if (printZeroBalances || account.Balance.Sign() != 0) && (depth < 0 || accDepth <= depth) {
			outBalanceString := account.Balance.StringFixedBank()
			amtColor := colorReset
			if account.Balance.Sign() < 0 {
				amtColor = colorNeg
			}
			colorAccount.Fprintf(&buf, formatAcc, account.Name)
			fmt.Fprint(&buf, " ")
			amtColor.Fprintf(&buf, formatAmt, outBalanceString)
			fmt.Fprintln(&buf, "")
		}
	}
	fmt.Fprintln(&buf, strings.Repeat("-", columns))
	outBalanceString := overallBalance.StringFixedBank()
	amtColor := colorReset
	if overallBalance.Sign() < 0 {
		amtColor = colorNeg
	}
	colorAccount.Fprintf(&buf, formatAcc, "")
	fmt.Fprint(&buf, " ")
	amtColor.Fprintf(&buf, formatAmt, outBalanceString)
	fmt.Fprintln(&buf, "")
	io.Copy(os.Stdout, &buf)
}

// WriteTransaction writes a transaction formatted to fit in specified column width.
func WriteTransaction(w io.Writer, trans *ledger.Transaction, columns int) {
	for _, c := range trans.Comments {
		fmt.Fprintln(w, c)
	}

	// Print accounts sorted by name
	slices.SortFunc(trans.AccountChanges, func(a, b ledger.Account) int {
		return strings.Compare(a.Name, b.Name)
	})

	fmt.Fprintf(w, "%s %s", trans.Date.Format(transactionDateFormat), trans.Payee)
	if len(trans.PayeeComment) > 0 {
		spaceCount := columns - 10 - utf8.RuneCountInString(trans.Payee)
		if spaceCount < 1 {
			spaceCount = 1
		}
		fmt.Fprintf(w, "%s%s", strings.Repeat(" ", spaceCount), trans.PayeeComment)
	}
	fmt.Fprintln(w, "")
	for _, accChange := range trans.AccountChanges {
		outBalanceString := accChange.Balance.StringFixedBank()
		spaceCount := columns - 4 - utf8.RuneCountInString(accChange.Name) - utf8.RuneCountInString(outBalanceString)
		if spaceCount < 1 {
			spaceCount = 1
		}
		fmt.Fprintf(w, "    %s%s%s", accChange.Name, strings.Repeat(" ", spaceCount), outBalanceString)
		if len(accChange.Comment) > 0 {
			fmt.Fprintf(w, " %s", accChange.Comment)
		}
		fmt.Fprintln(w, "")
	}
	fmt.Fprintln(w, "")
}

// PrintLedger prints all transactions as a formatted ledger file.
func PrintLedger(generalLedger []*ledger.Transaction, filterArr []string, columns int) {
	// Print transactions by date
	if len(generalLedger) > 1 {
		slices.SortStableFunc(generalLedger, func(a, b *ledger.Transaction) int {
			return a.Date.Compare(b.Date)
		})
	}

	var buf bytes.Buffer
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
			WriteTransaction(&buf, trans, columns)
		}
	}
	io.Copy(os.Stdout, &buf)
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
	formatDate := "%-10.10s"
	formatAmount := "%10.10s"
	formatPayee := fmt.Sprintf("%%-%[1]d.%[1]ds", col1width)
	formatAccount := fmt.Sprintf("%%-%[1]d.%[1]ds", col2width)

	colorNeg := color.New(color.FgRed)
	colorPayee := color.New(color.Bold)
	colorAccount := color.New(color.FgBlue)
	colorReset := color.New(color.Reset)

	var buf bytes.Buffer
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

				fmt.Fprintf(&buf, formatDate, trans.Date.Format(transactionDateFormat))
				buf.WriteString(" ")
				colorPayee.Fprintf(&buf, formatPayee, trans.Payee)
				buf.WriteString(" ")
				colorAccount.Fprintf(&buf, formatAccount, accChange.Name)
				buf.WriteString(" ")
				balamtColor.Fprintf(&buf, formatAmount, outBalanceString)
				buf.WriteString(" ")
				runamtColor.Fprintf(&buf, formatAmount, outRunningBalanceString)
				fmt.Fprintln(&buf, "")
			}
		}
	}
	io.Copy(os.Stdout, &buf)
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
