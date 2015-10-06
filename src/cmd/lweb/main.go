package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"ledger"

	"github.com/BurntSushi/toml"
	"github.com/go-martini/martini"
)

var ledgerBuffer bytes.Buffer

type reportConfig struct {
	Name      string
	Chart     string
	DateRange string `toml:"date_range"`
	DateFreq  string `toml:"date_freq"`
	Accounts  []string
	Exclude   []string `toml:"exclude"`
}

var reportConfigData struct {
	Reports []reportConfig `toml:"report"`
}

type pageData struct {
	Reports      []reportConfig
	Transactions []*ledger.Transaction
	Accounts     []*ledger.Account
}

func main() {
	var ledgerFileName string
	var reportConfigFileName string
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

	_, derr := toml.DecodeFile(reportConfigFileName, &reportConfigData)
	if derr != nil {
		fmt.Println(derr)
	}

	ledgerFileReader, err := os.Open(ledgerFileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	io.Copy(&ledgerBuffer, ledgerFileReader)
	ledgerFileReader.Close()

	m := martini.Classic()
	m.Get("/ledger", LedgerHandler)
	m.Get("/accounts", AccountsHandler)
	m.Get("/account/:accountName", AccountHandler)
	m.Get("/report/:reportName", ReportHandler)

	fmt.Println("Listening on port", serverPort)
	listenAddress := ""
	if localhost {
		listenAddress = fmt.Sprintf("127.0.0.1:%d", serverPort)
	} else {
		listenAddress = fmt.Sprintf(":%d", serverPort)
	}
	http.ListenAndServe(listenAddress, m)
}
