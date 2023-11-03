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

		children := make(map[string]int)
		for _, acc := range balances {
			if i := strings.LastIndex(acc.Name, ":"); i >= 0 {
				children[acc.Name[:i]]++
			}
		}

		for _, acc := range balances {
			match := true
			if accountLeavesOnly && children[acc.Name] > 0 {
				match = false
			}
			if accountMatchDepth && filterDepth != strings.Count(acc.Name, ":") {
				match = false
			}
			if match {
				fmt.Println(acc.Name)
			}
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
