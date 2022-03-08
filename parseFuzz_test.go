//go:build go1.18

package ledger

import (
	"bytes"
	"math/big"
	"testing"
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
		overall := new(big.Rat)
		for _, t := range trans {
			for _, p := range t.AccountChanges {
				overall.Add(overall, p.Balance)
			}
		}
		if overall.Cmp(new(big.Rat)) != 0 {
			t.Error("Bad balance")
		}
	})
}
