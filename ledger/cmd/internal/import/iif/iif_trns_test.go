package iif_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/howeyc/ledger/ledger/cmd/internal/import/iif"
	"github.com/shopspring/decimal"
)

func TestDeserializeTransactions(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		b       iif.Block
		want    []iif.Transaction
		wantErr bool
	}{
		{
			name: "empty",
			b: iif.Block{
				Headers: []iif.Header{
					{Type: iif.RecordType("ACCNT"), Fields: []string{"NAME", "ACCNTTYPE", "DESC", "ACCNUM", "EXTRA"}},
				},
			},
			want: nil,
		},
		{
			name: "simple",
			b: iif.Block{
				Headers: []iif.Header{
					{Type: iif.RecordType("TRNS"), Fields: []string{"TRNSID", "TRNSTYPE", "DATE", "ACCNT", "NAME", "CLASS", "AMOUNT", "DOCNUM", "MEMO", "CLEAR"}},
					{Type: iif.RecordType("SPL"), Fields: []string{"SPLID", "TRNSTYPE", "DATE", "ACCNT", "NAME", "CLASS", "AMOUNT", "DOCNUM", "MEMO", "CLEAR"}},
					{Type: iif.RecordType("ENDTRNS"), Fields: []string{}},
				},
				Records: [][]iif.Record{
					{
						{
							Type: iif.RecordType("TRNS"),
							Fields: map[string]string{
								"TRNSID":   " ",
								"TRNSTYPE": "DEPOSIT",
								"DATE":     "7/1/1998",
								"ACCNT":    "Checking",
								"NAME":     "",
								"CLASS":    "",
								"AMOUNT":   "10000",
								"DOCNUM":   "",
								"MEMO":     "Hello",
								"CLEAR":    "N",
							},
						},
						{
							Type: iif.RecordType("SPL"),
							Fields: map[string]string{
								"SPLID":    "",
								"TRNSTYPE": "DEPOSIT",
								"DATE":     "7/1/1998",
								"ACCNT":    "Income",
								"NAME":     "Customer",
								"CLASS":    "",
								"AMOUNT":   "-10000",
								"DOCNUM":   "",
								"MEMO":     "",
								"CLEAR":    "N",
							},
						},
						{
							Type:   iif.RecordType("ENDTRNS"),
							Fields: map[string]string{},
						},
					},
				},
			},
			want: []iif.Transaction{
				{
					Tr: iif.Trns{
						TransactionType: "DEPOSIT",
						Date:            time.Date(1998, 7, 1, 0, 0, 0, 0, time.UTC),
						Account:         "Checking",
						Amount:          decimal.NewFromInt(10000),
						Memo:            "Hello",
					},
					Splits: []iif.Spl{
						{
							TransactionType: "DEPOSIT",
							Date:            time.Date(1998, 7, 1, 0, 0, 0, 0, time.UTC),
							Account:         "Income",
							Name:            "Customer",
							Amount:          decimal.NewFromInt(-10000),
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := iif.DeserializeTransactions(tt.b)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("DeserializeTransactions() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("DeserializeTransactions() succeeded unexpectedly")
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DeserializeTransactions() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
