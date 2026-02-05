package cmd

import (
	"log"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/howeyc/ledger"
	"github.com/shopspring/decimal"
	"github.com/spf13/cobra"
)

// equityCmd represents the equity command
var equityCmd = &cobra.Command{
	Use:   "equity [account-substring-filter]...",
	Short: "Print account equity as transaction",
	Run: func(_ *cobra.Command, args []string) {
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

		filterArr := args
		balances := make(map[string]decimal.Decimal)
		for _, trans := range generalLedger {
			for _, accChange := range trans.AccountChanges {
				inFilter := len(filterArr) == 0
				for _, filter := range filterArr {
					if strings.Contains(accChange.Name, filter) {
						inFilter = true
					}
				}
				if inFilter {
					if decNum, ok := balances[accChange.Name]; !ok {
						balances[accChange.Name] = accChange.Balance
					} else {
						balances[accChange.Name] = decNum.Add(accChange.Balance)
					}
				}
			}
		}

		eqBal := decimal.Zero
		for name, bal := range balances {
			if !bal.IsZero() {
				trans.AccountChanges = append(trans.AccountChanges, ledger.Account{
					Name:    name,
					Balance: bal,
				})
			}
			eqBal = eqBal.Add(bal)
		}
		trans.AccountChanges = append(trans.AccountChanges, ledger.Account{
			Name:    "Equity",
			Balance: eqBal.Neg(),
		})

		slices.SortFunc(trans.AccountChanges, func(a, b ledger.Account) int {
			return strings.Compare(a.Name, b.Name)
		})

		WriteTransaction(os.Stdout, &trans, 80)
	},
}

func init() {
	rootCmd.AddCommand(equityCmd)

	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(1<<63 - 1)
	equityCmd.Flags().StringVarP(&startString, "begin-date", "b", startDate.Format(transactionDateFormat), "Begin date of transaction processing.")
	equityCmd.Flags().StringVarP(&endString, "end-date", "e", endDate.Format(transactionDateFormat), "End date of transaction processing.")
}
