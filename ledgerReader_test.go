package ledger

import (
	"testing"
)

func TestReaderSimple(t *testing.T) {
	_, err := NewLedgerReader("testdata/ledgerRoot.dat")
	if err != nil {
		t.Fatal(err)
	}
}

func TestReaderNonExistant(t *testing.T) {
	_, err := NewLedgerReader("testdata/ledger-xxxxx.dat")
	if err.Error() != "open testdata/ledger-xxxxx.dat: no such file or directory" {
		t.Fatal(err)
	}
}
