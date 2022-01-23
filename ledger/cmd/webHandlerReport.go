package cmd

import (
	"fmt"
	"math/big"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/howeyc/ledger"
	"github.com/howeyc/ledger/ledger/cmd/internal/pdr"
	"github.com/julienschmidt/httprouter"
	colorful "github.com/lucasb-eyer/go-colorful"
)

func getRangeAndPeriod(dateRange, dateFreq string) (start, end time.Time, period ledger.Period, err error) {
	period = ledger.Period(dateFreq)

	start, end, err = pdr.ParseRange(dateRange, time.Now())

	return
}

// getAccounts will return the accounts that match accountNeedle.
// If accountNeedle contains no wildcards (*), only case-sensitive matchs are returned.
// In the case of wildcards:
//	* matches any account that shares the prefix, and has the same depth as the *.
//	** matches any account that shares the prefix, at furthest depth possible,
//	ignoring parent accounts to avoid duplicates
func getAccounts(accountNeedle string, accountsHaystack []*ledger.Account) (results []*ledger.Account) {
	needleDepth := len(strings.Split(accountNeedle, ":"))

	if dblstarIdx := strings.Index(accountNeedle, "**"); dblstarIdx != -1 {
		foundAccountNames := make(map[string]*ledger.Account)
		prefixNeedle := accountNeedle[:dblstarIdx]
		for _, hay := range accountsHaystack {
			if strings.HasPrefix(hay.Name, prefixNeedle) {
				foundAccountNames[hay.Name] = hay
			}
		}
		// Remove any parents
		for k := range foundAccountNames {
			kpre := k[:strings.LastIndex(k, ":")]
			delete(foundAccountNames, kpre)
		}
		// Remaining are the results
		for _, hay := range foundAccountNames {
			results = append(results, hay)
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
	accVals := make(map[string]*big.Rat)
	for _, calcAccount := range calcAccts {
		for _, bal := range balances {
			for _, acctOp := range calcAccount.AccountOperations {
				if acctOp.Name == bal.Name {
					fval := big.NewRat(1, 1).Abs(bal.Balance)
					aval, found := accVals[calcAccount.Name]
					if !found {
						aval = big.NewRat(0, 1)
					}
					if acctOp.MultiplicationFactor != 0 {
						factor := big.NewRat(1, 1).SetFloat64(acctOp.MultiplicationFactor)
						fval = fval.Mul(factor, fval)
					}
					oval := big.NewRat(1, 1)
					if acctOp.SubAccount != "" {
						for _, obal := range balances {
							if acctOp.SubAccount == obal.Name {
								oval = oval.Abs(obal.Balance)
							}
						}
					}
					switch acctOp.Operation {
					case "+":
						aval.Add(aval, fval)
					case "-":
						aval.Sub(aval, fval)
					case "*":
						aval.Mul(fval, oval)
					case "/":
						aval.Quo(fval, oval)
					}
					accVals[calcAccount.Name] = aval
				}
			}
		}
		if calcAccount.UseAbs {
			if aval, found := accVals[calcAccount.Name]; !found {
				accVals[calcAccount.Name] = big.NewRat(0, 1)
			} else {
				accVals[calcAccount.Name] = aval.Abs(aval)
			}
		}
	}

	for _, calcAccount := range calcAccts {
		results = append(results, &ledger.Account{Name: calcAccount.Name, Balance: accVals[calcAccount.Name]})
	}

	return
}

// Merge multiple account changes for each distinct account
func mergeAccounts(input *ledger.Transaction) {
	balmap := make(map[string]*big.Rat)
	for _, accChange := range input.AccountChanges {
		if bal, found := balmap[accChange.Name]; found {
			bal = bal.Add(bal, accChange.Balance)
			balmap[accChange.Name] = bal
		} else {
			balmap[accChange.Name] = accChange.Balance
		}
	}
	input.AccountChanges = []ledger.Account{}
	for accName, bal := range balmap {
		input.AccountChanges = append(input.AccountChanges, ledger.Account{
			Name:    accName,
			Balance: bal,
		})
	}

	// Map is random order, order by name for consistency (helps with tests)
	sort.Slice(input.AccountChanges, func(i, j int) bool {
		return input.AccountChanges[i].Name < input.AccountChanges[j].Name
	})
}

func reportHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	reportName := params.ByName("reportName")

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
	rStart, rEnd, rPeriod, rerr := getRangeAndPeriod(rConf.DateRange, rConf.DateFreq)
	if rerr != nil {
		http.Error(w, rerr.Error(), 500)
		return
	}

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

	// Filter report to only show transactions that are for the accounts in the summary of the report
	var vtrans []*ledger.Transaction
	for _, trans := range rtrans {
		include := false
		for _, accChange := range trans.AccountChanges {
			for _, account := range reportSummaryAccounts {
				if strings.Contains(accChange.Name, account.Name) {
					include = true
				}
			}
		}
		if include {
			mergeAccounts(trans)
			vtrans = append(vtrans, trans)
		}
	}

	colorPalette := colorful.FastHappyPalette(len(reportSummaryAccounts))
	colorBlack := colorful.Color{R: 1, G: 1, B: 1}

	switch rConf.Chart {
	case "leaderboard":
		type lbAccount struct {
			Name       string
			Balance    *big.Rat
			Percentage int
		}

		var values []lbAccount
		maxValue := big.NewRat(1, 1)

		for _, account := range reportSummaryAccounts {
			accName := account.Name
			value := big.NewRat(1, 1).Set(account.Balance)
			values = append(values, lbAccount{Name: accName, Balance: value})

			if maxValue.Cmp(value) < 0 {
				maxValue = value
			}
		}

		sort.Slice(values, func(i, j int) bool { return values[i].Balance.Cmp(values[j].Balance) > 0 })

		maxIdx := 0
		for idx := range values {
			mf, _ := maxValue.Float64()
			cf, _ := values[idx].Balance.Float64()
			values[idx].Percentage = int((cf / mf) * 100.0)
			if values[idx].Percentage > 5 {
				maxIdx = idx
			}
		}
		values = values[:maxIdx]

		type lbPageData struct {
			pageData
			ReportName           string
			RangeStart, RangeEnd time.Time
			ChartType            string
			ChartAccounts        []lbAccount
			MaxValue             *big.Rat
		}

		var pData lbPageData
		pData.Reports = reportConfigData.Reports
		pData.Portfolios = portfolioConfigData.Portfolios
		pData.Transactions = vtrans
		pData.ChartType = "Leaderboard"
		pData.ChartAccounts = values
		pData.RangeStart = rStart
		pData.RangeEnd = rEnd
		pData.ReportName = reportName
		pData.MaxValue = maxValue

		pData.AccountNames = []string{"All"}
		for _, ca := range pData.ChartAccounts {
			pData.AccountNames = append(pData.AccountNames, ca.Name)
		}
		sort.Strings(pData.AccountNames[1:])

		t, err := loadTemplates("templates/template.leaderboardchart.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = t.Execute(w, pData)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}

	case "pie", "polar", "doughnut":
		type pieAccount struct {
			Name      string
			Balance   *big.Rat
			Color     string
			Highlight string
		}

		var values []pieAccount

		for colorIdx, account := range reportSummaryAccounts {
			accName := account.Name
			value := big.NewRat(1, 1).Set(account.Balance)
			values = append(values, pieAccount{Name: accName, Balance: value,
				Color:     colorPalette[colorIdx].Hex(),
				Highlight: colorPalette[colorIdx].BlendRgb(colorBlack, 0.1).Hex()})
		}

		type piePageData struct {
			pageData
			ReportName           string
			RangeStart, RangeEnd time.Time
			ChartType            string
			ChartAccounts        []pieAccount
		}

		var pData piePageData
		pData.Reports = reportConfigData.Reports
		pData.Portfolios = portfolioConfigData.Portfolios
		pData.Transactions = vtrans
		pData.ChartAccounts = values
		pData.RangeStart = rStart
		pData.RangeEnd = rEnd
		pData.ReportName = reportName

		pData.AccountNames = []string{"All"}
		for _, ca := range pData.ChartAccounts {
			pData.AccountNames = append(pData.AccountNames, ca.Name)
		}
		sort.Strings(pData.AccountNames[1:])

		switch rConf.Chart {
		case "pie":
			pData.ChartType = "Pie"
		case "polar":
			pData.ChartType = "Polar"
		case "doughnut":
			pData.ChartType = "Doughnut"
		}

		t, err := loadTemplates("templates/template.piechart.html")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		err = t.Execute(w, pData)
		if err != nil {
			http.Error(w, err.Error(), 500)
		}
	case "line", "radar", "bar", "stackedbar":
		type lineData struct {
			AccountName string
			RGBColor    string
			Values      []*big.Rat
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
		lData.Portfolios = portfolioConfigData.Portfolios
		lData.ReportName = reportName

		for colorIdx, repAccount := range reportSummaryAccounts {
			r, g, b := colorPalette[colorIdx].RGB255()
			lData.DataSets = append(lData.DataSets,
				lineData{AccountName: repAccount.Name,
					RGBColor: fmt.Sprintf("%d, %d, %d", r, g, b)})
		}

		var rType ledger.RangeType
		switch rConf.Chart {
		case "line":
			rType = ledger.RangeSnapshot
			lData.ChartType = "Line"
		case "radar":
			rType = ledger.RangePartition
			lData.ChartType = "Radar"
		case "bar":
			rType = ledger.RangePartition
			lData.ChartType = "Bar"
		case "stackedbar":
			rType = ledger.RangePartition
			lData.ChartType = "StackedBar"
		}
		if rConf.RangeBalanceType != "" {
			rType = rConf.RangeBalanceType
		}

		rangeBalances := ledger.BalancesByPeriod(rtrans, rPeriod, rType)
		for _, rb := range rangeBalances {
			if rConf.RangeBalanceSkipZero {
				allZero := true
				for _, acc := range rb.Balances {
					if acc.Balance.Sign() != 0 {
						allZero = false
						break
					}
				}
				if allZero {
					continue
				}
			}

			if lData.RangeStart.IsZero() {
				lData.RangeStart = rb.Start
			}
			lData.RangeEnd = rb.End
			lData.Labels = append(lData.Labels, rb.End.Format("2006-01-02"))

			accVals := make(map[string]*big.Rat)
			for _, confAccount := range rConf.Accounts {
				for _, freqAccount := range getAccounts(confAccount, rb.Balances) {
					accVals[freqAccount.Name] = big.NewRat(1, 1).Abs(freqAccount.Balance)
				}
			}

			for _, calcAccount := range calcBalances(rConf.CalculatedAccounts, rb.Balances) {
				accVals[calcAccount.Name] = calcAccount.Balance
			}

			for dIdx := range lData.DataSets {
				aval, afound := accVals[lData.DataSets[dIdx].AccountName]
				if !afound || aval == nil {
					aval = big.NewRat(0, 1)
				}
				lData.DataSets[dIdx].Values = append(lData.DataSets[dIdx].Values, aval)
			}
		}
		lData.AccountNames = []string{"All"}
		for _, ca := range lData.DataSets {
			lData.AccountNames = append(lData.AccountNames, ca.AccountName)
		}
		sort.Strings(lData.AccountNames[1:])

		// Radar chart flips everything. Dates are each dataset and the accounts become the labels
		if rConf.Chart == "radar" {
			dates := lData.Labels
			dateAccountMap := make(map[string]*big.Rat)
			var accNames []string
			for dsIdx := range lData.DataSets {
				for dIdx := range dates {
					dateAccountMap[dates[dIdx]+","+lData.DataSets[dsIdx].AccountName] = lData.DataSets[dsIdx].Values[dIdx]
				}
				accNames = append(accNames, lData.DataSets[dsIdx].AccountName)
			}

			lData.DataSets = []lineData{}

			radarcolorPalette := colorful.FastHappyPalette(len(dates))
			for colorIdx, date := range dates {
				r, g, b := radarcolorPalette[colorIdx].RGB255()
				var vals []*big.Rat
				for _, repAccount := range reportSummaryAccounts {
					vals = append(vals, dateAccountMap[date+","+repAccount.Name])
				}
				lData.DataSets = append(lData.DataSets,
					lineData{AccountName: date,
						RGBColor: fmt.Sprintf("%d, %d, %d", r, g, b),
						Values:   vals})
			}
			lData.Labels = accNames
			lData.AccountNames = append([]string{"All"}, accNames...)
			sort.Strings(lData.AccountNames[1:])
		}

		lData.Transactions = vtrans

		t, err := loadTemplates("templates/template.barlinechart.html")
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
