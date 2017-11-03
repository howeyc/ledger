package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/howeyc/ledger"
)

type iexQuote struct {
	Company       string  `json:"companyName"`
	Exchange      string  `json:"primaryExchange"`
	PreviousClose float64 `json:"close"`
	Last          float64 `json:"latestPrice"`
}

// https://iextrading.com/developer/docs
func stockQuote(symbol string) (quote iexQuote, err error) {
	resp, herr := http.Get("https://api.iextrading.com/1.0/stock/" + symbol + "/quote")
	if herr != nil {
		return quote, herr
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&quote)
	if quote.Company == "" && quote.Exchange == "" {
		return quote, errors.New("Unable to find data for symbol " + symbol)
	}
	return quote, nil
}

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
	siChan := make(chan stockInfo)

	for _, stock := range stockConfigData.Stocks {
		go func(name, account, symbol, section string, shares float64) {
			si := stockInfo{Name: name,
				Section: section,
				Ticker:  symbol,
				Shares:  shares}
			for _, bal := range balances {
				if account == bal.Name {
					si.Cost, _ = bal.Balance.Float64()
				}
			}
			cprice := si.Cost / si.Shares
			sprice := cprice
			sclose := cprice

			quote, qerr := stockQuote(symbol)
			if qerr == nil {
				sprice = quote.Last
				sclose = quote.PreviousClose
			}

			si.Price = sprice
			si.MarketValue = si.Shares * si.Price
			si.GainLossOverall = si.MarketValue - si.Cost
			si.PriceChangeDay = sprice - sclose
			si.PriceChangePctDay = (si.PriceChangeDay / sclose) * 100.0
			si.PriceChangeOverall = sprice - cprice
			si.PriceChangePctOverall = (si.PriceChangeOverall / cprice) * 100.0
			si.GainLossDay = si.Shares * si.PriceChangeDay
			siChan <- si
		}(stock.Name, stock.Account, stock.Ticker, stock.Section, stock.Shares)
	}
	for range stockConfigData.Stocks {
		pData.Stocks = append(pData.Stocks, <-siChan)
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
