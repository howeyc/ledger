package main

import (
	"fmt"
	"html/template"
	"path"
)

func parseAssetsWithFunc(funcMap template.FuncMap, filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	return template.New(path.Base(filenames[0])).Funcs(funcMap).ParseFS(contentTemplates, filenames...)
}

func parseAssets(filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	return template.New(path.Base(filenames[0])).ParseFS(contentTemplates, filenames...)
}
