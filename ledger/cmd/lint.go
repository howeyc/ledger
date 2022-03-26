package cmd

import (
	"fmt"

	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

// lintCmd represents the lint command
var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Check ledger for errors",
	Run: func(cmd *cobra.Command, args []string) {
		_, lerr := ledger.ParseLedgerFile(ledgerFilePath)
		if lerr != nil {
			fmt.Println("Ledger: ", lerr)
		}
	},
}

func init() {
	rootCmd.AddCommand(lintCmd)
}
