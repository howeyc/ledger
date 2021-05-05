package cmd

import (
	"bytes"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/howeyc/ledger"
	"github.com/julienschmidt/httprouter"
)

func quickviewHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if len(quickviewConfigData.Accounts) < 1 {
		http.Redirect(w, r, "/accounts", http.StatusFound)
		return
	}

	t, err := loadTemplates("templates/template.quickview.html")
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

func addTransactionPostHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	strDate := r.FormValue("transactionDate")
	strPayee := r.FormValue("transactionPayee")

	var accountLines []string
	for i := 1; i < 20; i++ {
		strAcc := r.FormValue(fmt.Sprintf("transactionAccount%d", i))
		strAmt := r.FormValue(fmt.Sprintf("transactionAmount%d", i))
		accountLines = append(accountLines, strings.Trim(fmt.Sprintf("%s          %s", strAcc, strAmt), " \t"))
	}

	date, _ := time.Parse("2006-01-02", strDate)

	var tbuf bytes.Buffer
	fmt.Fprintln(&tbuf, date.Format("2006/01/02"), strPayee)
	for _, accLine := range accountLines {
		if len(accLine) > 0 {
			fmt.Fprintf(&tbuf, "    %s", accLine)
			fmt.Fprintln(&tbuf, "")
		}
	}
	fmt.Fprintln(&tbuf, "")

	/* Check valid transaction is created */
	if trans, perr := ledger.ParseLedger(&tbuf); perr != nil {
		http.Error(w, perr.Error(), 500)
		return
	} else {
		f, err := os.OpenFile(ledgerFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		for _, t := range trans {
			WriteTransaction(f, t, 80)
		}

		f.Close()
	}

	if _, err := getTransactions(); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	fmt.Fprintf(w, "Transaction added!")
}

func addQuickTransactionHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	accountName := params.ByName("accountName")

	t, err := loadTemplates("templates/template.addtransaction.html")
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
	pData.AccountNames = []string{accountName}

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func addTransactionHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := loadTemplates("templates/template.addtransaction.html")
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

func accountsHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := loadTemplates("templates/template.accounts.html")
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

func accountHandler(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	accountName := params.ByName("accountName")

	t, err := loadTemplates("templates/template.account.html")
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
	pData.Portfolios = portfolioConfigData.Portfolios
	pData.Transactions = pageTrans
	pData.AccountNames = []string{accountName}

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
