package ledger

import (
	"errors"
	"testing"
)

func TestIncludeSimple(t *testing.T) {
	trans, err := ParseLedgerFile("testdata/ledgerRoot.dat")
	if err != nil {
		t.Fatal(err)
	}
	bals := GetBalances(trans, []string{"Assets"})
	if bals[0].Balance.StringRound() != "50" {
		t.Fatal(errors.New("should be 50"))
	}
}

func TestIncludeGlob(t *testing.T) {
	trans, err := ParseLedgerFile("testdata/ledgerRootGlob.dat")
	if err != nil {
		t.Fatal(err)
	}
	bals := GetBalances(trans, []string{"Assets"})
	if bals[0].Balance.StringRound() != "80" {
		t.Fatal(errors.New("should be 80"))
	}
}

func TestIncludeUnbalanced(t *testing.T) {
	_, err := ParseLedgerFile("testdata/ledgerRootUnbalanced.dat")
	if err.Error() != "testdata/ledger-2021-05.dat:12: Unable to parse transaction: Unable to balance transaction: no empty account to place extra balance" {
		t.Fatal(err)
	}
}
