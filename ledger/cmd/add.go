package cmd

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/erikgeiser/promptkit/confirmation"
	"github.com/erikgeiser/promptkit/selection"
	"github.com/erikgeiser/promptkit/textinput"
	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

type posting struct {
	Name   string
	Amount float64
}

var addDryRun bool

// addCmd represents the stats command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Interactive way to add a transaction to ledger",
	Run: func(cmd *cobra.Command, args []string) {
		transactions, terr := cliTransactions()
		if terr != nil {
			log.Fatalln(terr)
		}

		input := textinput.New("Date:")
		input.InitialValue = time.Now().Format("2006/01/02")
		input.Validate = func(v string) error {
			_, err := time.ParseInLocation("2006/01/02", v, time.Local)
			return err
		}

		date, err := input.RunPrompt()
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}

		payeeSet := make(map[string]struct{})
		accSet := make(map[string]struct{})
		for _, trans := range transactions {
			payeeSet[trans.Payee] = struct{}{}
			for _, p := range trans.AccountChanges {
				accSet[p.Name] = struct{}{}
			}
		}
		var payeeList []string
		for p := range payeeSet {
			payeeList = append(payeeList, p)
		}
		sort.Strings(payeeList)
		var accountList []string
		for a := range accSet {
			accountList = append(accountList, a)
		}
		sort.Strings(accountList)

		sp := selection.New("Payee:", selection.Choices(payeeList))
		sp.PageSize = 10
		sp.Filter = selection.FilterContainsCaseSensitive
		payee, err := sp.RunPrompt()
		if err != nil {
			log.Fatalf("Error: %v\n", err)
		}

		var postings []posting

		for {
			p := posting{}

			sa := selection.New("Account:", selection.Choices(accountList))
			sa.PageSize = 10
			sa.Filter = selection.FilterContainsCaseSensitive
			accSel, serr := sa.RunPrompt()
			if serr != nil {
				log.Fatalf("Error: %v\n", serr)
			}
			p.Name = accSel.String

			amtin := textinput.New("Amount:")
			amtin.Validate = func(v string) error {
				_, err := strconv.ParseFloat(v, 64)
				return err
			}
			amtStr, serr := amtin.RunPrompt()
			if serr != nil {
				log.Fatalf("Error: %v\n", serr)
			}
			p.Amount, _ = strconv.ParseFloat(amtStr, 64)

			postings = append(postings, p)
			if len(postings) < 2 {
				continue
			}

			another := confirmation.New("Another Posting?", confirmation.Undecided)
			ready, err := another.RunPrompt()
			if err != nil {
				log.Fatalf("Error: %v\n", err)
			}
			if !ready {
				break
			}
		}

		var tbuf bytes.Buffer
		fmt.Fprintln(&tbuf, date, payee.String)
		for _, p := range postings {
			fmt.Fprintf(&tbuf, "    %s  %f\n", p.Name, p.Amount)
		}
		fmt.Fprintln(&tbuf, "")

		/* Check valid transaction is created */
		trans, perr := ledger.ParseLedger(&tbuf)
		if perr != nil {
			log.Fatalf("Error: %v\n", perr)
		}

		var mw io.Writer
		if addDryRun {
			mw = os.Stdout
		} else {
			f, _ := os.OpenFile(ledgerFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			defer f.Close()
			mw = io.MultiWriter(f, os.Stdout)
		}

		if columnWide {
			columnWidth = 132
		}

		for _, t := range trans {
			WriteTransaction(mw, t, columnWidth)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().BoolVarP(&addDryRun, "dry-run", "n", false, "Do not add to ledger file. Display only.")
	addCmd.Flags().IntVar(&columnWidth, "columns", 80, "Set a column width for output.")
	addCmd.Flags().BoolVar(&columnWide, "wide", false, "Wide output (same as --columns=132).")
}
