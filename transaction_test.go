package ledger

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestIsBalanced(t *testing.T) {
	tests := []struct {
		name         string
		tx           *Transaction
		wantErr      error
		wantBalances []decimal.Decimal
	}{
		{
			name: "errors on too few postings",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:    "Assets:Bank",
						Balance: decimal.NewFromInt(10),
					},
				},
			},
			wantErr:      ErrNeedAtLeastTwoPostings,
			wantBalances: nil,
		},
		{
			name: "no empty account error",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:    "Assets:Bank",
						Balance: decimal.NewFromInt(10),
					},
					{
						Name:    "Expenses:Food",
						Balance: decimal.NewFromInt(-5),
					},
				},
			},
			wantErr:      ErrNoEmptyAccountForExtraBalance,
			wantBalances: nil,
		},
		{
			name: "more than one empty account error",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:    "Assets:Bank",
						Balance: decimal.NewFromInt(10),
					},
					{
						Name: "Expenses:Food",
					},
					{
						Name: "Equity:OpeningBalances",
					},
				},
			},
			wantErr:      ErrMoreThanOneEmptyAccountInTx,
			wantBalances: nil,
		},
		{
			name: "single empty account gets balancing amount",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:    "Assets:Bank",
						Balance: decimal.NewFromInt(-10),
					},
					{
						Name: "Expenses:Food",
					},
				},
			},
			wantErr:      nil,
			wantBalances: []decimal.Decimal{decimal.NewFromInt(-10), decimal.NewFromInt(10)},
		},
		{
			name: "already balanced with no empty account",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:    "Assets:Bank",
						Balance: decimal.NewFromInt(-10),
					},
					{
						Name:    "Expenses:Food",
						Balance: decimal.NewFromInt(10),
					},
				},
			},
			wantErr:      nil,
			wantBalances: []decimal.Decimal{decimal.NewFromInt(-10), decimal.NewFromInt(10)},
		},
		{
			name: "two currency implicit conversion factor inferred",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:     "Assets:Bank:USD",
						Currency: "USD",
						Balance:  decimal.NewFromInt(-10),
					},
					{
						Name:     "Assets:Bank:EUR",
						Currency: "EUR",
						Balance:  decimal.NewFromInt(5),
					},
				},
			},
			wantErr:      nil,
			wantBalances: nil,
		},
		{
			name: "two currency implicit conversion factor inferred multiple",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:     "Assets:Bank:USD",
						Currency: "USD",
						Balance:  decimal.NewFromInt(-10),
					},
					{
						Name:     "Assets:Bank:EUR",
						Currency: "EUR",
						Balance:  decimal.NewFromInt(5),
					},
					{
						Name:     "Assets:otherBank:EUR",
						Currency: "EUR",
						Balance:  decimal.NewFromInt(3),
					},
				},
			},
			wantErr:      nil,
			wantBalances: nil,
		},
		{
			name: "does not infer conversion factor for three currencies",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:     "Assets:Bank:USD",
						Currency: "USD",
						Balance:  decimal.NewFromInt(-10),
					},
					{
						Name:     "Assets:Bank:EUR",
						Currency: "EUR",
						Balance:  decimal.NewFromInt(5),
					},
					{
						Name:     "Assets:Bank:GBP",
						Currency: "GBP",
						Balance:  decimal.NewFromInt(3),
					},
				},
			},
			wantErr:      ErrNoEmptyAccountForExtraBalance,
			wantBalances: nil,
		},
		{
			name: "decimal precision bug",
			tx: &Transaction{
				AccountChanges: []Account{
					{
						Name:     "Assets:Wise:CZK",
						Currency: "CZK",
						Balance:  decimal.NewFromFloat(-2003.0),
					},
					{
						Name:     "Assets:Wise:EUR",
						Currency: "EUR",
						Balance:  decimal.NewFromFloat(1000.0),
					},
				},
			},
			wantErr:      nil,
			wantBalances: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.tx.IsBalanced()
			if err != tt.wantErr {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}

			if tt.wantBalances != nil {
				if len(tt.tx.AccountChanges) != len(tt.wantBalances) {
					t.Fatalf("expected %d account balances, got %d", len(tt.wantBalances), len(tt.tx.AccountChanges))
				}
				for i, want := range tt.wantBalances {
					if !tt.tx.AccountChanges[i].Balance.Equal(want) {
						t.Fatalf("account %d: expected balance %s, got %s", i, want.String(), tt.tx.AccountChanges[i].Balance.String())
					}
				}
			}
		})
	}
}
