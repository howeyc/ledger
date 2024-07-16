package cmd

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/howeyc/ledger/ledger/cmd/internal/httpcompress"

	"github.com/howeyc/ledger"
	"github.com/spf13/cobra"
)

var reportConfigFileName string
var stockConfigFileName string
var quickviewConfigFileName string

var serverPort int
var localhost bool
var webReadOnly bool

//go:embed static/*
var contentStatic embed.FS

//go:embed templates/*
var contentTemplates embed.FS

func getTransactions() ([]*ledger.Transaction, error) {
	trans, terr := ledger.ParseLedgerFile(ledgerFilePath)
	if terr != nil {
		return nil, fmt.Errorf("%s", terr.Error())
	}
	slices.SortStableFunc(trans, func(a, b *ledger.Transaction) int {
		return a.Date.Compare(b.Date)
	})
	return trans, nil
}

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Web service",
	Run: func(_ *cobra.Command, _ []string) {
		configLoaders(time.Minute * 5)

		// initialize cache
		if _, err := getTransactions(); err != nil {
			log.Fatalln(err)
		}

		m := http.NewServeMux()

		fileServer := http.FileServer(http.FS(contentStatic))
		m.HandleFunc("GET /static/{filepath...}", func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Vary", "Accept-Encoding")
			w.Header().Set("Cache-Control", "public, max-age=7776000")
			req.URL.Path = "/static/" + req.PathValue("filepath")
			fileServer.ServeHTTP(w, req)
		})

		if !webReadOnly {
			m.HandleFunc("GET /addtrans", httpcompress.Middleware(addTransactionHandler, false))
			m.HandleFunc("GET /addtrans/{accountName}", httpcompress.Middleware(addQuickTransactionHandler, false))
			m.HandleFunc("POST /addtrans", httpcompress.Middleware(addTransactionPostHandler, false))
		}

		m.HandleFunc("GET /ledger", httpcompress.Middleware(ledgerHandler, false))
		m.HandleFunc("GET /accounts", httpcompress.Middleware(accountsHandler, false))
		m.HandleFunc("GET /portfolio/{portfolioName}", httpcompress.Middleware(portfolioHandler, false))
		m.HandleFunc("GET /account/{accountName}", httpcompress.Middleware(accountHandler, false))
		m.HandleFunc("GET /report/{reportName}", httpcompress.Middleware(reportHandler, false))
		m.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, req *http.Request) {
			req.URL.Path = "/static/favicon.ico"
			fileServer.ServeHTTP(w, req)
		})
		m.HandleFunc("/", httpcompress.Middleware(quickviewHandler, false))

		log.Println("Listening on port", serverPort)
		var listenAddress string
		if localhost {
			listenAddress = fmt.Sprintf("127.0.0.1:%d", serverPort)
		} else {
			listenAddress = fmt.Sprintf(":%d", serverPort)
		}
		log.Fatalln(http.ListenAndServe(listenAddress, m))
	},
}

func init() {
	rootCmd.AddCommand(webCmd)

	webCmd.Flags().StringVarP(&reportConfigFileName, "reports", "r", "", "Report config file name.")
	webCmd.Flags().StringVarP(&stockConfigFileName, "portfolio", "s", "", "Stock config file name.")
	webCmd.Flags().StringVarP(&quickviewConfigFileName, "quickview", "q", "", "Quickview config file name.")
	webCmd.Flags().IntVar(&serverPort, "port", 8056, "Port to listen on.")
	webCmd.Flags().BoolVar(&localhost, "localhost", false, "Listen on localhost only.")
	webCmd.Flags().BoolVar(&webReadOnly, "read-only", false, "Disable adding transactions through web.")
}
