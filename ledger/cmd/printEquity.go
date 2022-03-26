package cmd

import (
	"log"
	"sort"
	"time"

	"github.com/howeyc/ledger"
	"github.com/howeyc/ledger/internal/decimal"
	"github.com/spf13/cobra"
)

// equityCmd represents the equity command
var equityCmd = &cobra.Command{
	Use:   "equity",
	Short: "Print account equity as transaction",
	Run: func(cmd *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}

		var trans ledger.Transaction
		trans.Payee = "Opening Balances"
		trans.Date = time.Now()
		if len(generalLedger) > 0 {
			trans.Date = generalLedger[len(generalLedger)-1].Date
		}

		balances := make(map[string]decimal.Decimal)
		for _, trans := range generalLedger {
			for _, accChange := range trans.AccountChanges {
				if decNum, ok := balances[accChange.Name]; !ok {
					balances[accChange.Name] = accChange.Balance
				} else {
					balances[accChange.Name] = decNum.Add(accChange.Balance)
				}
			}
		}

		for name, bal := range balances {
			if !bal.IsZero() {
				trans.AccountChanges = append(trans.AccountChanges, ledger.Account{
					Name:    name,
					Balance: bal,
				})
			}
		}

		sort.Slice(trans.AccountChanges, func(i, j int) bool {
			return trans.AccountChanges[i].Name < trans.AccountChanges[j].Name
		})

		PrintTransaction(&trans, 80)
	},
}

func init() {
	rootCmd.AddCommand(equityCmd)

	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(time.Hour * 24)
	equityCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	equityCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
}
