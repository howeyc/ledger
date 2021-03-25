package ledger

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"testing"
	"time"
)

type testCase struct {
	name         string
	data         string
	transactions []*Transaction
	err          error
}

var testCases = []testCase{
	testCase{
		"simple",
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
		"unbalanced error",
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Assets      123
`,
		nil,
		errors.New(":3: Unable to parse transaction: Unable to balance transaction: no empty account to place extra balance"),
	},
	testCase{
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
	testCase{
		"multiple empty lines",
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Assets



1970/01/01 Payee
	Expense/test   123
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
	testCase{
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
				Comments: []string{
					"; Handle tabs between account and amount",
					"; Also handle accounts with spaces",
				},
			},
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC().AddDate(0, 0, 1),
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
		},
		nil,
	},
	testCase{
		"accounts with slashes",
		`1970-01-01 Payee
    Expense/another     5
	Expense/test
	Assets      -128
`,
		[]*Transaction{
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense/another",
						big.NewRat(5.0, 1),
					},
					Account{
						"Expense/test",
						big.NewRat(123.0, 1),
					},
					Account{
						"Assets",
						big.NewRat(-128.0, 1),
					},
				},
			},
		},
		nil,
	},
	testCase{
		"comment inside transaction",
		`1970-01-01 Payee
	Expense/test  123
	; Expense/test  123
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
				Comments: []string{
					"; Expense/test  123",
				},
			},
		},
		nil,
	},
	testCase{
		"multiple comments",
		`; comment
	1970/01/01 Payee
	Expense/test   58
	Assets         -58           ; comment in trans
	Expense/unbalanced
`,
		[]*Transaction{
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense/test",
						big.NewRat(58, 1),
					},
					Account{
						"Assets",
						big.NewRat(-58, 1),
					},
					Account{
						"Expense/unbalanced",
						big.NewRat(0, 1),
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
	testCase{
		"header comment",
		`; comment
	1970/01/01 Payee
	Expense/test   58
	Assets         -58
	Expense/test   158
	Assets         -158
`,
		[]*Transaction{
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense/test",
						big.NewRat(58, 1),
					},
					Account{
						"Assets",
						big.NewRat(-58, 1),
					},
					Account{
						"Expense/test",
						big.NewRat(158, 1),
					},
					Account{
						"Assets",
						big.NewRat(-158, 1),
					},
				},
				Comments: []string{
					"; comment",
				},
			},
		},
		nil,
	},
	testCase{
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
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense/test",
						big.NewRat(246.0, 1),
					},
					Account{
						"Assets",
						big.NewRat(-246.0, 1),
					},
				},
			},
		},
		nil,
	},
	testCase{
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
			&Transaction{
				Payee: "Payee",
				Date:  time.Unix(0, 0).UTC(),
				AccountChanges: []Account{
					Account{
						"Expense/test",
						big.NewRat(246.0, 1),
					},
					Account{
						"Assets",
						big.NewRat(-246.0, 1),
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
	tc := testCase{
		"benchmark",
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
	}

	data := bytes.NewBufferString(tc.data)
	for n := 0; n < b.N; n++ {
		_, _ = ParseLedger(data)
	}
}
