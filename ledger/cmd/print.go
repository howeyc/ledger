package cmd

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/howeyc/ledger"
	"github.com/howeyc/ledger/internal/decimal"
	"github.com/spf13/cobra"
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
	}

	parsedStartDate, tstartErr := time.Parse(transactionDateFormat, startString)
	parsedEndDate, tendErr := time.Parse(transactionDateFormat, endString)

	if tstartErr != nil || tendErr != nil {
		return nil, errors.New("unable to parse start or end date string argument")
	}

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

	sort.Slice(generalLedger, func(i, j int) bool {
		return generalLedger[i].Date.Before(generalLedger[j].Date)
	})

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

	return generalLedger, nil
}

// printCmd represents the print command
var printCmd = &cobra.Command{
	Use:   "print [account-substring-filter]...",
	Short: "Print transactions in ledger file format",
	Run: func(cmd *cobra.Command, args []string) {
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
	endDate = time.Now().Add(time.Hour * 24)
	printCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	printCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
	printCmd.Flags().StringVar(&payeeFilter, "payee", "", "Filter output to payees that contain this string.")
	printCmd.Flags().IntVar(&columnWidth, "columns", 80, "Set a column width for output.")
	printCmd.Flags().BoolVar(&columnWide, "wide", false, "Wide output (same as --columns=132).")
}

// PrintBalances prints out account balances formatted to a window set to a width of columns.
// Only shows accounts with names less than or equal to the given depth.
func PrintBalances(accountList []*ledger.Account, printZeroBalances bool, depth, columns int) {
	overallBalance := decimal.Zero
	for _, account := range accountList {
		accDepth := len(strings.Split(account.Name, ":"))
		if accDepth == 1 {
			overallBalance = overallBalance.Add(account.Balance)
		}
		if (printZeroBalances || account.Balance.Sign() != 0) && (depth < 0 || accDepth <= depth) {
			outBalanceString := account.Balance.StringFixedBank()
			spaceCount := columns - utf8.RuneCountInString(account.Name) - utf8.RuneCountInString(outBalanceString)
			fmt.Printf("%s%s%s\n", account.Name, strings.Repeat(" ", spaceCount), outBalanceString)
		}
	}
	fmt.Println(strings.Repeat("-", columns))
	outBalanceString := overallBalance.StringFixedBank()
	spaceCount := columns - utf8.RuneCountInString(outBalanceString)
	fmt.Printf("%s%s\n", strings.Repeat(" ", spaceCount), outBalanceString)
}

// PrintTransaction prints a transaction formatted to fit in specified column width.
func PrintTransaction(trans *ledger.Transaction, columns int) {
	WriteTransaction(os.Stdout, trans, columns)
}

// WriteTransaction writes a transaction formatted to fit in specified column width.
func WriteTransaction(w io.Writer, trans *ledger.Transaction, columns int) {
	for _, c := range trans.Comments {
		fmt.Fprintln(w, c)
	}

	// Print accounts sorted by name
	sort.Slice(trans.AccountChanges, func(i, j int) bool {
		return trans.AccountChanges[i].Name < trans.AccountChanges[j].Name
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
		sort.Slice(generalLedger, func(i, j int) bool {
			return generalLedger[i].Date.Before(generalLedger[j].Date)
		})
	}

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
			PrintTransaction(trans, columns)
		}
	}
}

// PrintRegister prints each transaction that matches the given filters.
func PrintRegister(generalLedger []*ledger.Transaction, filterArr []string, columns int) {
	// Calculate widths for variable-length part of output
	// 3 10-width columns (date, account-change, running-total)
	// 4 spaces
	remainingWidth := columns - (10 * 3) - (4 * 1)
	col1width := remainingWidth / 3
	col2width := remainingWidth - col1width

	formatString := fmt.Sprintf("%%-10.10s %%-%[1]d.%[1]ds %%-%[2]d.%[2]ds %%10.10s %%10.10s\n", col1width, col2width)

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
				fmt.Printf(formatString,
					trans.Date.Format(transactionDateFormat),
					trans.Payee,
					accChange.Name,
					outBalanceString,
					outRunningBalanceString)
			}
		}
	}
}
