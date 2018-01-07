package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/go-martini/martini"
	"github.com/howeyc/ledger"
)

type wsjQuote struct {
	PreviousClose float64 `json:"close"`
	Last          float64 `json:"latestPrice"`
}

// Mutual fund quote from WSJ closing prices
func fundQuote(symbol string) (quote wsjQuote, err error) {
	letters := strings.Split(symbol, "")
	firstLetter := letters[0]
	resp, herr := http.Get("http://www.wsj.com/mdc/public/page/2_3048-usmfunds_" + firstLetter + "-usmfunds.html")
	if herr != nil {
		return quote, herr
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// WSJ is row of "name-link, symbol-link, nav, change, ytd, 3-yr" in td cells
		// find symbol-link and get next two cells
		if strings.Contains(line, "?sym="+symbol+"\">"+symbol+"<") {
			scanner.Scan() // end td
			scanner.Scan() // priceline
			tdline := scanner.Text()
			quote.Last, _ = strconv.ParseFloat(tdline[strings.Index(tdline, "\">")+2:strings.Index(tdline, "</td>")], 64)
			scanner.Scan() // change
			tdline = scanner.Text()
			changeAmount, _ := strconv.ParseFloat(tdline[strings.Index(tdline, "\">")+2:strings.Index(tdline, "</td>")], 64)
			quote.PreviousClose = quote.Last - changeAmount
			return quote, nil
		}
	}
	return quote, nil
}

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
	dec.Decode(&quote)
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
			sprice := cprice
			sclose := cprice

			switch securityType {
			case "Stock":
				quote, qerr := stockQuote(symbol)
				if qerr == nil {
					sprice = quote.Last
					sclose = quote.PreviousClose
				}
			case "Fund":
				quote, qerr := fundQuote(symbol)
				if qerr == nil {
					sprice = quote.Last
					sclose = quote.PreviousClose
				}
			case "Crypto":
				quote, qerr := cryptoQuote(symbol)
				if qerr == nil {
					sprice = quote.Last
					sclose = quote.PreviousClose
				}
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
