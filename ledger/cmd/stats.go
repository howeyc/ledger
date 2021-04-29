package cmd

import (
	"fmt"
	"log"
	"math"
	"time"

	"github.com/hako/durafmt"
	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

// statsCmd represents the stats command
var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "A small report of transaction stats",
	Run: func(cmd *cobra.Command, args []string) {
		lreader, err := ledger.NewLedgerReader(ledgerFilePath)
		if err != nil {
			log.Fatalln(err)
		}
		transactions, terr := ledger.ParseLedger(lreader)
		if terr != nil {
			log.Fatalln(terr)
		}
		PrintStats(transactions)
	},
}

func PrintStats(generalLedger []*ledger.Transaction) {
	if len(generalLedger) < 1 {
		fmt.Println("Empty ledger.")
		return
	}
	startDate := generalLedger[0].Date
	endDate := generalLedger[len(generalLedger)-1].Date

	payees := make(map[string]struct{})
	accounts := make(map[string]struct{})

	for _, trans := range generalLedger {
		payees[trans.Payee] = struct{}{}
		for _, account := range trans.AccountChanges {
			accounts[account.Name] = struct{}{}
		}
	}

	days := math.Floor(endDate.Sub(startDate).Hours() / 24)

	fmt.Printf("%-25s : %s to %s (%s)\n", "Transactions span", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), durafmt.Parse(endDate.Sub(startDate)).String())
	fmt.Printf("%-25s : %s\n", "Since last post", durafmt.ParseShort(time.Since(endDate)).String())
	fmt.Printf("%-25s : %d (%.1f per day)\n", "Transactions", len(generalLedger), float64(len(generalLedger))/days)
	fmt.Printf("%-25s : %d\n", "Payees", len(payees))
	fmt.Printf("%-25s : %d\n", "Referenced Accounts", len(accounts))
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
