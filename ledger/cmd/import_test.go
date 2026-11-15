package cmd

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/howeyc/ledger"
)

func Test_findMatchingAccount(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		generalLedger    []*ledger.Transaction
		accountSubstring string
		want             string
		wantErr          bool
	}{
		{
			"simple test",
			[]*ledger.Transaction{
				&ledger.Transaction{
					AccountChanges: []ledger.Account{
						{Name: "Equity:Fake"},
						{Name: "Liability:Real"},
					},
				},
			},
			"Fake",
			"Equity:Fake",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := findMatchingAccount(tt.generalLedger, tt.accountSubstring)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("findMatchingAccount() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("findMatchingAccount() succeeded unexpectedly")
			}
			// TODO: update the condition below to compare got with tt.want.
			if got != tt.want {
				t.Errorf("findMatchingAccount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_importQFX_signs(t *testing.T) {
	dir := t.TempDir()
	ledgerFile := filepath.Join(dir, "l.ledger")
	qfxFile := filepath.Join(dir, "bank.qfx")
	if err := os.WriteFile(ledgerFile, []byte("2026/10/01 Opening\n    Assets:Checking   1000\n    Equity:Opening\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// TRNAMT is the change of the bank account: a card purchase is negative
	qfxData := `<?xml version="1.0" ?>
<OFX><BANKMSGSRSV1><STMTTRNRS><STMTRS><BANKTRANLIST>
<STMTTRN><TRNTYPE>DEBIT</TRNTYPE><DTPOSTED>20261003</DTPOSTED><TRNAMT>-12.34</TRNAMT><FITID>1</FITID><MEMO>STARBUCKS</MEMO></STMTTRN>
</BANKTRANLIST></STMTRS></STMTTRNRS></BANKMSGSRSV1></OFX>`
	if err := os.WriteFile(qfxFile, []byte(qfxData), 0o600); err != nil {
		t.Fatal(err)
	}

	oldFile, oldScale, oldStdout := ledgerFilePath, scaleFactor, os.Stdout
	defer func() { ledgerFilePath, scaleFactor, os.Stdout = oldFile, oldScale, oldStdout }()
	ledgerFilePath, scaleFactor = ledgerFile, 1
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	importQFX("Checking", qfxFile)
	w.Close()
	out, _ := io.ReadAll(r)

	if !regexp.MustCompile(`Assets:Checking\s+-12.34\n`).Match(out) {
		t.Errorf("purchase should lower Assets:Checking, got:\n%s", out)
	}
}
