package decimal

import (
	"math/rand"
	"strings"
	"testing"

	sdec "github.com/shopspring/decimal"
)

type testCase struct {
	name          string
	Result, Input string
}

var testCases = []testCase{
	{
		"multiply",
		NewFromFloat(48.0).StringFixedBank(),
		NewFromInt(6).Mul(NewFromInt(8)).StringFixedBank(),
	},
	{
		"divide",
		NewFromFloat(6.0).StringFixedBank(),
		NewFromInt(48).Div(NewFromInt(8)).StringFixedBank(),
	},
	{
		"divide-1",
		NewFromFloat(11.111).StringFixedBank(),
		NewFromInt(100).Div(NewFromInt(9)).StringFixedBank(),
	},
	{
		"sum",
		NewFromFloat(234.56).StringFixedBank(),
		NewFromFloat(123.12).Add(NewFromInt(111)).Add(NewFromFloat(0.44)).StringFixedBank(),
	},
	{
		"bankrounduppos",
		NewFromFloat(234.56).StringFixedBank(),
		NewFromFloat(234.555).StringFixedBank(),
	},
	{
		"bankrounddownpos",
		NewFromFloat(234.54).StringFixedBank(),
		NewFromFloat(234.545).StringFixedBank(),
	},
	{
		"bankroundupneg",
		"-234.56",
		NewFromFloat(-234.555).StringFixedBank(),
	},
	{
		"bankrounddownneg",
		"-234.54",
		NewFromFloat(-234.545).StringFixedBank(),
	},
	{
		"rounduppos",
		NewFromFloat(234.56).StringFixedBank(),
		NewFromFloat(234.556).StringFixedBank(),
	},
	{
		"rounddownpos",
		NewFromFloat(234.55).StringFixedBank(),
		NewFromFloat(234.554).StringFixedBank(),
	},
	{
		"roundupneg",
		"-234.56",
		NewFromFloat(-234.556).StringFixedBank(),
	},
	{
		"rounddownneg",
		"-234.55",
		NewFromFloat(-234.554).StringFixedBank(),
	},
	{
		"truncate",
		NewFromInt(234).StringTruncate(),
		NewFromFloat(234.554).StringTruncate(),
	},
	{
		"2digits-1",
		"1.00",
		One.StringFixedBank(),
	},
	{
		"2digits-4.5",
		"4.50",
		NewFromFloat(4.5).StringFixedBank(),
	},
	{
		"roundintuppos",
		"6",
		NewFromFloat(5.6).StringRound(),
	},
	{
		"roundintdownpos",
		"5",
		NewFromFloat(5.4).StringRound(),
	},
	{
		"roundintupneg",
		"-5",
		NewFromFloat(-5.4).StringRound(),
	},
	{
		"roundintdownneg",
		"-6",
		NewFromFloat(-5.6).StringRound(),
	},
	{
		"negfrac",
		"-0.43",
		NewFromFloat(-0.43).StringFixedBank(),
	},
	{
		"sub",
		"5.12",
		NewFromFloat(5.56).Sub(NewFromFloat(0.44)).StringFixedBank(),
	},
	{
		"neg",
		"-5.12",
		NewFromFloat(5.12).Neg().StringFixedBank(),
	},
	{
		"abs-1",
		"5.12",
		NewFromFloat(-5.12).Abs().StringFixedBank(),
	},
	{
		"abs-1",
		"5.12",
		NewFromFloat(5.12).Abs().StringFixedBank(),
	},
}

func TestDecimal(t *testing.T) {
	for _, tc := range testCases {
		if tc.Result != tc.Input {
			t.Errorf("Error(%s): expected \n`%s`, \ngot \n`%s`", tc.name, tc.Result, tc.Input)
		}
	}
}

func TestFloat(t *testing.T) {
	d := NewFromFloat(5.56)
	f := float64(5.56)
	if df, _ := d.Float64(); df != f {
		t.Error("Float64 not exact")
	}
}

func TestCompare(t *testing.T) {
	l := NewFromInt(5)
	h := NewFromInt(10)
	z := NewFromInt(0)

	if !z.IsZero() {
		t.Error("zero failed")
	}

	if h.Cmp(l) != 1 || l.Cmp(h) != -1 || z.Cmp(Zero) != 0 {
		t.Error("compare fail")
	}
}

func TestSign(t *testing.T) {
	n := NewFromInt(-5)
	p := NewFromInt(5)
	z := NewFromInt(0)

	if z.Sign() != 0 {
		t.Error("zero failed")
	}

	if n.Sign() != -1 || p.Sign() != 1 {
		t.Error("sign fail")
	}
}

var testParseCases = []testCase{
	{
		"negzero",
		"-0.43",
		"-0.43",
	},
	{
		"poszero",
		"0.43",
		"0.43",
	},
	{
		"3digit",
		"5.56",
		"5.564",
	},
	{
		"truncateinput",
		"5.56",
		"5.56432342",
	},
	{
		"precise",
		"16.24",
		"16.24",
	},
	{
		"fuzz-1",
		"0.00",
		"0.0051",
	},
	{
		"fuzz-2",
		"8.00",
		"8.005",
	},
	{
		"fuzz-3",
		"0.00",
		"0.005",
	},
	{
		"fuzz-4",
		"1.00",
		"0.997",
	},
	{
		"fuzz-5",
		"2200000000000021.00",
		"2200000000000021",
	},
	{
		"fuzz-6",
		"0.01",
		"0.010e1",
	},
	{
		"fuzz-7",
		"-8.00",
		"-7.995",
	},
	{
		"fuzz-8",
		"-9.00",
		"-8.995",
	},
	{
		"fuzz-9",
		"8.00",
		"7.995",
	},
	{
		"fuzz-10",
		"9.00",
		"8.995",
	},
	{
		"fuzz-11",
		"-7.98",
		"-7.985",
	},
	{
		"fuzz-12",
		"-8.98",
		"-8.985",
	},
	{
		"fuzz-13",
		"7.98",
		"7.985",
	},
	{
		"fuzz-14",
		"8.98",
		"8.984",
	},
	{
		"fuzz-15",
		"-8.00",
		"-7.999",
	},
	{
		"fuzz-16",
		"-9.00",
		"-8.999",
	},
	{
		"fuzz-17",
		"8.00",
		"7.999",
	},
	{
		"fuzz-18",
		"9.00",
		"8.999",
	},
	{
		"error-1",
		errTooBig.Error(),
		"100000000000000000",
	},
	{
		"error-2",
		errTooBig.Error(),
		"10000000000000000",
	},
	{
		"error-3",
		errTooBig.Error(),
		"10000000000000000.56",
	},
	{
		"error-4",
		errInvalid.Error(),
		"0.e0",
	},
	{
		"error-5",
		errTooBig.Error(),
		"5555555555555555555555555550000000000000000",
	},
	{
		"error-6",
		errEmpty.Error(),
		"-",
	},
	{
		"error-7",
		errEmpty.Error(),
		"",
	},
	{
		"error-badint-1",
		errInvalid.Error(),
		"1QZ.56",
	},
	{
		"error-expr-1",
		errInvalid.Error(),
		"(123 * 6)",
	},
	{
		"missingwhole",
		"0.50",
		".50",
	},
	{
		"negmissingwhole",
		"-0.50",
		"-.50",
	},
	{
		"missingfrac",
		"5.00",
		"5.",
	},
	{
		"neg-missingfrac",
		"-5.00",
		"-5.",
	},
	{
		"just-a-decimal",
		"0.00",
		".",
	},
}

func TestStringParse(t *testing.T) {
	for _, tc := range testParseCases {
		d, err := NewFromString(tc.Input)
		if strings.HasPrefix(tc.name, "error") {
			if err == nil {
				t.Fatalf("Error(%s): expected error `%s`", tc.name, tc.Result)
			}
			if err.Error() != tc.Result {
				t.Fatalf("Error(%s): expected `%s`, got `%s`", tc.name, tc.Result, err)
			}
		}
		if !strings.HasPrefix(tc.name, "error") && err != nil {
			t.Fatalf("Error(%s): unexpected error `%s`", tc.name, err)
		}
		if !strings.HasPrefix(tc.name, "error") && tc.Result != d.StringFixedBank() {
			t.Errorf("Error(%s): expected \n`%s`, \ngot \n`%s`", tc.name, tc.Result, d.StringFixedBank())
		}
	}
}

func FuzzStringParse(f *testing.F) {
	f.Fuzz(func(t *testing.T, s string) {
		if _, after, split := strings.Cut(s, "."); split {
			if len(after) > 3 {
				return
			}
		}
		sd, serr := sdec.NewFromString(s)
		if serr != nil {
			return
		}
		d, err := NewFromString(s)
		if err != nil {
			return
		}
		ss := strings.TrimPrefix(sd.StringFixedBank(2), "-")
		ds := strings.TrimPrefix(d.StringFixedBank(), "-")

		if ds != ss {
			t.Fatalf("no match: decimal \n`%s`, \nsdec \n `%s`", ds, ss)
		}
	})
}

func BenchmarkNewFromString(b *testing.B) {
	numbers := []string{"10.0", "245.6", "354", "2.456", "-31.2"}
	for b.Loop() {
		for _, numStr := range numbers {
			NewFromString(numStr)
		}
	}
}

func BenchmarkStringFixedBank(b *testing.B) {
	var numbers [1000]Decimal
	for i := range len(numbers) {
		numbers[i] = NewFromFloat(rand.Float64() * 100000)
		if i%2 == 0 {
			numbers[i] *= -1
		}
	}
	for b.Loop() {
		for _, num := range numbers {
			num.StringFixedBank()
		}
	}
}
