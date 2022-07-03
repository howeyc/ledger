package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Aliases: []string{"reg"},
	Use:     "register [account-substring-filter]...",
	Short:   "Print register of transactions",
	Run: func(cmd *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}
		if period == "" {
			PrintRegister(generalLedger, args, columnWidth)
		} else {
			lperiod := ledger.Period(period)
			rtrans := ledger.TransactionsByPeriod(generalLedger, lperiod)
			for rIdx, rt := range rtrans {
				if len(rt.Transactions) < 1 {
					continue
				}

				if rIdx > 0 {
					fmt.Println(strings.Repeat("=", columnWidth))
				}
				fmt.Println(rt.Start.Format(transactionDateFormat), "-", rt.End.Format(transactionDateFormat))
				fmt.Println(strings.Repeat("=", columnWidth))
				PrintRegister(rt.Transactions, args, columnWidth)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)

	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(time.Hour * 24)
	registerCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	registerCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
	registerCmd.Flags().StringVar(&payeeFilter, "payee", "", "Filter output to payees that contain this string.")
	registerCmd.Flags().IntVar(&columnWidth, "columns", 80, "Set a column width for output.")
	registerCmd.Flags().BoolVar(&columnWide, "wide", false, "Wide output (same as --columns=132).")

	registerCmd.Flags().StringVar(&period, "period", "", "Split output into periods (Monthly,Quarterly,SemiYearly,Yearly).")
}
