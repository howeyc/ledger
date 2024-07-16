package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Aliases: []string{"bal"},
	Use:     "balance [account-substring-filter]...",
	Short:   "Print account balances",
	Run: func(_ *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}
		if period == "" {
			PrintBalances(ledger.GetBalances(generalLedger, args), showEmptyAccounts, transactionDepth, columnWidth)
		} else {
			lperiod := ledger.Period(period)
			rtrans := ledger.TransactionsByPeriod(generalLedger, lperiod)
			for rIdx, rt := range rtrans {
				balances := ledger.GetBalances(rt.Transactions, args)
				if len(balances) < 1 {
					continue
				}

				if rIdx > 0 {
					fmt.Println("")
					fmt.Println(strings.Repeat("=", columnWidth))
				}
				fmt.Println(rt.Start.Format(transactionDateFormat), "-", rt.End.Format(transactionDateFormat))
				fmt.Println(strings.Repeat("=", columnWidth))
				PrintBalances(balances, showEmptyAccounts, transactionDepth, columnWidth)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(balanceCmd)

	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(1<<63 - 1)
	balanceCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	balanceCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
	balanceCmd.Flags().StringVar(&payeeFilter, "payee", "", "Filter output to payees that contain this string.")
	balanceCmd.Flags().IntVar(&columnWidth, "columns", 80, "Set a column width for output.")
	balanceCmd.Flags().BoolVar(&columnWide, "wide", false, "Wide output (use terminal width).")

	balanceCmd.Flags().StringVar(&period, "period", "", "Split output into periods (Monthly,Quarterly,SemiYearly,Yearly).")
	balanceCmd.Flags().BoolVar(&showEmptyAccounts, "empty", false, "Show empty (zero balance) accounts.")
	balanceCmd.Flags().IntVar(&transactionDepth, "depth", -1, "Depth of transaction output (balance).")
}
