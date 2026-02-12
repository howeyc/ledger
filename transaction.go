package ledger

import (
	"errors"

	"github.com/shopspring/decimal"
)

var (
	ErrNeedAtLeastTwoPostings        = errors.New("need at least two postings")
	ErrNoEmptyAccountForExtraBalance = errors.New("unable to balance transaction: no empty account to place extra balance")
	ErrMoreThanOneEmptyAccountInTx   = errors.New("unable to balance transaction: more than one account empty")
)

func (t *Transaction) IsBalanced() error {
	if len(t.AccountChanges) < 2 {
		return ErrNeedAtLeastTwoPostings
	}

	if err := t.inferConversionFactorForTwoCurrencyTx(); err != nil {
		return err
	}

	transBal := decimal.Zero
	var numEmpty int
	var emptyAccIndex int

	for i, acc := range t.AccountChanges {
		if acc.Balance.IsZero() {
			numEmpty++
			emptyAccIndex = i
		}

		if acc.Converted != nil {
			transBal = transBal.Add(acc.Converted.Neg())
		} else if acc.ConversionFactor != nil {
			transBal = transBal.Add(acc.Balance.Mul(*acc.ConversionFactor))
		} else {
			transBal = transBal.Add(acc.Balance)
		}
	}

	if !transBal.IsZero() {
		switch numEmpty {
		case 0:
			return ErrNoEmptyAccountForExtraBalance
		case 1:
			// If there is a single empty account, then it is obvious where to
			// place the remaining balance.
			t.AccountChanges[emptyAccIndex].Balance = transBal.Neg()
		default:
			return ErrMoreThanOneEmptyAccountInTx
		}
	}

	return nil
}

func (t *Transaction) inferConversionFactorForTwoCurrencyTx() error {
	type currencyGroup struct {
		indices []int
	}

	currencyMap := make(map[string]*currencyGroup)

	getCurrencyKey := func(a *Account) string {
		if a.Converted != nil {
			// TODO: explicit currency for conversion?
			return a.Currency
		}
		return a.Currency
	}

	for i := range t.AccountChanges {
		acc := &t.AccountChanges[i]
		key := getCurrencyKey(acc)
		if key == "" {
			return nil
		}
		group, ok := currencyMap[key]
		if !ok {
			group = &currencyGroup{}
			currencyMap[key] = group
		}
		group.indices = append(group.indices, i)
	}

	if len(currencyMap) != 2 {
		return nil
	}

	var (
		curKeys [2]string
		groups  [2]*currencyGroup
		i       int
	)
	for k, g := range currencyMap {
		if i >= 2 {
			break
		}
		curKeys[i] = k
		groups[i] = g
		i++
	}

	var baseCurIdx, otherCurIdx int
	hasConv0 := false
	for _, idx := range groups[0].indices {
		if t.AccountChanges[idx].ConversionFactor != nil {
			hasConv0 = true
			break
		}
	}
	hasConv1 := false
	for _, idx := range groups[1].indices {
		if t.AccountChanges[idx].ConversionFactor != nil {
			hasConv1 = true
			break
		}
	}

	switch {
	case hasConv0 && hasConv1:
		return nil
	case hasConv0:
		baseCurIdx, otherCurIdx = 1, 0
	case hasConv1:
		baseCurIdx, otherCurIdx = 0, 1
	default:
		baseCurIdx, otherCurIdx = 0, 1
	}

	sumForGroup := func(g *currencyGroup) (decimal.Decimal, error) {
		total := decimal.Zero
		for _, idx := range g.indices {
			acc := &t.AccountChanges[idx]
			if acc.Converted != nil {
				total = total.Add(acc.Converted.Neg())
			} else if acc.ConversionFactor != nil {
				total = total.Add(acc.Balance.Mul(*acc.ConversionFactor))
			} else {
				total = total.Add(acc.Balance)
			}
		}
		return total, nil
	}

	sumBase, _ := sumForGroup(groups[baseCurIdx])
	sumOtherRaw := decimal.Zero
	for _, idx := range groups[otherCurIdx].indices {
		acc := &t.AccountChanges[idx]
		if acc.Converted != nil || acc.ConversionFactor != nil {
			if acc.Converted != nil {
				sumOtherRaw = sumOtherRaw.Add(acc.Converted.Neg())
			} else if acc.ConversionFactor != nil {
				sumOtherRaw = sumOtherRaw.Add(acc.Balance.Mul(*acc.ConversionFactor))
			}
		} else {
			sumOtherRaw = sumOtherRaw.Add(acc.Balance)
		}
	}

	if sumOtherRaw.IsZero() {
		return nil
	}
	if sumBase.Add(sumOtherRaw).IsZero() {
		return nil
	}

	for _, idx := range groups[otherCurIdx].indices {
		acc := &t.AccountChanges[idx]
		if acc.ConversionFactor == nil && acc.Converted == nil {
			conv := acc.Balance.Mul(sumBase).Div(sumOtherRaw)
			acc.Converted = &conv
		}
	}

	return nil
}
