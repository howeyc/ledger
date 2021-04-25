package main

import (
	"bytes"
	"crypto/sha256"
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/howeyc/ledger"
	"github.com/howeyc/ledger/cmd/lweb/internal/httpcompress"

	"github.com/julienschmidt/httprouter"
)

var ledgerFileName string
var reportConfigFileName string
var stockConfigFileName string
var quickviewConfigFileName string
var ledgerLock sync.Mutex
var currentSum []byte
var currentTrans []*ledger.Transaction

//go:embed static/*
var contentStatic embed.FS

//go:embed templates/*
var contentTemplates embed.FS

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
	_, err = io.Copy(&buf, tr)
	if err != nil {
		return nil, err
	}

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

	configLoaders(time.Minute * 5)

	// initialize cache
	if _, err := getTransactions(); err != nil {
		log.Fatalln(err)
	}

	m := httprouter.New()

	fileServer := http.FileServer(http.FS(contentStatic))
	m.GET("/static/*filepath", func(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		req.URL.Path = "/static/" + ps.ByName("filepath")
		fileServer.ServeHTTP(w, req)
	})

	m.GET("/ledger", httpcompress.Middleware(ledgerHandler, false))
	m.GET("/accounts", httpcompress.Middleware(accountsHandler, false))
	m.GET("/addtrans", httpcompress.Middleware(addTransactionHandler, false))
	m.GET("/addtrans/:accountName", httpcompress.Middleware(addQuickTransactionHandler, false))
	m.POST("/addtrans", httpcompress.Middleware(addTransactionPostHandler, false))
	m.GET("/portfolio/:portfolioName", httpcompress.Middleware(portfolioHandler, false))
	m.GET("/account/:accountName", httpcompress.Middleware(accountHandler, false))
	m.GET("/report/:reportName", httpcompress.Middleware(reportHandler, false))
	m.GET("/favicon.ico", func(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		req.URL.Path = "/static/favicon.ico"
		fileServer.ServeHTTP(w, req)
	})
	m.GET("/", httpcompress.Middleware(quickviewHandler, false))

	log.Println("Listening on port", serverPort)
	var listenAddress string
	if localhost {
		listenAddress = fmt.Sprintf("127.0.0.1:%d", serverPort)
	} else {
		listenAddress = fmt.Sprintf(":%d", serverPort)
	}
	log.Fatalln(http.ListenAndServe(listenAddress, m))
}
