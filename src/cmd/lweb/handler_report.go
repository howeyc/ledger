package main

import (
	"bytes"
	"fmt"
	"html/template"
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

	type reportAccount struct {
		Name      string
		Balance   float64
		Color     string
		Highlight string
	}

	values := make([]reportAccount, 0)

	type reportColor struct {
		Color     string
		Highlight string
	}

	colorlist := []reportColor{{"#F7464A", "#FF5A5E"},
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
			values = append(values, reportAccount{Name: accName, Balance: value,
				Color:     colorlist[colorIdx].Color,
				Highlight: colorlist[colorIdx].Highlight})
			colorIdx++
		}
	}

	switch rConf.Chart {
	case "pie":
		type piePageData struct {
			pageData
			ChartAccounts []reportAccount
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
	default:
		fmt.Fprint(w, "Unsupported chart type.")
	}
}
