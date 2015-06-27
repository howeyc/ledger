package main

import (
	"bytes"
	"html/template"
	"net/http"

	"ledger"
)

func LedgerHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/template.ledger.html", "templates/template.nav.html")
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

	err = t.Execute(w, trans)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
