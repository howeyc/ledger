package cmd

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

func ledgerHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	t, err := loadTemplates("templates/template.ledger.html")
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

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
