package main

import (
	"fmt"
	"math"
	"math/big"
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

// getAccounts will return the accounts that match accountNeedle.
// If accountNeedle contains no wildcards (*), only case-sensitive matchs are returned.
// In the case of wildcards:
//	* matches any account that shares the prefix, and has the same depth as the *.
//	** matches any account that shares the prefix, at any depth.
func getAccounts(accountNeedle string, accountsHaystack []*ledger.Account) (results []*ledger.Account) {
	needleDepth := len(strings.Split(accountNeedle, ":"))

	if dblstarIdx := strings.Index(accountNeedle, "**"); dblstarIdx != -1 {
		prefixNeedle := accountNeedle[:dblstarIdx]
		for _, hay := range accountsHaystack {
			if strings.HasPrefix(hay.Name, prefixNeedle) {
				results = append(results, hay)
			}
		}
	} else if starIdx := strings.Index(accountNeedle, "*"); starIdx != -1 {
		prefixNeedle := accountNeedle[:starIdx]
		for _, hay := range accountsHaystack {
			hayDepth := len(strings.Split(hay.Name, ":"))
			if strings.HasPrefix(hay.Name, prefixNeedle) && hayDepth == needleDepth {
				results = append(results, hay)
			}
		}
	} else {
		for _, hay := range accountsHaystack {
			if hay.Name == accountNeedle {
				results = append(results, hay)
			}
		}
	}

	return
}

func calcBalances(calcAccts []calculatedAccount, balances []*ledger.Account) (results []*ledger.Account) {
	accVals := make(map[string]float64)
	for _, calcAccount := range calcAccts {
		for _, bal := range balances {
			for _, acctOp := range calcAccount.AccountOperations {
				if acctOp.Name == bal.Name {
					fval, _ := bal.Balance.Float64()
					fval = math.Abs(fval)
					aval := accVals[calcAccount.Name]
					switch acctOp.Operation {
					case "+":
						aval += fval
					case "-":
						aval -= fval
					}
					accVals[calcAccount.Name] = aval
				}
			}
		}
	}

	for _, calcAccount := range calcAccts {
		bal := big.NewRat(1, 1)
		bal.SetFloat64(accVals[calcAccount.Name])
		results = append(results, &ledger.Account{Name: calcAccount.Name, Balance: bal})
	}

	return
}

func reportHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
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

	var rtrans []*ledger.Transaction
	for _, tran := range trans {
		include := true
		for _, accChange := range tran.AccountChanges {
			for _, excludeName := range rConf.ExcludeAccountTrans {
				if strings.Contains(accChange.Name, excludeName) {
					include = false
				}
			}
		}

		if include {
			rtrans = append(rtrans, tran)
		}
	}

	balances := ledger.GetBalances(rtrans, []string{})
	var initialAccounts []*ledger.Account
	for _, confAccount := range rConf.Accounts {
		initialAccounts = append(initialAccounts, getAccounts(confAccount, balances)...)
	}
	initialAccounts = append(initialAccounts, calcBalances(rConf.CalculatedAccounts, balances)...)
	var reportSummaryAccounts []*ledger.Account
	for _, account := range initialAccounts {
		include := true
		for _, excludeName := range rConf.ExcludeAccountsSummary {
			if strings.Contains(account.Name, excludeName) {
				include = false
			}
		}

		if include {
			reportSummaryAccounts = append(reportSummaryAccounts, account)
		}
	}

	switch rConf.Chart {
	case "pie":
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
		for _, account := range reportSummaryAccounts {
			accName := account.Name
			value, _ := account.Balance.Float64()
			values = append(values, pieAccount{Name: accName, Balance: value,
				Color:     colorlist[colorIdx].Color,
				Highlight: colorlist[colorIdx].Highlight})
			colorIdx++
		}

		type piePageData struct {
			pageData
			ReportName           string
			RangeStart, RangeEnd time.Time
			ChartAccounts        []pieAccount
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
	case "line", "bar", "stackedbar":
		colorlist := []string{"220 ,220, 220", // grey
			"128, 128, 0", // olive-green
			"198, 113, 113", // salmon
			"0, 255, 255", // cyan
			"113, 113, 198", // light purple
			"255, 255, 0", // yellow
			"0, 0, 255", // blue
			"0, 255, 0", // green
			"255, 128, 0", // orange
			"128, 0, 128", // purple 
			"56, 142, 142", // teal
			"139, 69, 0", // brown
			"30, 30, 30", // black
			"128, 0, 0", // darkred
		}
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
		for _, repAccount := range reportSummaryAccounts {
			lData.DataSets = append(lData.DataSets,
				lineData{AccountName: repAccount.Name,
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
		case "stackedbar":
			rType = ledger.RangePartition
			lData.ChartType = "StackedBar"
		}

		lData.Transactions = rtrans

		rangeBalances := ledger.BalancesByPeriod(rtrans, rPeriod, rType)
		for _, rb := range rangeBalances {
			if lData.RangeStart.IsZero() {
				lData.RangeStart = rb.Start
			}
			lData.RangeEnd = rb.End
			lData.Labels = append(lData.Labels, rb.End.Format("2006-01-02"))

			accVals := make(map[string]float64)
			for _, confAccount := range rConf.Accounts {
				for _, freqAccount := range getAccounts(confAccount, rb.Balances) {
					fval, _ := freqAccount.Balance.Float64()
					fval = math.Abs(fval)
					accVals[freqAccount.Name] = fval
				}
			}

			for _, calcAccount := range calcBalances(rConf.CalculatedAccounts, rb.Balances) {
				fval, _ := calcAccount.Balance.Float64()
				fval = math.Abs(fval)
				accVals[calcAccount.Name] = fval
			}

			for dIdx := range lData.DataSets {
				lData.DataSets[dIdx].Values = append(lData.DataSets[dIdx].Values, accVals[lData.DataSets[dIdx].AccountName])
			}
		}

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
