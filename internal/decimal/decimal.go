// Package decimal implements fixed-point decimal with accuracy to 3 digits of
// precision after the decimal point.
//
// int64 is the underlying data type for speed of computation. However, using
// an int64 casted to Decimal will not work, one of the "New" functions must
// be used to get accurate results.
//
// The package multiplies every source value by 1000, and then does integer
// math from that point forward, maintaining all values at that scale over
// every operation.
//
// Note: For use in ledger. Cannot handle values over approx 900 trillion.
package decimal

import (
	"fmt"
	"strconv"
)

// Decimal represents a fixed-point decimal.
type Decimal int64

// scaleFactor used for math operations, 3 digit precision
const scaleFactor Decimal = 1000

// Zero constant, to make initializations easier.
const Zero = Decimal(0)

// One constant, to make initializations easier.
const One = scaleFactor

// NewFromFloat converts a float64 to Decimal. Only 3 digits of precision after
// the decimal point are preserved.
func NewFromFloat(f float64) Decimal {
	return Decimal(f * float64(scaleFactor))
}

// NewFromInt converts a int64 to Decimal. Multiplies by 1000 to get into
// Decimal scale.
func NewFromInt(i int64) Decimal {
	return Decimal(i) * scaleFactor
}

// NewFromString returns a Decimal from a string representation. Throws an
// erorr if parsing a float64 from string fails.
func NewFromString(s string) (Decimal, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return Zero, err
	}
	return NewFromFloat(f), nil
}

// IsZero returns true if d == 0
func (d Decimal) IsZero() bool {
	return d == Zero
}

// Neg returns -d
func (d Decimal) Neg() Decimal {
	return -d
}

// Sign returns:
//
// -1 if d <  0
//  0 if d == 0
// +1 if d >  0
//
func (d Decimal) Sign() int {
	if d < 0 {
		return -1
	} else if d > 0 {
		return 1
	}
	return 0
}

// Add returns d + d1
func (d Decimal) Add(d1 Decimal) Decimal {
	return d + d1
}

// Sub returns d - d1
func (d Decimal) Sub(d1 Decimal) Decimal {
	return d - d1
}

// Mul returns d * d1
func (d Decimal) Mul(d1 Decimal) Decimal {
	return (d * d1) / scaleFactor
}

// Div returns d / d1
func (d Decimal) Div(d1 Decimal) Decimal {
	return (d / d1) * scaleFactor
}

// Abs returns the absolute value of the decimal
func (d Decimal) Abs() Decimal {
	if d < 0 {
		return d.Neg()
	}
	return d
}

// Float64 returns the float64 value for d, and exact is always set to false.
// The signature is this way to match big.Rat
func (d Decimal) Float64() (f float64, exact bool) {
	return float64(d) / float64(scaleFactor), false
}

// Cmp compares the numbers represented by d and d1 and returns:
//
//     -1 if d <  d1
//      0 if d == d1
//     +1 if d >  d1
//
func (d Decimal) Cmp(d1 Decimal) int {
	if d < d1 {
		return -1
	} else if d > d1 {
		return 1
	}
	return 0
}

// StringFixedBank returns a banker rounded fixed-point string with 2 digits
// after the decimal point.
//
// Example:
//
// NewFromFloat(5.455).StringFixedBank() == "5.46"
// NewFromFloat(5.445).StringFixedBank() == "5.44"
//
func (d Decimal) StringFixedBank() string {
	whole := d / scaleFactor
	frac := (d % scaleFactor) / 10
	rem := d % 10

	if frac < 0 {
		frac = -frac
	}
	if rem < 0 {
		rem = -rem
	}

	if rem > 5 {
		frac++
	} else if rem == 5 && frac%2 != 0 {
		frac++
	}

	return fmt.Sprintf("%d.%02d", whole, frac)
}

// StringTruncate returns the whole-number (Int) part of d.
//
// Example:
//
// NewFromFloat(5.44).StringTruncate() == "5"
//
func (d Decimal) StringTruncate() string {
	whole := d / scaleFactor
	return fmt.Sprintf("%d", whole)
}

// StringRound returns the nearest rounded whole-number (Int) part of d.
// Example:
//
// NewFromFloat(5.5).StringRound() == "6"
// NewFromFloat(5.4).StringRound() == "5"
// NewFromFloat(-5.4).StringRound() == "5"
// NewFromFloat(-5.5).StringRound() == "6"
//
func (d Decimal) StringRound() string {
	whole := d / scaleFactor
	frac := (d % scaleFactor) / 100
	neg := false
	if frac < 0 {
		frac = -frac
		neg = true
	}
	if frac >= 5 {
		if neg {
			whole--
		} else {
			whole++
		}
	}
	return fmt.Sprintf("%d", whole)
}
