package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

var ledgerBuffer bytes.Buffer

func main() {
	var ledgerFileName string
	var serverPort int

	flag.StringVar(&ledgerFileName, "f", "", "Ledger file name (*Required).")
	flag.IntVar(&serverPort, "port", 8056, "Port to listen on.")

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

	r := mux.NewRouter()
	r.HandleFunc("/ledger", LedgerHandler).Methods("GET")
	r.HandleFunc("/accounts", AccountsHandler).Methods("GET")
	r.HandleFunc("/account/{accountName}", AccountHandler).Methods("GET")
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("webroot"))))

	http.Handle("/", r)

	fmt.Println("Listening on port", serverPort)
	http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)
}
