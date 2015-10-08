package main

import (
	"bytes"
	"net/http"

	"ledger"
)

func LedgerHandler(w http.ResponseWriter, r *http.Request) {
	t, err := parseAssets("templates/template.ledger.html", "templates/template.nav.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	ledgerFileReader := bytes.NewReader(ledgerBuffer.Bytes())

	trans, terr := ledger.ParseLedger(ledgerFileReader)
	if terr != nil {
		http.Error(w, terr.Error(), 500)
		return
	}

	var pData pageData
	pData.Reports = reportConfigData.Reports
	pData.Transactions = trans

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
