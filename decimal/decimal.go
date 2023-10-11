// Package decimal implements fixed-point decimal with accuracy to 3 digits of
// precision after the decimal point.
//
// int64 is the underlying data type for speed of computation. However, using
// an int64 cast to Decimal will not work, one of the "New" functions must
// be used to get accurate results.
//
// The package multiplies every source value by 1000, and then does integer
// math from that point forward, maintaining all values at that scale over
// every operation.
//
// Note: For use in ledger. Cannot handle values over approx 900 trillion.
package decimal

import (
	"errors"
	"strconv"
	"strings"
)

// Decimal represents a fixed-point decimal.
type Decimal int64

// scaleFactor used for math operations,
const scaleFactor = 1000

// precision of 3 digits
const precision = 3

// Zero constant, to make initializations easier.
const Zero = Decimal(0)

// One constant, to make initializations easier.
const One = Decimal(scaleFactor)

// Parse max/min for whole number part
const parseMax = (1<<63 - 1) / scaleFactor
const parseMin = (-1 << 63) / scaleFactor

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

// atoi64 is equivalent to strconv.Atoi
func atoi64(s string) (bool, int64, error) {
	sLen := len(s)
	if sLen < 1 || sLen > 18 {
		return false, 0, errors.New("atoi failed")
	}
	neg := false
	if s[0] == '-' {
		neg = true
		s = s[1:]
		if len(s) < 1 {
			return false, 0, errors.New("atoi failed")
		}
	}

	var n int64
	for _, ch := range []byte(s) {
		ch -= '0'
		if ch > 9 {
			return false, 0, errors.New("atoi failed")
		}
		n = n*10 + int64(ch)
	}
	if neg {
		n = -n
	}
	return neg, n, nil
}

// NewFromString returns a Decimal from a string representation. Throws an
// error if integer parsing fails.
func NewFromString(s string) (Decimal, error) {
	if whole, frac, split := strings.Cut(s, "."); split {
		neg, w, err := atoi64(whole)
		if err != nil {
			return Zero, err
		}

		// overflow
		if w > parseMax || w < parseMin {
			return Zero, errors.New("number too big")
		}
		w = w * int64(scaleFactor)

		// Parse up to *precision* digits and scale up
		var f int64
		var seen int
		for _, b := range frac {
			f *= 10
			if b < '0' || b > '9' {
				return Zero, errors.New("invalid syntax")
			}
			f += int64(b - '0')
			seen++
			if seen == precision {
				break
			}
		}
		for seen < precision {
			f *= 10
			seen++
		}

		if neg {
			f = -f
		}
		return Decimal(w + f), err
	} else {
		_, i, err := atoi64(s)
		if i > parseMax || i < parseMin {
			return Zero, errors.New("number too big")
		}
		i = i * int64(scaleFactor)
		return Decimal(i), err
	}
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
//
//	0 if d == 0
//
// +1 if d >  0
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
	return (d * scaleFactor) / d1
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
//	-1 if d <  d1
//	 0 if d == d1
//	+1 if d >  d1
func (d Decimal) Cmp(d1 Decimal) int {
	if d < d1 {
		return -1
	} else if d > d1 {
		return 1
	}
	return 0
}

// fmtFrac formats the fraction of v/10**prec (e.g., ".12345") into the
// tail of buf. It returns the index where the
// output bytes begin and the value v/10**prec.
func fmtFrac(buf []byte, v uint64, prec int) (nw int, nv uint64) {
	w := len(buf)
	for i := 0; i < prec; i++ {
		digit := v % 10
		w--
		buf[w] = byte(digit) + '0'
		v /= 10
	}
	w--
	buf[w] = '.'
	return w, v
}

// fmtInt formats v into the tail of buf.
// It returns the index where the output begins.
func fmtInt(buf []byte, v uint64) int {
	w := len(buf)
	if v == 0 {
		w--
		buf[w] = '0'
	} else {
		for v > 0 {
			w--
			buf[w] = byte(v%10) + '0'
			v /= 10
		}
	}
	return w
}

// StringFixedBank returns a banker rounded fixed-point string with 2 digits
// after the decimal point.
//
// Example:
//
// NewFromFloat(5.455).StringFixedBank() == "5.46"
// NewFromFloat(5.445).StringFixedBank() == "5.44"
func (d Decimal) StringFixedBank() string {
	var buf [24]byte
	w := len(buf)

	u := uint64(d)
	neg := d < 0
	if neg {
		u = -u
	}

	// Bank rounding
	rem := u % 10
	u /= 10
	if rem > 5 || (rem == 5 && u%2 != 0) {
		u++
	}

	// fmt functions from time.Duration
	w, u = fmtFrac(buf[:w], u, precision-1)
	w = fmtInt(buf[:w], u)

	if neg {
		w--
		buf[w] = '-'
	}

	return string(buf[w:])
}

// StringTruncate returns the whole-number (Int) part of d.
//
// Example:
//
// NewFromFloat(5.44).StringTruncate() == "5"
func (d Decimal) StringTruncate() string {
	whole := d / scaleFactor
	return strconv.FormatInt(int64(whole), 10)
}

// StringRound returns the nearest rounded whole-number (Int) part of d.
// Example:
//
// NewFromFloat(5.5).StringRound() == "6"
// NewFromFloat(5.4).StringRound() == "5"
// NewFromFloat(-5.4).StringRound() == "5"
// NewFromFloat(-5.5).StringRound() == "6"
func (d Decimal) StringRound() string {
	whole := d / scaleFactor
	frac := (d % scaleFactor)
	neg := false
	if frac < 0 {
		frac = -frac
		neg = true
	}
	if frac >= (5 * (scaleFactor / 10)) {
		if neg {
			whole--
		} else {
			whole++
		}
	}
	return strconv.FormatInt(int64(whole), 10)
}
