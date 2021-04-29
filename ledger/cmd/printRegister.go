package cmd

import (
	"fmt"
	"log"
	"strings"

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
	printCmd.AddCommand(registerCmd)

	registerCmd.Flags().StringVar(&period, "period", "", "Split output into periods (Monthly,Quarterly,SemiYearly,Yearly).")
}
