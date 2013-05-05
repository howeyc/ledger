package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"

	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/plotutil"
	"github.com/howeyc/ledger/pkg/ledger"
)

func usage() {
	flag.Usage()
	os.Exit(1)
}

func main() {
	var ledgerFileName string

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.Parse()

	ledgerFileReader, err := os.Open(ledgerFileName)
	if err != nil {
		fmt.Println("Ledger: ", err)
		return
	}
	defer ledgerFileReader.Close()

	generalLedger, parseError := ledger.ParseLedger(ledgerFileReader)
	if parseError != nil {
		fmt.Println(parseError)
		return
	}

	if len(generalLedger) < 1 {
		fmt.Println("No Transactions.")
		os.Exit(0)
	}

	lines := make(map[string]plotter.XYs)
	currentDate := generalLedger[0].Date
	for idx, trans := range generalLedger {
		if trans.Date.After(currentDate) {
			currentDate = trans.Date
			// Quarterly
			if currentDate.Day() == 1 && (currentDate.Month() == 1 || currentDate.Month() == 4 || currentDate.Month() == 7 || currentDate.Month() == 10) {
				balances := ledger.GetBalances(generalLedger[:idx], []string{})
				XCoord := float64(currentDate.Year()*100 + int(currentDate.Month()))
				netWorth := new(big.Rat)
				for _, balance := range balances {
					switch balance.Name {
					case "Assets":
						YCoord, _ := balance.Balance.Float64()
						line, lineFound := lines["Assets"]
						if !lineFound {
							line = plotter.XYs{{XCoord, YCoord}}
						} else {
							line = append(line, plotter.XYs{{XCoord, YCoord}}...)
						}
						lines["Assets"] = line
						netWorth = netWorth.Add(netWorth, balance.Balance)
					case "Liabilities":
						YCoord, _ := balance.Balance.Float64()
						YCoord *= -1
						line, lineFound := lines["Liabilities"]
						if !lineFound {
							line = plotter.XYs{{XCoord, YCoord}}
						} else {
							line = append(line, plotter.XYs{{XCoord, YCoord}}...)
						}
						lines["Liabilities"] = line
						netWorth = netWorth.Add(netWorth, balance.Balance)
					}
				}
				YCoord, _ := netWorth.Float64()
				line, lineFound := lines["NetWorth"]
				if !lineFound {
					line = plotter.XYs{{XCoord, YCoord}}
				} else {
					line = append(line, plotter.XYs{{XCoord, YCoord}}...)
				}
				lines["NetWorth"] = line
			}
		}
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Net Worth"
	p.X.Label.Text = "Quarter"
	p.Y.Label.Text = "Balance"

	plotArgs := make([]interface{}, 0)
	for label, line := range lines {
		plotArgs = append(plotArgs, label, line)

	}
	err = plotutil.AddLinePoints(p, plotArgs...)
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(11, 8.5, "networth.png"); err != nil {
		panic(err)
	}

	fmt.Println("Hello World!")
}
