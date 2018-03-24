package main

import (
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/howeyc/ledger"

	"github.com/BurntSushi/toml"
	"github.com/go-martini/martini"
)

type quickviewAccountConfig struct {
	Name      string
	ShortName string `toml:"short_name"`
}

type quickviewConfigStruct struct {
	Accounts []quickviewAccountConfig `toml:"account"`
}

func quickviewHandler(w http.ResponseWriter, r *http.Request) {
	if len(quickviewConfigFileName) == 0 {
		http.Redirect(w, r, "/accounts", http.StatusFound)
	}
	var quickviewConfigData quickviewConfigStruct
	if _, lerr := toml.DecodeFile(quickviewConfigFileName, &quickviewConfigData); lerr != nil || len(quickviewConfigData.Accounts) < 1 {
		http.Redirect(w, r, "/accounts", http.StatusFound)
	}

	shorten := func(accname string) string {
		for _, qvc := range quickviewConfigData.Accounts {
			if qvc.Name == accname {
				return qvc.ShortName
			}
		}
		return abbrev(accname)
	}

	funcMap := template.FuncMap{
		"abbrev": shorten,
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

	var pData pageData
	pData.Reports = reportConfigData.Reports
	pData.Portfolios = portfolioConfigData.Portfolios
	pData.Transactions = trans

	includeNames := make(map[string]bool)
	for _, qvc := range quickviewConfigData.Accounts {
		includeNames[qvc.Name] = true
	}

	balances := ledger.GetBalances(trans, []string{})
	for _, bal := range balances {
		if includeNames[bal.Name] {
			pData.Accounts = append(pData.Accounts, bal)
		}
	}

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func addTransactionPostHandler(w http.ResponseWriter, r *http.Request) {
	strDate := r.FormValue("transactionDate")
	strPayee := r.FormValue("transactionPayee")

	var accountLines []string
	for i := 1; i < 20; i++ {
		strAcc := r.FormValue(fmt.Sprintf("transactionAccount%d", i))
		strAmt := r.FormValue(fmt.Sprintf("transactionAmount%d", i))
		accountLines = append(accountLines, strings.Trim(fmt.Sprintf("%s          %s", strAcc, strAmt), " \t"))
	}

	date, _ := time.Parse("2006-01-02", strDate)

	f, err := os.OpenFile(ledgerFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	fmt.Fprintln(f, date.Format("2006/01/02"), strPayee)
	for _, accLine := range accountLines {
		if len(accLine) > 0 {
			fmt.Fprintf(f, "    %s", accLine)
			fmt.Fprintln(f, "")
		}
	}
	fmt.Fprintln(f, "")

	f.Close()
	getTransactions()

	http.Redirect(w, r, "/addtrans", http.StatusFound)
}

func addTransactionHandler(w http.ResponseWriter, r *http.Request) {
	funcMap := template.FuncMap{
		"abbrev": abbrev,
	}

	t, err := parseAssetsWithFunc(funcMap, "templates/template.addtransaction.html", "templates/template.nav.html")
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
	pData.Portfolios = portfolioConfigData.Portfolios
	pData.Accounts = balances
	pData.Transactions = trans

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
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
	pData.Portfolios = portfolioConfigData.Portfolios
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
	var mergeTrans []*ledger.Transaction
	for _, tran := range trans {
		include := false
		bal := new(big.Rat)
		for _, accChange := range tran.AccountChanges {
			if strings.Contains(accChange.Name, accountName) {
				include = true
				bal = bal.Add(bal, accChange.Balance)
				pageTrans = append(pageTrans, &ledger.Transaction{
					Payee:          tran.Payee,
					Date:           tran.Date,
					AccountChanges: []ledger.Account{accChange},
				})
			}
		}
		if include {
			mergeTrans = append(mergeTrans, &ledger.Transaction{
				Payee:          tran.Payee,
				Date:           tran.Date,
				AccountChanges: []ledger.Account{ledger.Account{Name: accountName, Balance: bal}},
			})
		}
	}

	type accPageData struct {
		pageData
		MergedTransactions []*ledger.Transaction
	}
	var pData accPageData
	pData.Reports = reportConfigData.Reports
	pData.Portfolios = portfolioConfigData.Portfolios
	pData.Transactions = pageTrans
	pData.MergedTransactions = mergeTrans

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
