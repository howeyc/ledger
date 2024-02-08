package cmd

import (
	"net/http"
)

func ledgerHandler(w http.ResponseWriter, r *http.Request) {
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
	pData.Init()
	pData.Transactions = trans

	err = t.Execute(w, pData)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}
