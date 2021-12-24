package cmd

import (
	"net/http"
	"sort"

	"github.com/howeyc/ledger"
	"github.com/julienschmidt/httprouter"
)

func portfolioHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	portfolioName := params.ByName("portfolioName")

	var portfolio portfolioStruct
	for _, port := range portfolioConfigData.Portfolios {
		if port.Name == portfolioName {
			portfolio = port
		}
	}

	t, err := loadTemplates("templates/template.portfolio.html")
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
		ShowDividends bool
		ShowWeight    bool
	}

	var pData portPageData
	pData.Reports = reportConfigData.Reports
	pData.Portfolios = portfolioConfigData.Portfolios
	pData.Transactions = trans
	pData.PortfolioName = portfolioName
	pData.ShowDividends = portfolio.ShowDividends
	pData.ShowWeight = portfolio.ShowWeight

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
				if portfolio.ShowDividends {
					div, _ := stockAnnualDividends(symbol)
					si.AnnualDividends = div * shares
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
				if portfolio.ShowDividends {
					div, _ := stockAnnualDividends(symbol)
					si.AnnualDividends = div * shares
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
			si.AnnualYield = (si.AnnualDividends / si.MarketValue) * 100
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
		sectionInfo.AnnualDividends += si.AnnualDividends
		sectionInfo.AnnualYield = (sectionInfo.AnnualDividends / sectionInfo.MarketValue) * 100
		sectionTotals[si.Section] = sectionInfo

		stotal.Cost += si.Cost
		stotal.MarketValue += si.MarketValue
		stotal.GainLossOverall += si.GainLossOverall
		stotal.GainLossDay += si.GainLossDay
		stotal.AnnualDividends += si.AnnualDividends
	}
	stotal.PriceChangePctDay = (stotal.GainLossDay / stotal.Cost) * 100.0
	stotal.PriceChangePctOverall = (stotal.GainLossOverall / stotal.Cost) * 100.0
	stotal.AnnualYield = (stotal.AnnualDividends / stotal.MarketValue) * 100
	pData.Stocks = append(pData.Stocks, stotal)

	for _, sectionInfo := range sectionTotals {
		sectionInfo.PriceChangePctDay = (sectionInfo.GainLossDay / sectionInfo.Cost) * 100.0
		sectionInfo.PriceChangePctOverall = (sectionInfo.GainLossOverall / sectionInfo.Cost) * 100.0

		for i, si := range pData.Stocks {
			if si.Section == sectionInfo.Name {
				pData.Stocks[i].Weight = (si.MarketValue / sectionInfo.MarketValue) * 100
			}
		}
		sectionInfo.Weight = (sectionInfo.MarketValue / stotal.MarketValue) * 100

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
