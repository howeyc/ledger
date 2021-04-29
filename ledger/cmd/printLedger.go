package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

// ledgerCmd represents the ledger command
var ledgerCmd = &cobra.Command{
	Use:   "ledger [account-substring-filter]...",
	Short: "Print transactions in ledger file format",
	Run: func(cmd *cobra.Command, args []string) {
		generalLedger, err := cliTransactions()
		if err != nil {
			log.Fatalln(err)
		}

		PrintLedger(generalLedger, args, columnWidth)
	},
}

func init() {
	printCmd.AddCommand(ledgerCmd)
}
