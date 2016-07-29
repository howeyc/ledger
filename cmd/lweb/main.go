package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
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
var ledgerLock sync.Mutex

func getTransactions() ([]*ledger.Transaction, error) {
	ledgerLock.Lock()
	defer ledgerLock.Unlock()

	ledgerFileReader, err := os.Open(ledgerFileName)
	if err != nil {
		return nil, err

	}

	trans, terr := ledger.ParseLedger(ledgerFileReader)
	if terr != nil {
		return nil, fmt.Errorf("%s:%s", ledgerFileName, terr.Error())
	}
	ledgerFileReader.Close()

	return trans, nil
}

type accountOp struct {
	Name      string `toml:"name"`
	Operation string `toml:"operation"` // +, -
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

type pageData struct {
	Reports      []reportConfig
	Transactions []*ledger.Transaction
	Accounts     []*ledger.Account
}

func main() {
	var serverPort int
	var localhost bool

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.StringVar(&reportConfigFileName, "r", "", "Report config file name (*Required).")
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

	m := martini.Classic()
	m.Use(gzip.All())
	m.Use(staticbin.Static("public", Asset))

	m.Get("/ledger", ledgerHandler)
	m.Get("/accounts", accountsHandler)
	m.Get("/account/:accountName", accountHandler)
	m.Get("/report/:reportName", reportHandler)
	m.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/accounts", http.StatusFound)
	})

	fmt.Println("Listening on port", serverPort)
	listenAddress := ""
	if localhost {
		listenAddress = fmt.Sprintf("127.0.0.1:%d", serverPort)
	} else {
		listenAddress = fmt.Sprintf(":%d", serverPort)
	}
	http.ListenAndServe(listenAddress, m)
}
