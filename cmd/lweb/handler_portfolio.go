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
		if quote != nil {
			sprice = quote.Last
		}
		si := stockInfo{Name: stock.Name, Ticker: stock.Ticker, Price: sprice, Shares: stock.Shares}
		for _, bal := range balances {
			if stock.Account == bal.Name {
				si.Cost, _ = bal.Balance.Float64()
			}
		}
		si.MarketValue = si.Shares * si.Price
		si.GainLoss = si.MarketValue - si.Cost
		pData.Stocks = append(pData.Stocks, si)
	}

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
