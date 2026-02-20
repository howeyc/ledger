package qif_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/howeyc/ledger/ledger/cmd/internal/import/qif"
)

//go:embed sample.qif
var qifSample []byte

func TestParseQIF(t *testing.T) {
	entries, err := qif.ParseQIF(bytes.NewBuffer(qifSample))
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	tests := []struct {
		index   int
		typ     string
		date    string
		amount  string
		payee   string
		memo    string
		cat     string
		splitCt string
		splitAm string
	}{
		{
			index:   0,
			typ:     "Cash",
			date:    "08/14/2024",
			amount:  "15.00",
			payee:   "",
			memo:    "~@~CLD:1723446000~@~",
			cat:     "Bank Deposit to PP Account ",
			splitCt: "Bank Deposit to PP Account ",
			splitAm: "15.00",
		},
		{
			index:   1,
			typ:     "Cash",
			date:    "08/14/2024",
			amount:  "-15.00",
			payee:   "9171-5573 Quebec Inc",
			memo:    "VOIPMS15",
			cat:     "PreApproved Payment Bill User Payment",
			splitCt: "PreApproved Payment Bill User Payment",
			splitAm: "-15.00",
		},
		{
			index:   2,
			typ:     "Cash",
			date:    "08/27/2024",
			amount:  "80.00",
			payee:   "",
			memo:    "",
			cat:     "Bank Deposit to PP Account ",
			splitCt: "Bank Deposit to PP Account ",
			splitAm: "80.00",
		},
	}

	for _, tt := range tests {
		if tt.index >= len(entries) {
			t.Fatalf("test index %d out of range, len(entries)=%d", tt.index, len(entries))
		}
		e := entries[tt.index]

		if e.Type != tt.typ {
			t.Errorf("entry %d: expected Type %q, got %q", tt.index, tt.typ, e.Type)
		}
		if e.Date != tt.date {
			t.Errorf("entry %d: expected Date %q, got %q", tt.index, tt.date, e.Date)
		}
		if e.Amount != tt.amount {
			t.Errorf("entry %d: expected Amount %q, got %q", tt.index, tt.amount, e.Amount)
		}
		if e.Payee != tt.payee {
			t.Errorf("entry %d: expected Payee %q, got %q", tt.index, tt.payee, e.Payee)
		}
		if e.Memo != tt.memo {
			t.Errorf("entry %d: expected Memo %q, got %q", tt.index, tt.memo, e.Memo)
		}
		if e.Category != tt.cat {
			t.Errorf("entry %d: expected Category %q, got %q", tt.index, tt.cat, e.Category)
		}
		if e.SplitCategory != tt.splitCt {
			t.Errorf("entry %d: expected SplitCategory %q, got %q", tt.index, tt.splitCt, e.SplitCategory)
		}
		if e.SplitAmount != tt.splitAm {
			t.Errorf("entry %d: expected SplitAmount %q, got %q", tt.index, tt.splitAm, e.SplitAmount)
		}
	}
}
