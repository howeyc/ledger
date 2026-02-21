package iif_test

import (
	"bytes"
	"reflect"
	"testing"

	_ "embed"

	"github.com/howeyc/ledger/ledger/cmd/internal/import/iif"
)

var (
	//go:embed "Full Deposit.iif"
	fullDepositIIF []byte

	//go:embed "Full Invoice.iif"
	fullInvoiceIIF []byte

	//go:embed "Full Bill payment.iif"
	fullBillPaymentIIF []byte

	//go:embed "Full Sales Tax Payment.iif"
	fullSalesTaxPaymentIIF []byte

	//go:embed "Full Transfer.iif"
	fullTransferIIF []byte
)

func TestDecodeEncode(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		blocks []iif.Block
	}{
		{
			name: "fullDepositIIF",
			data: fullDepositIIF,
			blocks: []iif.Block{
				{
					Headers: []iif.Header{
						{Type: iif.RecordType("ACCNT"), Fields: []string{"NAME", "ACCNTTYPE", "DESC", "ACCNUM", "EXTRA"}},
					},
				},
				{
					Headers: []iif.Header{
						{Type: iif.RecordType("CLASS"), Fields: []string{"NAME"}},
					},
				},
				{
					Headers: []iif.Header{
						{Type: iif.RecordType("CUST"), Fields: []string{"NAME", "BADDR1", "BADDR2", "BADDR3", "BADDR4", "BADDR5", "SADDR1"}},
					},
				},
				{
					Headers: []iif.Header{
						{Type: iif.RecordType("OTHERNAME"), Fields: []string{"NAME", "BADDR1", "BADDR2", "BADDR3", "BADDR4", "BADDR5", "PHONE1", "PHONE2", "FAXNUM", "EMAIL", "NOTE", "CONT1", "CONT2", "NOTEPAD", "SALUTATION", "COMPANYNAME", "FIRSTNAME", "MIDINIT", "LASTNAME"}},
					},
				},
				{
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
									"MEMO":     "",
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
			},
		},
		{
			name: "fullInvoiceIIF",
			data: fullInvoiceIIF,
		},
		{
			name: "fullBillPaymentIIF",
			data: fullBillPaymentIIF,
		},
		{
			name: "fullSalesTaxPaymentIIF",
			data: fullSalesTaxPaymentIIF,
		},
		{
			name: "fullTransferIIF",
			data: fullTransferIIF,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dec := iif.NewDecoder(bytes.NewReader(tt.data))
			f, err := dec.Decode()
			if err != nil {
				t.Fatalf("Decode error: %v", err)
			}

			if len(f.Blocks) == 0 {
				t.Error("missing blocks from file")
			}

			for i, b := range tt.blocks {
				if i >= len(f.Blocks) {
					t.Errorf("expected at least %d blocks, got %d", len(tt.blocks), len(f.Blocks))
					break
				}
				if !reflect.DeepEqual(b.Headers, f.Blocks[i].Headers) {
					t.Errorf("expected headers to equal %+v != %+v", b.Headers, f.Blocks[i].Headers)
				}
				if b.Records != nil && !reflect.DeepEqual(b.Records, f.Blocks[i].Records) {
					t.Errorf("expected records to equal %+v != %+v", b.Records, f.Blocks[i].Records)
				}
			}
		})
	}
}
