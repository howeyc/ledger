package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Aliases: []string{"exp"},
	Use:     "export [account-substring-filter]...",
	Short:   "export to CSV",
	Run: func(_ *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}
		PrintCSV(generalLedger, args)
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
}
