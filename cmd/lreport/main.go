package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

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
	var startDate, endDate time.Time
	startDate = time.Date(1970, 1, 1, 0, 0, 0, 0, time.Local)
	endDate = time.Now().Add(time.Hour * 24)
	var startString, endString string
	var periodQuarterly, periodMonthly, periodWeekly bool

	var ledgerFileName string

	ledger.TransactionDateFormat = "2006/01/02"
	TransactionDateFormat := ledger.TransactionDateFormat

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.StringVar(&startString, "s", startDate.Format(TransactionDateFormat), "Start date of transaction processing.")
	flag.StringVar(&endString, "e", endDate.Format(TransactionDateFormat), "End date of transaction processing.")
	flag.BoolVar(&periodQuarterly, "quarterly", false, "Plot quarterly values.")
	flag.BoolVar(&periodMonthly, "monthly", false, "Plot monthly values.")
	flag.BoolVar(&periodWeekly, "weekly", false, "Plot weekly values.")
	flag.Parse()

	parsedStartDate, tstartErr := time.Parse(TransactionDateFormat, startString)
	parsedEndDate, tendErr := time.Parse(TransactionDateFormat, endString)

	if tstartErr != nil || tendErr != nil {
		fmt.Println("Unable to parse start or end date string argument.")
		fmt.Println("Expected format: YYYY/MM/dd")
		return
	}

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

	timeStartIndex, timeEndIndex := 0, 0
	for idx := 0; idx < len(generalLedger); idx++ {
		if generalLedger[idx].Date.After(parsedStartDate) {
			timeStartIndex = idx
			break
		}
	}
	for idx := len(generalLedger) - 1; idx >= 0; idx-- {
		if generalLedger[idx].Date.Before(parsedEndDate) {
			timeEndIndex = idx
			break
		}
	}
	generalLedger = generalLedger[timeStartIndex : timeEndIndex+1]

	if len(generalLedger) < 1 {
		fmt.Println("No Transactions.")
		os.Exit(0)
	}

	lines := make(map[string]plotter.XYs)
	currentDate := generalLedger[0].Date
	sYear := currentDate.Year()
	prevQuarter := Quarter(0)
	prevMonth := time.Month(0)
	prevWeek := 0
	for idx, trans := range generalLedger {
		if trans.Date.After(currentDate) {
			currentDate = trans.Date
			if _, currWeek := currentDate.ISOWeek(); (periodQuarterly && prevQuarter != getQuarter(currentDate)) ||
				(periodMonthly && prevMonth != currentDate.Month()) ||
				(periodWeekly && prevWeek != currWeek) {
				balances := ledger.GetBalances(generalLedger[:idx], flag.Args())
				XCoord := float64((currentDate.Year()-sYear)*365 + currentDate.YearDay())
				netWorth := new(big.Rat)
				for _, balance := range balances {
					switch balance.Name {
					case "Assets", "Liabilities":
						netWorth = netWorth.Add(netWorth, balance.Balance)
					}
					YCoord, _ := balance.Balance.Float64()
					switch balance.Name {
					case "Liabilities", "Income":
						YCoord *= -1
						fallthrough
					case "Assets":
						line, lineFound := lines[balance.Name]
						if !lineFound {
							line = plotter.XYs{{XCoord, YCoord}}
						} else {
							line = append(line, plotter.XYs{{XCoord, YCoord}}...)
						}
						lines[balance.Name] = line
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

			prevQuarter = getQuarter(currentDate)
			prevMonth = currentDate.Month()
			_, prevWeek = currentDate.ISOWeek()
		}
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Accounts"
	p.X.Label.Text = "Day"
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
}
