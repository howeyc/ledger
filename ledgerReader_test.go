package ledger

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"
)

func TestLedgerScannerBasic(t *testing.T) {
	r, err := NewLedgerReader("testdata/ledgerReader_input_0")
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	expected, _ := ioutil.ReadFile(filepath.Join("testdata", "ledgerReader_expected_0"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(parsed, expected) {
		t.Fatalf("expected:\n%s\n\ngot:\n%s\n", expected, parsed)
	}
}

func TestLedgerScannerSingleInclude(t *testing.T) {
	r, err := NewLedgerReader("testdata/ledgerReader_input_1_root")
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	expected, _ := ioutil.ReadFile(filepath.Join("testdata", "ledgerReader_expected_1"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(parsed, expected) {
		t.Fatalf("expected:\n%s\n\n got:\n%s", expected, parsed)
	}
}

func TestLedgerScannerWildcardInclude(t *testing.T) {
	r, err := NewLedgerReader("testdata/ledgerReader_input_wildcard_root")
	if err != nil {
		t.Fatal(err)
	}

	parsed, err := ioutil.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}

	expected, _ := ioutil.ReadFile(filepath.Join("testdata", "ledgerReader_expected_wildcard"))
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(parsed, expected) {
		t.Fatalf("expected:\n%s\n\n got:\n%s", expected, parsed)
	}
}

func TestMarkerSplit(t *testing.T) {
	filename, lineNum := parseMarker(";__ledger_file*-*/somedir/somefile*-*45")
	if filename != "/somedir/somefile" {
		t.Fatalf("expected: %s got:%s", "/somedir/somefile", filename)
	}
	if lineNum != 45 {
		t.Fatalf("expected: %d got:%d", 45, lineNum)
	}
}
