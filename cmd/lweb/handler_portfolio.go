package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"

	"github.com/go-martini/martini"
	"github.com/howeyc/ledger"
)

type iexQuote struct {
	Company       string  `json:"companyName"`
	Exchange      string  `json:"primaryExchange"`
	Close         float64 `json:"close"`
	PreviousClose float64 `json:"previousClose"`
	Last          float64 `json:"latestPrice"`
}

// https://iexcloud.io/docs/api/
func stockQuote(symbol string) (quote iexQuote, err error) {
	resp, herr := http.Get("https://cloud.iexapis.com/beta/stock/" + symbol + "/quote?token=" + portfolioConfigData.IEXToken)
	if herr != nil {
		return quote, herr
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	derr := dec.Decode(&quote)
	if derr != nil {
		return quote, derr
	}
	if quote.Company == "" && quote.Exchange == "" {
		return quote, errors.New("Unable to find data for symbol " + symbol)
	}
	return quote, nil
}

type gdaxQuote struct {
	Volume        string  `json:"volume"`
	PreviousClose float64 `json:"open,string"`
	Last          float64 `json:"last,string"`
}

// https://docs.gdax.com/
func cryptoQuote(symbol string) (quote gdaxQuote, err error) {
	resp, herr := http.Get("https://api.gdax.com/products/" + symbol + "/stats")
	if herr != nil {
		return quote, herr
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	derr := dec.Decode(&quote)
	if derr != nil {
		return quote, derr
	}
	if quote.Volume == "" {
		return quote, errors.New("Unable to find data for symbol " + symbol)
	}
	return quote, nil
}

func portfolioHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	portfolioName := params["portfolioName"]

	var portfolio portfolioStruct
	for _, port := range portfolioConfigData.Portfolios {
		if port.Name == portfolioName {
			portfolio = port
		}
	}

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

	type portPageData struct {
		pageData
		PortfolioName string
	}

	var pData portPageData
	pData.Reports = reportConfigData.Reports
	pData.Portfolios = portfolioConfigData.Portfolios
	pData.Transactions = trans
	pData.PortfolioName = portfolioName

	sectionTotals := make(map[string]stockInfo)
	siChan := make(chan stockInfo)

	for _, stock := range portfolio.Stocks {
		go func(name, account, symbol, securityType, section string, shares float64) {
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
			var sprice, sclose float64
			switch securityType {
			case "Stock":
				quote, qerr := stockQuote(symbol)
				if qerr == nil {
					sprice = quote.Last
					if quote.Close > 0 {
						sclose = quote.Close
					} else {
						sclose = quote.PreviousClose
					}
				}
			case "Fund":
				quote, qerr := stockQuote(symbol)
				if qerr == nil {
					sprice = quote.Last
					if quote.Close > 0 {
						sclose = quote.Close
					} else {
						sclose = quote.PreviousClose
					}
				}
			case "Crypto":
				quote, qerr := cryptoQuote(symbol)
				if qerr == nil {
					sprice = quote.Last
					sclose = quote.PreviousClose
				}
			case "Cash":
				sprice = 1
				sclose = 1
				si.Shares = si.Cost
			default:
				sprice = cprice
				sclose = cprice
			}

			if sprice == 0 {
				sprice = sclose
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
		}(stock.Name, stock.Account, stock.Ticker, stock.SecurityType, stock.Section, stock.Shares)
	}
	for range portfolio.Stocks {
		pData.Stocks = append(pData.Stocks, <-siChan)
	}

	stotal := stockInfo{Name: "Total", Section: "zzzTotal", Type: "Total"}
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
