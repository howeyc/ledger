package ledger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/howeyc/ledger/internal/decimal"
)

type testBalCase struct {
	name     string
	data     string
	balances []Account
	err      error
}

var testBalCases = []testBalCase{
	{
		"simple case",
		`1970/01/01 Payee
	Expense/test  (123 * 3)
	Assets

1970/01/01 Payee
	Expense/test   123
	Assets
`,
		[]Account{
			{
				Name:    "Assets",
				Balance: decimal.NewFromFloat(-4 * 123),
			},
			{
				Name:    "Expense/test",
				Balance: decimal.NewFromFloat(123 + 369),
			},
		},
		nil,
	},
}

func TestBalanceLedger(t *testing.T) {
	for _, tc := range testBalCases {
		b := bytes.NewBufferString(tc.data)
		transactions, err := ParseLedger(b)
		bals := GetBalances(transactions, []string{})
		if (err != nil && tc.err == nil) || (err != nil && tc.err != nil && err.Error() != tc.err.Error()) {
			t.Errorf("Error: expected `%s`, got `%s`", tc.err, err)
		}
		exp, _ := json.Marshal(tc.balances)
		got, _ := json.Marshal(bals)
		if string(exp) != string(got) {
			t.Errorf("Error(%s): expected \n`%s`, \ngot \n`%s`", tc.name, exp, got)
		}
	}
}
