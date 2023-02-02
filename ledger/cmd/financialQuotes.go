package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
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

type avQuote struct {
	Symbol        string
	Open          float64
	PreviousClose float64
	Last          float64
}

// https://www.alphavantage.co/documentation/#latestprice
func fundQuote(symbol string) (quote avQuote, err error) {
	resp, herr := http.Get("https://alphavantage.co/query?function=GLOBAL_QUOTE&symbol=" + symbol + "&datatype=csv&apikey=" + portfolioConfigData.AVToken)
	if herr != nil {
		return quote, herr
	}
	defer resp.Body.Close()
	cr := csv.NewReader(resp.Body)

	recs, cerr := cr.ReadAll()
	if cerr != nil {
		return quote, cerr
	}
	// symbol,open,high,low,price,volume,latestDay,previousClose,change,changePercent
	if len(recs) != 2 || len(recs[0]) != 10 {
		return quote, errors.New("bad csv")
	}

	quote.Symbol = recs[1][0]
	quote.Open, _ = strconv.ParseFloat(recs[1][1], 64)
	quote.Last, _ = strconv.ParseFloat(recs[1][4], 64)
	quote.PreviousClose, _ = strconv.ParseFloat(recs[1][7], 64)

	return quote, nil
}

type gdaxQuote struct {
	Volume        string  `json:"volume"`
	PreviousClose float64 `json:"open,string"`
	Last          float64 `json:"last,string"`
}

// https://docs.pro.coinbase.com/
func cryptoQuote(symbol string) (quote gdaxQuote, err error) {
	resp, herr := http.Get("https://api.pro.coinbase.com/products/" + symbol + "/stats")
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

type iexDividend struct {
	Amount       float64 `json:"amount"`
	Currency     string  `json:"currency"`
	DeclaredDate string  `json:"declaredDate"`
	Description  string  `json:"description"`
	ExDate       string  `json:"exDate"`
	Flag         string  `json:"flag"`
	Frequency    string  `json:"frequency"`
	PaymentDate  string  `json:"paymentDate"`
	RecordDate   string  `json:"recordDate"`
	Refid        int     `json:"refid"`
	Symbol       string  `json:"symbol"`
	ID           string  `json:"id"`
	Key          string  `json:"key"`
	Subkey       string  `json:"subkey"`
	Date         int64   `json:"date"`
	Updated      float64 `json:"updated"`
}

// https://iexcloud.io/docs/api/
func stockAnnualDividends(symbol string) (amount float64, err error) {
	var dividends []iexDividend
	resp, herr := http.Get("https://cloud.iexapis.com/beta/stock/" + symbol + "/dividends/1y?token=" + portfolioConfigData.IEXToken)
	if herr != nil {
		return 0, herr
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	derr := dec.Decode(&dividends)
	if derr != nil {
		return 0, derr
	}

	// possible exDate issues, may get an extra
	if len(dividends) > 0 {
		switch dividends[0].Frequency {
		case "quarterly":
			if len(dividends) > 4 {
				dividends = dividends[:4]
			}
		case "monthly":
			if len(dividends) > 12 {
				dividends = dividends[:12]
			}
		case "semi-annual":
			if len(dividends) > 2 {
				dividends = dividends[:2]
			}
		}
	}

	for _, div := range dividends {
		amount += div.Amount
	}
	return amount, nil
}

// https://www.alphavantage.co/documentation/#weeklyadj
func fundAnnualDividends(symbol string) (amount float64, err error) {
	yearAgo := time.Now().AddDate(-1, 0, 0).Format(time.DateOnly)
	resp, herr := http.Get("https://www.alphavantage.co/query?function=TIME_SERIES_WEEKLY_ADJUSTED&datatype=csv&symbol=" + symbol + "&apikey=" + portfolioConfigData.AVToken)
	if herr != nil {
		return 0, herr
	}
	defer resp.Body.Close()
	cr := csv.NewReader(resp.Body)
	recs, cerr := cr.ReadAll()
	if cerr != nil {
		return 0, cerr
	}
	divIdx := -1
	if len(recs) < 2 {
		return 0, errors.New("csv reponse empty")
	}

	for i := range recs[0] {
		if strings.Contains(recs[0][i], "dividend") {
			divIdx = i
		}
	}

	if divIdx < 0 {
		return 0, errors.New("unable to find dividend column")
	}

	for _, rec := range recs[1:] {
		if div, derr := strconv.ParseFloat(rec[divIdx], 64); rec[0] > yearAgo && derr == nil {
			amount += div
		}
	}

	return amount, nil
}
