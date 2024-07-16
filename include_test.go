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
	if err.Error() != "testdata/ledger-2021-05.dat:12: unable to parse transaction: unable to balance transaction: no empty account to place extra balance" {
		t.Fatal(err)
	}
}

func TestIncludeNonExistant(t *testing.T) {
	_, err := ParseLedgerFile("testdata/ledgerRootNonExist.dat")
	if err.Error() != "testdata/ledgerRootNonExist.dat:3: unable to include file(ledger-xxxxx.dat): not found" {
		t.Fatal(err)
	}
}

func TestNonExistant(t *testing.T) {
	_, err := ParseLedgerFile("testdata/ledger-xxxxx.dat")
	if err.Error() != "open testdata/ledger-xxxxx.dat: no such file or directory" {
		t.Fatal(err)
	}
}
