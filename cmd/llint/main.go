package main

import (
	"fmt"
	"os"

	"github.com/howeyc/ledger"
)

func usage(name string) {
	fmt.Printf("Usage: %s <ledger-file>\n", name)
	os.Exit(1)
}

func main() {
	if len(os.Args) != 2 {
		usage(os.Args[0])
	}
	ledgerFileName := os.Args[1]
	ledgerFileReader, err := ledger.NewLedgerReader(ledgerFileName)
	if err != nil {
		fmt.Println("Ledger: ", err)
		return
	}

	c, e := ledger.ParseLedgerAsync(ledgerFileReader)
	errorCount := 0
	for {
		select {
		case <-c:
			continue
		case err := <-e:
			if err == nil {
				os.Exit(errorCount)
			}
			fmt.Println("Ledger: ", err)
			errorCount++
		}
	}
}
