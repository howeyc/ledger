package main

import (
	"fmt"
	"strings"

	"github.com/jbrukh/bayesian"
)

func classifyPayee(generalLedger []*Transaction, balances []*Account, payee string) {
	classes := make([]bayesian.Class, len(balances))
	for i, bal := range balances {
		classes[i] = bayesian.Class(bal.Name)
	}
	classifier := bayesian.NewClassifier(classes...)
	for _, tran := range generalLedger {
		payeeWords := strings.Split(tran.Payee, " ")
		for _, accChange := range tran.AccountChanges {
			classifier.Learn(payeeWords, bayesian.Class(accChange.Name))
		}
	}
	inputPayeeWords := strings.Split(payee, " ")
	_, likely, _ := classifier.LogScores(inputPayeeWords)
	fmt.Println(classifier.Classes[likely])
}
