package main

import (
	"net/http"
	"sort"

	"github.com/doneland/yquotes"
	"github.com/howeyc/ledger"
)

func portfolioHandler(w http.ResponseWriter, r *http.Request) {
	t, err := parseAssets("templates/template.portfolio.html", "templates/template.nav.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	trans, terr := getTransactions()
	if terr != nil {
		http.Error(w, terr.Error(), 500)
		return
	}
	balances := ledger.GetBalances(trans, []string{})

	var pData pageData
	pData.Reports = reportConfigData.Reports
	pData.Transactions = trans

	sectionTotals := make(map[string]stockInfo)

	for _, stock := range stockConfigData.Stocks {
		quote, _ := yquotes.GetPrice(stock.Ticker)
		var sprice float64
		var sclose float64
		var cprice float64
		if quote != nil {
			sprice = quote.Last
			sclose = quote.PreviousClose
		}
		si := stockInfo{Name: stock.Name,
			Section: stock.Section,
			Ticker:  stock.Ticker,
			Price:   sprice,
			Shares:  stock.Shares}
		for _, bal := range balances {
			if stock.Account == bal.Name {
				si.Cost, _ = bal.Balance.Float64()
			}
		}
		cprice = si.Cost / si.Shares
		si.MarketValue = si.Shares * si.Price
		si.GainLossOverall = si.MarketValue - si.Cost
		si.PriceChangeDay = sprice - sclose
		si.PriceChangePctDay = (si.PriceChangeDay / sclose) * 100.0
		si.PriceChangeOverall = sprice - cprice
		si.PriceChangePctOverall = (si.PriceChangeOverall / cprice) * 100.0
		si.GainLossDay = si.Shares * si.PriceChangeDay
		pData.Stocks = append(pData.Stocks, si)
	}
	stotal := stockInfo{Name: "Total", Section: "Total", Type: "Total"}
	for _, si := range pData.Stocks {
		sectionInfo := sectionTotals[si.Section]
		sectionInfo.Name = si.Section
		sectionInfo.Section = si.Section
		sectionInfo.Type = "Section Total"
		sectionInfo.Ticker = "zzz"
		sectionInfo.Cost += si.Cost
		sectionInfo.MarketValue += si.MarketValue
		sectionInfo.GainLossOverall += si.GainLossOverall
		sectionInfo.GainLossDay += si.GainLossDay
		sectionTotals[si.Section] = sectionInfo

		stotal.Cost += si.Cost
		stotal.MarketValue += si.MarketValue
		stotal.GainLossOverall += si.GainLossOverall
		stotal.GainLossDay += si.GainLossDay
	}
	stotal.PriceChangePctDay = (stotal.GainLossDay / stotal.Cost) * 100.0
	stotal.PriceChangePctOverall = (stotal.GainLossOverall / stotal.Cost) * 100.0
	pData.Stocks = append(pData.Stocks, stotal)

	for _, sectionInfo := range sectionTotals {
		sectionInfo.PriceChangePctDay = (sectionInfo.GainLossDay / sectionInfo.Cost) * 100.0
		sectionInfo.PriceChangePctOverall = (sectionInfo.GainLossOverall / sectionInfo.Cost) * 100.0
		pData.Stocks = append(pData.Stocks, sectionInfo)
	}

	sort.Slice(pData.Stocks, func(i, j int) bool {
		return pData.Stocks[i].Ticker < pData.Stocks[j].Ticker
	})
	sort.SliceStable(pData.Stocks, func(i, j int) bool {
		return pData.Stocks[i].Section < pData.Stocks[j].Section
	})

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
