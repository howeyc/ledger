package cmd

import (
	"fmt"
	"html/template"
	"path"
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

func qvshortname(accname string) string {
	for _, qvc := range quickviewConfigData.Accounts {
		if qvc.Name == accname {
			return qvc.ShortName
		}
	}
	return abbrev(accname)
}

func loadTemplates(filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	funcMap := template.FuncMap{
		"abbrev":      abbrev,
		"lastaccount": lastaccount,
		"qvshortname": qvshortname,
		"substr":      strings.Contains,
	}

	filenames = append(filenames, "templates/template.common.html")
	return template.New(path.Base(filenames[0])).Funcs(funcMap).ParseFS(contentTemplates, filenames...)
}
