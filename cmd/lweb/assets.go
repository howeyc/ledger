package main

//go:generate go-bindata -o bindata.go public/... templates/...

import (
	"fmt"
	"html/template"
)

func parseAssetsWithFunc(funcMap template.FuncMap, filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	tresult := template.New("result").Funcs(funcMap)

	for _, filename := range filenames {
		tdata, aerr := Asset(filename)
		if aerr != nil {
			return nil, aerr
		}

		_, err := tresult.Parse(string(tdata))
		if err != nil {
			return nil, err
		}
	}

	return tresult, nil
}

func parseAssets(filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	tresult := template.New("result")

	for _, filename := range filenames {
		tdata, aerr := Asset(filename)
		if aerr != nil {
			return nil, aerr
		}

		_, err := tresult.Parse(string(tdata))
		if err != nil {
			return nil, err
		}
	}

	return tresult, nil
}
