package cmd

import (
	"encoding/json"
	"errors"
	"net/http"
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
