package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/howeyc/ledger"

	"github.com/BurntSushi/toml"
	"github.com/go-martini/martini"
	"github.com/martini-contrib/gzip"
	"github.com/martini-contrib/staticbin"
)

var ledgerFileName string
var reportConfigFileName string
var stockConfigFileName string
var quickviewConfigFileName string
var ledgerLock sync.Mutex
var currentSum []byte
var currentTrans []*ledger.Transaction

func getTransactions() ([]*ledger.Transaction, error) {
	ledgerLock.Lock()
	defer ledgerLock.Unlock()

	var buf bytes.Buffer
	h := sha256.New()

	ledgerFileReader, err := ledger.NewLedgerReader(ledgerFileName)
	if err != nil {
		return nil, err

	}
	tr := io.TeeReader(ledgerFileReader, h)
	io.Copy(&buf, tr)

	sum := h.Sum(nil)
	if bytes.Equal(currentSum, sum) {
		return currentTrans, nil
	}

	trans, terr := ledger.ParseLedger(&buf)
	if terr != nil {
		return nil, fmt.Errorf("%s", terr.Error())
	}
	currentSum = sum
	currentTrans = trans

	return trans, nil
}

type accountOp struct {
	Name                 string  `toml:"name"`
	Operation            string  `toml:"operation"` // +, -
	MultiplicationFactor float64 `toml:"factor"`
	SubAccount           string  `toml:"other_account"` // *, /
}

type calculatedAccount struct {
	Name              string      `toml:"name"`
	AccountOperations []accountOp `toml:"account_operation"`
}

type reportConfig struct {
	Name                   string
	Chart                  string
	DateRange              string `toml:"date_range"`
	DateFreq               string `toml:"date_freq"`
	Accounts               []string
	ExcludeAccountTrans    []string            `toml:"exclude_account_trans"`
	ExcludeAccountsSummary []string            `toml:"exclude_account_summary"`
	CalculatedAccounts     []calculatedAccount `toml:"calculated_account"`
}

type reportConfigStruct struct {
	Reports []reportConfig `toml:"report"`
}

var reportConfigData reportConfigStruct

type stockConfig struct {
	Name         string
	SecurityType string `toml:"security_type"`
	Section      string
	Ticker       string
	Account      string
	Shares       float64
}

type stockInfo struct {
	Name                  string
	Section               string
	Type                  string
	Ticker                string
	Account               string
	Shares                float64
	Price                 float64
	PriceChangeDay        float64
	PriceChangePctDay     float64
	PriceChangeOverall    float64
	PriceChangePctOverall float64
	Cost                  float64
	MarketValue           float64
	GainLossDay           float64
	GainLossOverall       float64
}

type portfolioStruct struct {
	Name   string
	Stocks []stockConfig `toml:"stock"`
}

type portfolioConfigStruct struct {
	Portfolios []portfolioStruct `toml:"portfolio"`
}

var portfolioConfigData portfolioConfigStruct

type pageData struct {
	Reports      []reportConfig
	Transactions []*ledger.Transaction
	Accounts     []*ledger.Account
	Stocks       []stockInfo
	Portfolios   []portfolioStruct
}

func main() {
	var serverPort int
	var localhost bool

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.StringVar(&reportConfigFileName, "r", "", "Report config file name (*Required).")
	flag.StringVar(&stockConfigFileName, "s", "", "Stock config file name (*Optional).")
	flag.StringVar(&quickviewConfigFileName, "q", "", "Quickview config file name (*Optional).")
	flag.IntVar(&serverPort, "port", 8056, "Port to listen on.")
	flag.BoolVar(&localhost, "localhost", false, "Listen on localhost only.")

	flag.Parse()

	if len(ledgerFileName) == 0 || len(reportConfigFileName) == 0 {
		flag.Usage()
		return
	}

	go func() {
		for {
			var rLoadData reportConfigStruct
			toml.DecodeFile(reportConfigFileName, &rLoadData)
			reportConfigData = rLoadData
			time.Sleep(time.Minute * 5)
		}
	}()

	if len(stockConfigFileName) > 0 {
		go func() {
			for {
				var sLoadData portfolioConfigStruct
				toml.DecodeFile(stockConfigFileName, &sLoadData)
				portfolioConfigData = sLoadData
				time.Sleep(time.Minute * 5)
			}
		}()
	}

	// initialize cache
	getTransactions()

	m := martini.Classic()
	m.Use(gzip.All())
	m.Use(staticbin.Static("public", Asset))

	m.Get("/ledger", ledgerHandler)
	m.Get("/accounts", accountsHandler)
	m.Get("/addtrans", addTransactionHandler)
	m.Get("/addtrans/:accountName", addQuickTransactionHandler)
	m.Post("/addtrans", addTransactionPostHandler)
	m.Get("/portfolio/:portfolioName", portfolioHandler)
	m.Get("/account/:accountName", accountHandler)
	m.Get("/report/:reportName", reportHandler)
	m.Get("/", quickviewHandler)

	fmt.Println("Listening on port", serverPort)
	var listenAddress string
	if localhost {
		listenAddress = fmt.Sprintf("127.0.0.1:%d", serverPort)
	} else {
		listenAddress = fmt.Sprintf(":%d", serverPort)
	}
	http.ListenAndServe(listenAddress, m)
}
