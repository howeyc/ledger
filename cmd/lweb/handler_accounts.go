package main

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/howeyc/ledger"
	"github.com/juztin/numeronym"

	"github.com/go-martini/martini"
)

func abbrev(acctName string) string {
	accounts := strings.Split(acctName, ":")
	shortAccounts := make([]string, len(accounts))
	for i := range accounts[:len(accounts)-1] {
		shortAccounts[i] = string(numeronym.Parse([]byte(accounts[i])))
	}
	shortAccounts[len(accounts)-1] = accounts[len(accounts)-1]
	return strings.Join(shortAccounts, ":")
}

func accountsHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"abbrev": abbrev,
	}

	t, err := parseAssetsWithFunc(funcMap, "templates/template.accounts.html", "templates/template.nav.html")
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
