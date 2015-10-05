package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/go-martini/martini"
)

var ledgerBuffer bytes.Buffer

func main() {
	var ledgerFileName string
	var serverPort int
	var localhost bool

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.IntVar(&serverPort, "port", 8056, "Port to listen on.")
	flag.BoolVar(&localhost, "localhost", false, "Listen on localhost only.")

	flag.Parse()

	if len(ledgerFileName) == 0 {
		flag.Usage()
		return
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

	fmt.Println("Listening on port", serverPort)
	listenAddress := ""
	if localhost {
		listenAddress = fmt.Sprintf("127.0.0.1:%d", serverPort)
	} else {
		listenAddress = fmt.Sprintf(":%d", serverPort)
	}
	http.ListenAndServe(listenAddress, m)
}
