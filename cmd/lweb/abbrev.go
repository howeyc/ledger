package main

import (
	"strings"

	"github.com/juztin/numeronym"
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

func lastaccount(acctName string) string {
	accounts := strings.Split(acctName, ":")
	return accounts[len(accounts)-1]
}
