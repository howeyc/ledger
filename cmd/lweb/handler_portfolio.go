package main

import (
	"net/http"

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

	for _, stock := range stockConfigData.Stocks {
		quote, _ := yquotes.GetPrice(stock.Ticker)
		var sprice float64
		var sclose float64
		var cprice float64
		if quote != nil {
			sprice = quote.Last
			sclose = quote.PreviousClose
		}
		si := stockInfo{Name: stock.Name, Ticker: stock.Ticker, Price: sprice, Shares: stock.Shares}
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
	stotal := stockInfo{Name: "Total"}
	for _, si := range pData.Stocks {
		stotal.Cost += si.Cost
		stotal.MarketValue += si.MarketValue
		stotal.GainLossOverall += si.GainLossOverall
		stotal.GainLossDay += si.GainLossDay
	}
	stotal.PriceChangePctDay = (stotal.GainLossDay / stotal.Cost) * 100.0
	stotal.PriceChangePctOverall = (stotal.GainLossOverall / stotal.Cost) * 100.0
	pData.Stocks = append(pData.Stocks, stotal)

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
