package ledger

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"
	"time"
)

type testCase struct {
	data         string
	transactions []*Transaction
	err          error
}

var testCases = []testCase{
	testCase{
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Assets
`,
		[]*Transaction{
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense/test",
						big.NewRat(369.0, 1),
					},
					Account{
						"Assets",
						big.NewRat(-369.0, 1),
					},
				},
			},
		},
		nil,
	},
}

func TestParseLedger(t *testing.T) {
	for _, tc := range testCases {
		b := bytes.NewBufferString(tc.data)
		transactions, err := ParseLedger(b)
		if err != tc.err {
			t.Errorf("Error: expected `%s`, got `%s`", tc.err, err)
		}
		exp, _ := json.Marshal(tc.transactions)
		got, _ := json.Marshal(transactions)
		if string(exp) != string(got) {
			t.Errorf("Error: expected \n`%s`, \ngot \n`%s`", exp, got)
		}
	}
}
