package cmd

import (
	"strings"
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

func Test_predictAccount(t *testing.T) {
	tran := func(payee, account string) *ledger.Transaction {
		return &ledger.Transaction{
			Payee: payee,
			AccountChanges: []ledger.Account{
				{Name: "Assets:Checking"},
				{Name: account},
			},
		}
	}
	generalLedger := []*ledger.Transaction{
		tran("Amazon Marketplace", "Expenses:Books"),
		tran("Amazon Marketplace", "Expenses:Household"),
		tran("Shell Gas Station", "Expenses:Fuel"),
	}
	classifier := trainClassifier(generalLedger, "Assets:Checking")

	tests := []struct {
		payee string
		want  string
	}{
		// booked to two accounts equally: not a confident match
		{"Amazon Marketplace", "unknown:unknown"},
		{"Shell Gas Station", "Expenses:Fuel"},
	}
	for _, tt := range tests {
		t.Run(tt.payee, func(t *testing.T) {
			if got := predictAccount(classifier, strings.Fields(tt.payee)); got != tt.want {
				t.Errorf("predictAccount(%q) = %q, want %q", tt.payee, got, tt.want)
			}
		})
	}
}
