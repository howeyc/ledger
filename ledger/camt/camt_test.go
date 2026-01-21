package camt_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/howeyc/ledger/ledger/camt"
)

//go:embed sample.xml
var camtSample []byte

func TestParseCamt(t *testing.T) {
	entries, err := camt.ParseCamt(bytes.NewBuffer(camtSample))
	if err != nil {
		t.Error(err)
	}
	if len(entries) != 2 {
		t.Error("Expected 2 got ", len(entries))
	}
}
