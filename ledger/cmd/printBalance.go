package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

// balanceCmd represents the balance command
var balanceCmd = &cobra.Command{
	Aliases: []string{"bal"},
	Use:     "balance [account-substring-filter]...",
	Short:   "Print account balances",
	Run: func(cmd *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}
		if period == "" {
			PrintBalances(ledger.GetBalances(generalLedger, args), showEmptyAccounts, transactionDepth, columnWidth)
		} else {
			lperiod := ledger.Period(period)
			rbalances := ledger.BalancesByPeriod(generalLedger, lperiod, ledger.RangePartition)
			for rIdx, rb := range rbalances {
				if rIdx > 0 {
					fmt.Println("")
					fmt.Println(strings.Repeat("=", columnWidth))
				}
				fmt.Println(rb.Start.Format(transactionDateFormat), "-", rb.End.Format(transactionDateFormat))
				fmt.Println(strings.Repeat("=", columnWidth))
				PrintBalances(rb.Balances, showEmptyAccounts, transactionDepth, columnWidth)
			}
		}
	},
}

func init() {
	printCmd.AddCommand(balanceCmd)

	balanceCmd.Flags().StringVar(&period, "period", "", "Split output into periods (Monthly,Quarterly,SemiYearly,Yearly).")
	balanceCmd.Flags().BoolVar(&showEmptyAccounts, "empty", false, "Show empty (zero balance) accounts.")
	balanceCmd.Flags().IntVar(&transactionDepth, "depth", -1, "Depth of transaction output (balance).")
}
