package ledger

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/howeyc/ledger/decimal"
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

func TestBalancesByPeriod(t *testing.T) {
	b := bytes.NewBufferString(`
2022/02/02 Payee
	Assets     50
	Income

2022/01/02 Payee
	Assets     50
	Income

2022/03/02 Payee
	Assets     50
	Income

2022/04/02 Payee
	Assets     50
	Income

2022/05/02 Payee
	Assets     50
	Income

`)

	trans, _ := ParseLedger(b)
	partitionRb := BalancesByPeriod(trans, PeriodQuarter, RangePartition)
	snapshotRb := BalancesByPeriod(trans, PeriodQuarter, RangeSnapshot)

	if partitionRb[len(partitionRb)-1].Balances[0].Balance.Abs().Cmp(decimal.NewFromInt(100)) != 0 {
		t.Error("range balance by partition not accurate")
	}
	if snapshotRb[len(snapshotRb)-1].Balances[0].Balance.Abs().Cmp(decimal.NewFromInt(250)) != 0 {
		t.Error("range balance by snapshot not accurate")
	}

	transPeriod := TransactionsByPeriod(trans, PeriodQuarter)
	lastBals := GetBalances(transPeriod[len(transPeriod)-1].Transactions, []string{})
	if partitionRb[len(partitionRb)-1].Balances[0].Balance.Abs().Cmp(lastBals[0].Balance.Abs()) != 0 {
		t.Error("range balance by partition not equal to trans by period balance")
	}

	var blanktrans []*Transaction
	rb := BalancesByPeriod(blanktrans, PeriodDay, RangeSnapshot)
	if len(rb) > 1 {
		t.Error("range balances for non-existent transactions")
	}
}
