package yquotes

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

const (
	// URL of Yahoo quotes for stock quotes.
	YURL = "http://finance.yahoo.com/d/quotes.csv?s=%s&f=%s"

	// Base data formating.
	// s - symbol
	// n - name
	// b - bid
	// a - ask
	// o - open
	// p - previous price
	// l1 - last price without time
	// d1 - last trade date
	BaseFormat = "snbaopl1d1"

	// Historical data URL with params:
	// s - symbol
	// a - from month (zero based)
	// b - from day
	// c - from year
	// d - to month (zero based)
	// e - to day
	// f - to year
	// g - period frequence (d - daily, w - weekly, m - monthly, y -yearly)
	YURL_H = "http://ichart.yahoo.com/table.csv?s=%s&a=%d&b=%d&c=%d&d=%d&e=%d&f=%d&g=%s&ignore=.csv"
)

// Formatting constants.
const (
	Symbol   = "s"
	Name     = "n"
	Bid      = "b"
	Ask      = "a"
	Open     = "o"
	Previous = "p"
	Last     = "l1"
	LastDate = "d1"
)

// Constance of frequesncy for historical data requests.
const (
	Daily   = "d"
	Weekly  = "w"
	Monthly = "m"
	Yearly  = "y"
)

// Price struct represents price in single point in time.
type Price struct {
	Bid           float64   `json:"bid,omitempty"`
	Ask           float64   `json:"ask,omitempty"`
	Open          float64   `json:"open,omitempty"`
	PreviousClose float64   `json:"previousClose,omitempty"`
	Last          float64   `json:"last,omitempty"`
	Date          time.Time `json:"date,omitempty"`
}

// Price type that is used for historical price data.
type PriceH struct {
	Date     time.Time `json:"date,omitempty"`
	Open     float64   `json:"open,omitempty"`
	High     float64   `json:"high,omitempty"`
	Low      float64   `json:"low,omitempty"`
	Close    float64   `json:"close,omitempty"`
	Volume   float64   `json:"volume,omitempty"`
	AdjClose float64   `json:"adjClose,omitempty"`
}

// Stock is used as container for stock price data.
type Stock struct {
	// Symbol of stock that should meet requirements of Yahoo. Otherwise,
	// there will be no possibility to find stock.
	Symbol string `json:"symbol,omitempty"`

	// Name of the company will be filled from request of stock data.
	Name string `json:"name,omitempty"`

	// Information about last price of stock.
	Price *Price `json:"price,omitempty"`

	// Contains historical price information. If client asks information
	// for recent price, this field will be omited.
	History []PriceH `json:"history,omitempty"`
}

// Create new stock with recent price data and historical prices. All the prices
// are represented in daily basis.
//
// symbol - Symbol of company (ticker)
// history - If true 3 year price history will be loaded.
// Returns a pointer to value of Stock type.
func NewStock(symbol string, history bool) (*Stock, error) {
	var stock *Stock

	data, err := loadSingleStockPrice(symbol)
	if err != nil {
		return nil, err
	}
	stock = parseStock(data)

	// If history is TRUE, we need to load historical price for 3 year period.
	if history == true {
		histPrices, err := HistoryForYears(symbol, 3, Daily)
		if err != nil {
			return stock, err
		}
		stock.History = histPrices
	}

	return stock, nil
}

// Get single stock price data.
func GetPrice(symbol string) (*Price, error) {
	data, err := loadSingleStockPrice(symbol)
	if err != nil {
		return nil, err
	}

	price := parsePrice(data)
	return price, nil
}

// Get single stock price for certain date.
func GetPriceForDate(symbol string, date time.Time) (PriceH, error) {
	var price PriceH
	// We need to get price information for single date, so date passed to
	// this function will be used both for "from" and "to" arguments.
	data, err := loadHistoricalPrice(symbol, date, date, Daily)
	if err != nil {
		return price, err
	}

	prices, err := parseHistorical(data)
	if err != nil {
		return PriceH{}, err
	}
	price = prices[0]

	// Return single price.
	return price, nil
}

// Get historical prices for the stock.
func GetDailyHistory(symbol string, from, to time.Time) ([]PriceH, error) {
	var prices []PriceH
	// Create URL with daily frequency of data.
	data, err := loadHistoricalPrice(symbol, from, to, Daily)
	if err != nil {
		return prices, err
	}

	prices, err = parseHistorical(data)
	if err != nil {
		return prices, err
	}

	return prices, nil
}

// Get stock price history for number of years backwards.
func HistoryForYears(symbol string, years int, period string) ([]PriceH, error) {
	var prices []PriceH
	duration := time.Duration(int64(time.Hour) * 24 * 365 * int64(years))
	to := time.Now()
	from := to.Add(-duration)

	data, err := loadHistoricalPrice(symbol, from, to, period)
	if err != nil {
		return prices, err
	}

	prices, err = parseHistorical(data)
	if err != nil {
		return prices, err
	}

	return prices, nil
}

// Single company data request to Yahoo Finance.
func loadSingleStockPrice(symbol string) ([]string, error) {
	url := singleStockUrl(symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	data, err := reader.Read()
	if err != nil {
		return nil, err
	}

	return data, err
}

// Load historical data price.
func loadHistoricalPrice(symbol string, from, to time.Time, period string) ([][]string, error) {
	url := stockHistoryURL(symbol, from, to, period)
	var data [][]string

	resp, err := http.Get(url)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	reader := csv.NewReader(resp.Body)
	data, err = reader.ReadAll()
	if err != nil {
		return data, err
	}

	return data, nil
}

// Generate request URL for single stock.
func singleStockUrl(symbol string) string {
	url := fmt.Sprintf(YURL, symbol, BaseFormat)
	return url
}

// Generate request URL for historicat stock data.
func stockHistoryURL(symbol string, from, to time.Time, frequency string) string {
	// From date
	fMonth := (from.Month() - 1) // Need to subtract 1 because months in query is 0 based.
	fDay := from.Day()
	fYear := from.Year()
	// To date
	tMonth := (to.Month() - 1)
	tDay := to.Day()
	tYear := to.Year()

	url := fmt.Sprintf(
		YURL_H,
		symbol,
		fMonth,
		fDay,
		fYear,
		tMonth,
		tDay,
		tYear,
		frequency)

	return url
}

// Parse base information of stock price. Base inforamtion is reperesented by
// request format BaseFormat.
func parseStock(data []string) *Stock {
	s := &Stock{
		Symbol: data[0],
		Name:   data[1],
		Price:  parsePrice(data),
	}

	return s
}

// Parse collection of historical prices.
func parseHistorical(data [][]string) ([]PriceH, error) {
	// This is the list of prices with allocated space. Length of space should
	// subtracted by 1 because the first row of data is title.
	var list = make([]PriceH, len(data)-1)
	// We need to leave the first row, because it contains title of columns.
	for k, v := range data {
		if k == 0 {
			continue
		}
		// Parse row of data into PriceH type and append it to collection of prices.
		p, err := parseHistoricalRow(v)
		if err != nil {
			return list, err
		}

		// (k - 1) because we remove header from the list so index should be
		// reduced by one.
		list[k-1] = p
	}
	return list, nil
}

// Parse data row that comes from historical data. Data row contains
// 7 columns:
// 0 - Date
// 1 - Open
// 2 - High
// 3 - Low
// 4 - Close
// 5 - Volume
// 6 - Adj Close
// This function will return PriceH type that wraps all these columns.
func parseHistoricalRow(data []string) (PriceH, error) {
	p := PriceH{}

	// Parse date.
	d, err := time.Parse("2006-01-02", data[0])
	if err != nil {
		return p, err
	}

	p.Date = d
	p.Open, _ = strconv.ParseFloat(data[1], 64)
	p.High, _ = strconv.ParseFloat(data[2], 64)
	p.Low, _ = strconv.ParseFloat(data[3], 64)
	p.Close, _ = strconv.ParseFloat(data[4], 64)
	p.Volume, _ = strconv.ParseFloat(data[5], 64)
	p.AdjClose, _ = strconv.ParseFloat(data[6], 64)

	return p, nil
}

// Parse date from string into time.Time type.
func parseDate(date string) (time.Time, error) {
	d, err := time.Parse("1/2/2006", date)
	if err != nil {
		return time.Time{}, err
	}

	return d, nil
}

// Parse price information from base data.
func parsePrice(data []string) *Price {
	p := &Price{}
	p.Bid, _ = strconv.ParseFloat(data[2], 64)
	p.Ask, _ = strconv.ParseFloat(data[3], 64)
	p.Open, _ = strconv.ParseFloat(data[4], 64)
	p.PreviousClose, _ = strconv.ParseFloat(data[5], 64)
	p.Last, _ = strconv.ParseFloat(data[6], 64)
	p.Date, _ = parseDate(data[7])

	return p
}
