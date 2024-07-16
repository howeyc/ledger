package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/time/rate"
)

// Alpha Vantage allows 5 per minute, we do 4 per minute
var avLimiter *rate.Limiter = rate.NewLimiter(rate.Every(time.Minute/4), 1)

type avQuote struct {
	Symbol        string
	Open          float64
	PreviousClose float64
	Last          float64
}

// Alpha Vantage allows 500 requests per day. Since we don't care about realtime
// values, we cache results for 24 hours.
var avqCache *cache.Cache = cache.New(time.Hour*24, time.Hour)

// https://www.alphavantage.co/documentation/#latestprice
func fundQuote(symbol string) (quote avQuote, err error) {
	if avq, found := avqCache.Get(symbol); found {
		return avq.(avQuote), nil
	}

	go func() {
		avLimiter.Wait(context.Background())

		req, rerr := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://alphavantage.co/query?function=GLOBAL_QUOTE&symbol="+symbol+"&datatype=csv&apikey="+portfolioConfigData.AVToken, http.NoBody)
		if rerr != nil {
			return
		}
		resp, herr := http.DefaultClient.Do(req)
		if herr != nil {
			return
		}
		defer resp.Body.Close()
		cr := csv.NewReader(resp.Body)

		recs, cerr := cr.ReadAll()
		if cerr != nil {
			return
		}
		// symbol,open,high,low,price,volume,latestDay,previousClose,change,changePercent
		if len(recs) != 2 || len(recs[0]) != 10 {
			return
		}

		quote.Symbol = recs[1][0]
		quote.Open, _ = strconv.ParseFloat(recs[1][1], 64)
		quote.Last, _ = strconv.ParseFloat(recs[1][4], 64)
		quote.PreviousClose, _ = strconv.ParseFloat(recs[1][7], 64)

		avqCache.Add(symbol, quote, cache.DefaultExpiration)
	}()

	return quote, errors.New("not cached")
}

type gdaxQuote struct {
	Volume        string  `json:"volume"`
	PreviousClose float64 `json:"open,string"`
	Last          float64 `json:"last,string"`
}

// https://docs.pro.coinbase.com/
func cryptoQuote(symbol string) (quote gdaxQuote, err error) {
	req, rerr := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://api.pro.coinbase.com/products/"+symbol+"/stats", http.NoBody)
	if rerr != nil {
		return
	}
	resp, herr := http.DefaultClient.Do(req)
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

// Alpha Vantage allows 500 requests per day. Since we don't care about realtime
// values, we cache results for 24 hours.
var avdCache *cache.Cache = cache.New(time.Hour*24, time.Hour)

// https://www.alphavantage.co/documentation/#weeklyadj
func fundAnnualDividends(symbol string) float64 {
	if div, found := avdCache.Get(symbol); found {
		return div.(float64)
	}

	go func() {
		avLimiter.Wait(context.Background())

		req, rerr := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://www.alphavantage.co/query?function=TIME_SERIES_WEEKLY_ADJUSTED&datatype=csv&symbol="+symbol+"&apikey="+portfolioConfigData.AVToken, http.NoBody)
		if rerr != nil {
			return
		}
		resp, herr := http.DefaultClient.Do(req)
		if herr != nil {
			return
		}
		defer resp.Body.Close()
		cr := csv.NewReader(resp.Body)
		recs, cerr := cr.ReadAll()
		if cerr != nil {
			return
		}
		divIdx := -1
		if len(recs) < 2 {
			return
		}

		for i := range recs[0] {
			if strings.Contains(recs[0][i], "dividend") {
				divIdx = i
			}
		}

		if divIdx < 0 {
			return
		}

		yearAgo := time.Now().AddDate(-1, 0, 0).Format(time.DateOnly)

		var amount float64
		for _, rec := range recs[1:] {
			if div, derr := strconv.ParseFloat(rec[divIdx], 64); rec[0] > yearAgo && derr == nil {
				amount += div
			}
		}

		avdCache.Add(symbol, amount, cache.DefaultExpiration)
	}()

	return 0
}
