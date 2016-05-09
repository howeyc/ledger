package main

import (
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/howeyc/ledger"

	"github.com/go-martini/martini"
)

func getRangeAndPeriod(dateRange, dateFreq string) (start, end time.Time, period ledger.Period) {
	switch dateFreq {
	case "Monthly":
		period = ledger.PeriodMonth
	case "Quarterly":
		period = ledger.PeriodQuarter
	case "Yearly":
		period = ledger.PeriodYear
	default:
		period = ledger.PeriodMonth
	}

	currentTime := time.Now()
	switch dateRange {
	case "All Time":
		end = currentTime.Add(time.Hour * 24)
	case "YTD":
		start = time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		end = currentTime.Add(time.Hour * 24)
	case "Previous Year":
		start = time.Date(currentTime.Year()-1, time.January, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	case "Previous Month":
		start = time.Date(currentTime.Year(), currentTime.Month()-1, 1, 0, 0, 0, 0, time.UTC)
		end = time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "Current Month":
		start = time.Date(currentTime.Year(), currentTime.Month(), 1, 0, 0, 0, 0, time.UTC)
		end = currentTime.Add(time.Hour * 24)
	case "Current Quarter":
		switch currentTime.Month() {
		case time.January, time.February, time.March:
			start = time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year(), time.April, 1, 0, 0, 0, 0, time.UTC)
		case time.April, time.May, time.June:
			start = time.Date(currentTime.Year(), time.April, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
		case time.July, time.August, time.September:
			start = time.Date(currentTime.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year(), time.October, 1, 0, 0, 0, 0, time.UTC)
		case time.October, time.November, time.December:
			start = time.Date(currentTime.Year(), time.October, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year()+1, time.January, 1, 0, 0, 0, 0, time.UTC)
		}
	case "Previous Quarter":
		switch currentTime.Month() {
		case time.January, time.February, time.March:
			start = time.Date(currentTime.Year()-1, time.October, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		case time.April, time.May, time.June:
			start = time.Date(currentTime.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year(), time.April, 1, 0, 0, 0, 0, time.UTC)
		case time.July, time.August, time.September:
			start = time.Date(currentTime.Year(), time.April, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
		case time.October, time.November, time.December:
			start = time.Date(currentTime.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
			end = time.Date(currentTime.Year(), time.October, 1, 0, 0, 0, 0, time.UTC)
		}
	}

	return
}

func ReportHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	reportName := params["reportName"]

	trans, terr := getTransactions()
	if terr != nil {
		http.Error(w, terr.Error(), 500)
		return
	}

	var rConf reportConfig
	for _, reportConf := range reportConfigData.Reports {
		if reportConf.Name == reportName {
			rConf = reportConf
		}
	}
	rStart, rEnd, rPeriod := getRangeAndPeriod(rConf.DateRange, rConf.DateFreq)

	trans = ledger.TransactionsInDateRange(trans, rStart, rEnd)

	switch rConf.Chart {
	case "pie":
		accountName := rConf.Accounts[0]
		balances := ledger.GetBalances(trans, []string{accountName})

		skipCount := 0
		for _, account := range balances {
			if !strings.HasPrefix(account.Name, accountName) {
				skipCount++
			}
			if account.Name == accountName {
				skipCount++
			}
		}

		accStartLen := len(accountName)

		type pieAccount struct {
			Name      string
			Balance   float64
			Color     string
			Highlight string
		}

		var values []pieAccount

		type pieColor struct {
			Color     string
			Highlight string
		}

		colorlist := []pieColor{{"#F7464A", "#FF5A5E"},
			{"#46BFBD", "#5AD3D1"},
			{"#FDB45C", "#FFC870"},
			{"#B48EAD", "#C69CBE"},
			{"#949FB1", "#A8B3C5"},
			{"#4D5360", "#616774"},
			{"#23A1A3", "#34B3b5"},
			{"#bf9005", "#D1A216"},
			{"#1742d1", "#2954e2"},
			{"#E228BA", "#E24FC2"},
			{"#A52A2A", "#B73C3C"},
			{"#3EB73C", "#4CBA4A"},
			{"#A014CE", "#AB49CC"},
			{"#F9A200", "#F9B12A"},
			{"#075400", "#4B7C47"},
		}

		colorIdx := 0
		for _, account := range balances[skipCount:] {
			accName := account.Name[accStartLen+1:]
			value, _ := account.Balance.Float64()

			include := true
			for _, excludeName := range rConf.Exclude {
				if strings.Contains(accName, excludeName) {
					include = false
				}
			}

			if include && !strings.Contains(accName, ":") {
				values = append(values, pieAccount{Name: accName, Balance: value,
					Color:     colorlist[colorIdx].Color,
					Highlight: colorlist[colorIdx].Highlight})
				colorIdx++
			}
		}

		type piePageData struct {
			pageData
			ReportName           string
			RangeStart, RangeEnd time.Time
			ChartAccounts        []pieAccount
		}

		var rtrans []*ledger.Transaction
		for _, tran := range trans {
			for _, accChange := range tran.AccountChanges {
				include := false
				if strings.HasPrefix(accChange.Name, accountName) {
					include = true
				}
				for _, excludeName := range rConf.Exclude {
					if strings.Contains(accChange.Name, excludeName) {
						include = false
					}
				}
				if include {
					rtrans = append(rtrans, tran)
				}
			}
		}

		var pData piePageData
		pData.Reports = reportConfigData.Reports
		pData.Transactions = rtrans
		pData.ChartAccounts = values
		pData.RangeStart = rStart
		pData.RangeEnd = rEnd
		pData.ReportName = reportName

		t, err := parseAssets("templates/template.piechart.html", "templates/template.nav.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = t.Execute(w, pData)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	case "line", "bar":
		colorlist := []string{"220,220,220", "151,187,205", "70, 191, 189", "191, 71, 73", "191, 71, 133", "71, 191, 129", "165,42,42"}
		type lineData struct {
			AccountName string
			RGBColor    string
			Values      []float64
		}
		type linePageData struct {
			pageData
			ReportName           string
			RangeStart, RangeEnd time.Time
			ChartType            string
			Labels               []string
			DataSets             []lineData
		}
		var lData linePageData
		lData.Reports = reportConfigData.Reports
		lData.ReportName = reportName

		colorIdx := 0
		for _, freqAccountName := range rConf.Accounts {
			lData.DataSets = append(lData.DataSets,
				lineData{AccountName: freqAccountName,
					RGBColor: colorlist[colorIdx]})
			colorIdx++
		}

		var rType ledger.RangeType
		switch rConf.Chart {
		case "line":
			rType = ledger.RangeSnapshot
			lData.ChartType = "Line"
		case "bar":
			rType = ledger.RangePartition
			lData.ChartType = "Bar"
		}

		rangeBalances := ledger.BalancesByPeriod(trans, rPeriod, rType)
		for _, rb := range rangeBalances {
			if lData.RangeStart.IsZero() {
				lData.RangeStart = rb.Start
			}
			lData.RangeEnd = rb.End
			lData.Labels = append(lData.Labels, rb.End.Format("2006-01-02"))

			accVals := make(map[string]float64)
			for _, freqAccountName := range rConf.Accounts {
				accVals[freqAccountName] = 0
			}
			for _, freqAccountName := range rConf.Accounts {
				for _, bal := range rb.Balances {
					if bal.Name == freqAccountName {
						fval, _ := bal.Balance.Float64()
						fval = math.Abs(fval)
						accVals[freqAccountName] = fval
					}
				}
			}
			for dIdx := range lData.DataSets {
				lData.DataSets[dIdx].Values = append(lData.DataSets[dIdx].Values, accVals[lData.DataSets[dIdx].AccountName])
			}
		}

		var rtrans []*ledger.Transaction
		for _, tran := range trans {
			include := false
			for _, freqAccountName := range rConf.Accounts {
				for _, accChange := range tran.AccountChanges {
					if strings.HasPrefix(accChange.Name, freqAccountName) {
						include = true
					}
				}
			}

			if include {
				rtrans = append(rtrans, tran)
			}
		}
		lData.Transactions = rtrans

		t, err := parseAssets("templates/template.barlinechart.html", "templates/template.nav.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = t.Execute(w, lData)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

	default:
		fmt.Fprint(w, "Unsupported chart type.")
	}
}
