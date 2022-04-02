package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/howeyc/ledger/internal/decimal"
)

type testCase struct {
	name         string
	data         string
	transactions []*Transaction
	err          error
}

var testCases = []testCase{
	{
		"simple",
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Assets
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(369.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-369.0),
					},
				},
			},
		},
		nil,
	},
	{
		"unbalanced error",
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Assets      123
`,
		nil,
		errors.New(":3: Unable to parse transaction: Unable to balance transaction: no empty account to place extra balance"),
	},
	{
		"single posting",
		`1970/01/01 Payee
	Assets:Account    5`,
		nil,
		errors.New(":2: Unable to parse transaction: Unable to balance transaction: need at least two postings"),
	},
	{
		"no posting",
		`1970/01/01 Payee
`,
		nil,
		errors.New(":1: Unable to parse transaction: Unable to balance transaction: need at least two postings"),
	},
	{
		"multiple empty",
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Wallet
	Assets      123
	Bank
`,
		nil,
		errors.New(":5: Unable to parse transaction: Unable to balance transaction: more than one account empty"),
	},
	{
		"multiple empty lines",
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Assets



1970/01/01 Payee
	Expense/test   123
	Assets
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(369.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-369.0),
					},
				},
			},
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(123.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-123.0),
					},
				},
			},
		},
		nil,
	},
	{
		"accounts with spaces",
		`1970/01/02 Payee
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
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC().AddDate(0, 0, 1),
				AccountChanges: []Account{
					{
						"Expense:test",
						decimal.NewFromFloat(369.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-369.0),
					},
				},
			},
			{
				Payee: "Payee 5",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense:Cars R Us",
						decimal.NewFromFloat(-388.0),
					},
					{
						"Expense:Cars",
						decimal.NewFromFloat(358.0),
					},
					{
						"Expense:Cranks",
						decimal.NewFromFloat(10.0),
					},
					{
						"Expense:Cranks Unlimited",
						decimal.NewFromFloat(10.0),
					},
					{
						"Expense:Cranks United",
						decimal.NewFromFloat(10.0),
					},
				},
				Comments: []string{
					"; Handle tabs between account and amount",
					"; Also handle accounts with spaces",
				},
			},
		},
		nil,
	},
	{
		"accounts with slashes",
		`1970-01-01 Payee
    Expense/another     5
	Expense/test
	Assets      -128
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/another",
						decimal.NewFromFloat(5.0),
					},
					{
						"Expense/test",
						decimal.NewFromFloat(123.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-128.0),
					},
				},
			},
		},
		nil,
	},
	{
		"comment after payee",
		`1970-01-01 Payee      ; payee comment
	Expense/test  123
	Assets
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(123.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-123.0),
					},
				},
				Comments: []string{
					"; payee comment",
				},
			},
		},
		nil,
	},
	{
		"comment inside transaction",
		`1970-01-01 Payee
	Expense/test  123
	; Expense/test  123
	Assets
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(123.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-123.0),
					},
				},
				Comments: []string{
					"; Expense/test  123",
				},
			},
		},
		nil,
	},
	{
		"multiple comments",
		`; comment
	1970/01/01 Payee
	Expense/test   58
	Assets         -58           ; comment in trans
	Expense/unbalanced
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(58),
					},
					{
						"Assets",
						decimal.NewFromFloat(-58),
					},
					{
						"Expense/unbalanced",
						decimal.NewFromFloat(0),
					},
				},
				Comments: []string{
					"; comment",
					"; comment in trans",
				},
			},
		},
		nil,
	},
	{
		"header comment",
		`; comment
	1970/01/01 Payee
	Expense/test   58
	Assets         -58
	Expense/test   158
	Assets         -158
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(58),
					},
					{
						"Assets",
						decimal.NewFromFloat(-58),
					},
					{
						"Expense/test",
						decimal.NewFromFloat(158),
					},
					{
						"Assets",
						decimal.NewFromFloat(-158),
					},
				},
				Comments: []string{
					"; comment",
				},
			},
		},
		nil,
	},
	{
		"account skip",
		`1970/01/01 Payee
	Expense/test  123
	Assets

account Expense/test

account Assets
	note bambam
	payee junkjunk

1970/01/01 Payee
	Expense/test  (123 * 2)
	Assets
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(123.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-123.0),
					},
				},
			},
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(246.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-246.0),
					},
				},
			},
		},
		nil,
	},
	{
		"multiple account skip",
		`1970/01/01 Payee
	Expense/test  123
	Assets

account Banking
account Expense/test
account Assets

1970/01/01 Payee
	Expense/test  (123 * 2)
	Assets
`,
		[]*Transaction{
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(123.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-123.0),
					},
				},
			},
			{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					{
						"Expense/test",
						decimal.NewFromFloat(246.0),
					},
					{
						"Assets",
						decimal.NewFromFloat(-246.0),
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
		if (err != nil && tc.err == nil) || (err != nil && tc.err != nil && err.Error() != tc.err.Error()) {
			t.Errorf("Error: expected `%s`, got `%s`", tc.err, err)
		}
		exp, _ := json.Marshal(tc.transactions)
		got, _ := json.Marshal(transactions)
		if string(exp) != string(got) {
			t.Errorf("Error(%s): expected \n`%s`, \ngot \n`%s`", tc.name, exp, got)
		}
	}
}

func BenchmarkParseLedger(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_, _ = ParseLedgerFile("testdata/ledgerBench.dat")
	}
}
