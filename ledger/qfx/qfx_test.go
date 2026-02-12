package qfx_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/howeyc/ledger/ledger/qfx"
)

//go:embed sample.qfx
var qfxSample []byte

func TestParseQFX(t *testing.T) {
	entries, err := qfx.ParseQFX(bytes.NewBuffer(qfxSample))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 26 {
		t.Fatalf("Expected 26 entries, got %d", len(entries))
	}

	// Spot-check a few transactions to ensure fields are parsed correctly.
	tests := []struct {
		index    int
		trnType  string
		dtPosted string
		trnAmt   string
		fitID    string
		memo     string
	}{
		{
			index:    0,
			trnType:  "CREDIT",
			dtPosted: "20251231000000.000",
			trnAmt:   "0.13",
			fitID:    "202512311",
			memo:     "IOD INTEREST PAID",
		},
		{
			index:    6,
			trnType:  "DEBIT",
			dtPosted: "20250829000000.000",
			trnAmt:   "-30",
			fitID:    "202508292",
			memo:     "Minimum balance charge",
		},
		{
			index:    14,
			trnType:  "DEBIT",
			dtPosted: "20250609000000.000",
			trnAmt:   "-200",
			fitID:    "202506091",
			memo:     "ACH Withdrawal CAPITAL ONE",
		},
		{
			index:    21,
			trnType:  "DEBIT",
			dtPosted: "20250219000000.000",
			trnAmt:   "-620",
			fitID:    "202502192",
			memo:     "ACH Withdrawal",
		},
		{
			index:    25,
			trnType:  "CREDIT",
			dtPosted: "20250123000000.000",
			trnAmt:   "11892",
			fitID:    "202501231",
			memo:     "ACH deposit INTERACTIVE BROK ACH TRANSF",
		},
	}

	for _, tt := range tests {
		if tt.index >= len(entries) {
			t.Fatalf("test index %d out of range, len(entries)=%d", tt.index, len(entries))
		}
		e := entries[tt.index]

		if e.TrnType != tt.trnType {
			t.Errorf("entry %d: expected TrnType %q, got %q", tt.index, tt.trnType, e.TrnType)
		}
		if e.DtPosted != tt.dtPosted {
			t.Errorf("entry %d: expected DtPosted %q, got %q", tt.index, tt.dtPosted, e.DtPosted)
		}
		if e.TrnAmt != tt.trnAmt {
			t.Errorf("entry %d: expected TrnAmt %q, got %q", tt.index, tt.trnAmt, e.TrnAmt)
		}
		if e.FitID != tt.fitID {
			t.Errorf("entry %d: expected FitID %q, got %q", tt.index, tt.fitID, e.FitID)
		}
		if e.Memo != tt.memo {
			t.Errorf("entry %d: expected Memo %q, got %q", tt.index, tt.memo, e.Memo)
		}
	}
}
