package ledger

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	testCase{
		`1970/01/01 Payee
 Expense:test	369.0
 Assets

; Handle tabs between account and amount
; Also handle accounts with spaces
1970/01/01 Payee 5
	Expense:Cars R Us
	Expense:Cars  358.0
	Expense:Cranks	10
	Expense:Cranks Unlimited	10
	Expense:Cranks United  10
`,
		[]*Transaction{
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense:test",
						big.NewRat(369.0, 1),
					},
					Account{
						"Assets",
						big.NewRat(-369.0, 1),
					},
				},
			},
			&Transaction{
				Payee: "Payee 5",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense:Cars R Us",
						big.NewRat(-388.0, 1),
					},
					Account{
						"Expense:Cars",
						big.NewRat(358.0, 1),
					},
					Account{
						"Expense:Cranks",
						big.NewRat(10.0, 1),
					},
					Account{
						"Expense:Cranks Unlimited",
						big.NewRat(10.0, 1),
					},
					Account{
						"Expense:Cranks United",
						big.NewRat(10.0, 1),
					},
				},
			},
		},
		nil,
	},
	testCase{
		`1970-01-01 Payee
	Expense/test  123
	Assets
`,
		[]*Transaction{
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense/test",
						big.NewRat(123.0, 1),
					},
					Account{
						"Assets",
						big.NewRat(-123.0, 1),
					},
				},
			},
		},
		nil,
	},
}

func TestParseLedger(t *testing.T) {
	for i, tc := range testCases {
		b := bytes.NewBufferString(tc.data)
		transactions, err := ParseLedger(b, fmt.Sprintf("testCase[%d]", i))
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
