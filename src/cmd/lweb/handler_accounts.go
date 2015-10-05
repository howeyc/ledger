package main

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"

	"ledger"

	"github.com/go-martini/martini"
)

func AccountsHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("templates/template.accounts.html", "templates/template.nav.html")
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

	balances := ledger.GetBalances(trans, []string{})

	err = t.Execute(w, balances)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func AccountHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	accountName := params["accountName"]

	t, err := template.ParseFiles("templates/template.account.html", "templates/template.nav.html")
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

	pageTrans := make([]*ledger.Transaction, 0)
	for _, tran := range trans {
		for _, accChange := range tran.AccountChanges {
			if strings.Contains(accChange.Name, accountName) {
				pageTrans = append(pageTrans, &ledger.Transaction{
					Payee:          tran.Payee,
					Date:           tran.Date,
					AccountChanges: []ledger.Account{accChange},
				})
			}
		}
	}

	err = t.Execute(w, pageTrans)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
