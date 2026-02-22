//go:build go1.18

package ledger

import (
	"bytes"
	"testing"

	"github.com/shopspring/decimal"
)

func FuzzParseLedger(f *testing.F) {
	for _, tc := range testCases {
		if tc.err == nil {
			f.Add(tc.data)
		}
	}
	f.Fuzz(func(t *testing.T, s string) {
		b := bytes.NewBufferString(s)
		trans, _ := ParseLedger(b)
		overall := decimal.Zero
		for _, t := range trans {
			for _, p := range t.AccountChanges {
				switch {
				case p.Converted != nil:
					overall = overall.Add(p.Converted.Neg())
				case p.ConversionFactor != nil:
					overall = overall.Add(p.Balance.Mul(
						*p.ConversionFactor,
					))
				default:
					overall = overall.Add(p.Balance)
				}
			}
		}
		if !overall.IsZero() {
			t.Error("Bad balance")
		}
	})
}
