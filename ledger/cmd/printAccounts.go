package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

var accountLeavesOnly bool
var accountMatchDepth bool

// accountsCmd represents the accounts command
var accountsCmd = &cobra.Command{
	Use:   "accounts [account-substring-filter]...",
	Short: "Print accounts list",
	Run: func(cmd *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}

		if accountMatchDepth && len(args) != 1 {
			log.Fatalln("account depth matches with one filter")
		}

		var filterDepth int
		if accountMatchDepth {
			filterDepth = strings.Count(args[0], ":")
		}

		balances := ledger.GetBalances(generalLedger, args)
		var currentAccount string
		if len(balances) > 0 {
			currentAccount = balances[0].Name
		}
		for _, account := range balances[1:] {
			if accountLeavesOnly && !strings.HasPrefix(account.Name, currentAccount) {
				fmt.Println(currentAccount)
			} else if accountMatchDepth && filterDepth == strings.Count(currentAccount, ":") {
				fmt.Println(currentAccount)
			} else if !accountLeavesOnly && !accountMatchDepth {
				fmt.Println(currentAccount)
			}
			currentAccount = account.Name
		}
		if accountMatchDepth && filterDepth == strings.Count(currentAccount, ":") {
			fmt.Println(currentAccount)
		} else if !accountMatchDepth {
			fmt.Println(currentAccount)
		}
	},
}

func init() {
	rootCmd.AddCommand(accountsCmd)

	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(1<<63 - 1)
	accountsCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	accountsCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
	accountsCmd.Flags().BoolVarP(&accountLeavesOnly, "leaves-only", "l", false, "Only show most-depth accounts")
	accountsCmd.Flags().BoolVarP(&accountMatchDepth, "match-depth", "m", false, "Show accounts with same depth as filter")
}
