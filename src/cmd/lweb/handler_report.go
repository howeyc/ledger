package main

import (
	"bytes"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strings"
	"time"

	"ledger"

	"github.com/go-martini/martini"
	"github.com/studiofrenetic/period"
)

func getReportPeriod(currentTime time.Time, duration string) period.Period {
	var per period.Period
	switch duration {
	case "YTD":
		per, _ = period.CreateFromYear(currentTime.Year())
	case "Previous Year":
		per, _ = period.CreateFromYear(currentTime.Year() - 1)
	case "Current Month":
		per, _ = period.CreateFromMonth(currentTime.Year(), int(currentTime.Month()))
	case "Previous Month":
		per, _ = period.CreateFromMonth(currentTime.Year(), int(currentTime.Month())-1)
	}

	return per
}

func ReportHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	reportName := params["reportName"]

	ledgerFileReader := bytes.NewReader(ledgerBuffer.Bytes())

	trans, terr := ledger.ParseLedger(ledgerFileReader)
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
	reportPeriod := getReportPeriod(time.Now(), rConf.DateRange)

	timeStartIndex, timeEndIndex := 0, 0
	for idx := 0; idx < len(trans); idx++ {
		if reportPeriod.Contains(trans[idx].Date) {
			timeStartIndex = idx
			break
		}
	}
	for idx := len(trans) - 1; idx >= 0; idx-- {
		if !reportPeriod.Contains(trans[idx].Date) {
			timeEndIndex = idx
			break
		}
	}
	trans = trans[timeStartIndex : timeEndIndex+1]

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

		values := make([]pieAccount, 0)

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
			{"#E228BA", "#E24FC2"}}

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
			ChartAccounts []pieAccount
		}

		var pData piePageData
		pData.Reports = reportConfigData.Reports
		pData.Transactions = trans
		pData.ChartAccounts = values

		t, err := template.ParseFiles("templates/template.piechart.html", "templates/template.nav.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = t.Execute(w, pData)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	case "line", "bar":
		colorlist := []string{"220,220,220", "151,187,205", "70, 191, 189", "191, 71, 73", "191, 71, 133", "71, 191, 129"}
		type lineData struct {
			AccountName string
			RGBColor    string
			Values      []float64
		}
		type linePageData struct {
			pageData
			RangeStart, RangeEnd time.Time
			ChartType            string
			Labels               []string
			DataSets             []lineData
		}
		var lData linePageData
		lData.Reports = reportConfigData.Reports
		lData.Transactions = trans

		colorIdx := 0
		for _, freqAccountName := range rConf.Accounts {
			lData.DataSets = append(lData.DataSets,
				lineData{AccountName: freqAccountName,
					RGBColor: colorlist[colorIdx]})
			colorIdx++
		}

		ledgerFileReader = bytes.NewReader(ledgerBuffer.Bytes())
		trans, _ = ledger.ParseLedger(ledgerFileReader)

		var rType ledger.RangeType
		switch rConf.Chart {
		case "line":
			rType = ledger.RangeSnapshot
			lData.ChartType = "Line"
		case "bar":
			rType = ledger.RangePartition
			lData.ChartType = "Bar"
		}

		rangeBalances := ledger.BalancesByPeriod(trans, ledger.PeriodQuarter, rType)
		for _, rb := range rangeBalances {
			if lData.RangeStart.IsZero() {
				lData.RangeStart = rb.Start
			}
			lData.RangeEnd = rb.End
			lData.Labels = append(lData.Labels, rb.End.Format("2006-01-02"))

			for _, freqAccountName := range rConf.Accounts {
				for _, bal := range rb.Balances {
					if bal.Name == freqAccountName {
						for dIdx, _ := range lData.DataSets {
							fval, _ := bal.Balance.Float64()
							fval = math.Abs(fval)
							if lData.DataSets[dIdx].AccountName == bal.Name {
								lData.DataSets[dIdx].Values = append(lData.DataSets[dIdx].Values, fval)
							}
						}
					}
				}
			}
		}

		t, err := template.ParseFiles("templates/template.barlinechart.html", "templates/template.nav.html")
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
