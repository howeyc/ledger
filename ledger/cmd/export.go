package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

// PrintCSV prints each transaction that matches the given filters in CSV format
func PrintCSV(generalLedger []*ledger.Transaction, filterArr []string) {
	csvWriter := csv.NewWriter(os.Stdout)
	csvWriter.Comma, _ = utf8.DecodeRuneInString(fieldDelimiter)

	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
			if inFilter {
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

var exportType string

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Aliases: []string{"exp"},
	Use:     "export [account-substring-filter]...",
	Short:   "export to different file type",
	Run: func(cmd *cobra.Command, args []string) {
		generalLedger, err := cliTransactions(cmd)
		if err != nil {
			log.Fatalln(err)
		}
		switch exportType {
		case "csv":
			PrintCSV(generalLedger, args)
		default:
			fmt.Fprintln(os.Stderr, "unknown export type specified")
		}
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(1<<63 - 1)
	exportCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	exportCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
	exportCmd.Flags().StringVar(&payeeFilter, "payee", "", "Filter output to payees that contain this string.")
	exportCmd.Flags().StringVar(&fieldDelimiter, "delimiter", ",", "Field delimiter.")
	exportCmd.Flags().StringVar(&exportType, "type", "csv", "Export file type (csv/beancount)")
}
