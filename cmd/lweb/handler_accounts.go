package main

import (
	"net/http"
	"strings"

	"github.com/howeyc/ledger"

	"github.com/go-martini/martini"
)

func accountsHandler(w http.ResponseWriter, r *http.Request) {
	t, err := parseAssets("templates/template.accounts.html", "templates/template.nav.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	trans, terr := getTransactions()
	if terr != nil {
		http.Error(w, terr.Error(), 500)
		return
	}

	balances := ledger.GetBalances(trans, []string{})

	var pData pageData
	pData.Reports = reportConfigData.Reports
	pData.Accounts = balances
	pData.Transactions = trans

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func accountHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	accountName := params["accountName"]

	t, err := parseAssets("templates/template.account.html", "templates/template.nav.html")
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	trans, terr := getTransactions()
	if terr != nil {
		http.Error(w, terr.Error(), 500)
		return
	}

	var pageTrans []*ledger.Transaction
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

	var pData pageData
	pData.Reports = reportConfigData.Reports
	pData.Transactions = pageTrans

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
