package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/howeyc/ledger"

	"github.com/go-martini/martini"
	"github.com/pelletier/go-toml"
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
	ifile, ierr := os.Open(quickviewConfigFileName)
	if ierr != nil {
		http.Redirect(w, r, "/accounts", http.StatusFound)
	}
	defer ifile.Close()
	tdec := toml.NewDecoder(ifile)
	var quickviewConfigData quickviewConfigStruct
	if lerr := tdec.Decode(&quickviewConfigData); lerr != nil || len(quickviewConfigData.Accounts) < 1 {
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

	t, err := parseAssetsWithFunc(funcMap, "templates/template.quickview.html", "templates/template.nav.html")
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

	var cbuf, tbuf bytes.Buffer
	mw := io.MultiWriter(&cbuf, &tbuf)
	fmt.Fprintln(mw, date.Format("2006/01/02"), strPayee)
	for _, accLine := range accountLines {
		if len(accLine) > 0 {
			fmt.Fprintf(mw, "    %s", accLine)
			fmt.Fprintln(mw, "")
		}
	}
	fmt.Fprintln(mw, "")

	/* Check valid transaction is created */
	if _, perr := ledger.ParseLedger(&tbuf); perr != nil {
		http.Error(w, perr.Error(), 500)
		return
	}

	f, err := os.OpenFile(ledgerFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	_, err = io.Copy(f, &cbuf)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	f.Close()

	_, err = getTransactions()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, "Transaction added!")
}

func addQuickTransactionHandler(w http.ResponseWriter, r *http.Request, params martini.Params) {
	accountName := params["accountName"]
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

	// Recent accounts
	monthsago := time.Now().AddDate(0, -3, 0)
	var atrans []*ledger.Transaction
	for _, tran := range trans {
		includeTrans := false
		if tran.Date.After(monthsago) {
			includeTrans = true
			// Filter by supplied account
			if accountName != "" {
				includeTrans = false
				for _, acc := range tran.AccountChanges {
					if acc.Name == accountName {
						includeTrans = true
					}
				}
			}
		}
		if includeTrans {
			atrans = append(atrans, tran)
		}
	}

	// Child non-zero balance accounts
	balances := ledger.GetBalances(atrans, []string{})
	var abals []*ledger.Account
	for _, bal := range balances {
		accDepth := len(strings.Split(bal.Name, ":"))
		if bal.Balance.Cmp(big.NewRat(0, 1)) != 0 && accDepth > 2 {
			abals = append(abals, bal)
		}
	}

	var pData pageData
	pData.Reports = reportConfigData.Reports
	pData.Portfolios = portfolioConfigData.Portfolios
	pData.Accounts = abals
	pData.Transactions = atrans

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
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
