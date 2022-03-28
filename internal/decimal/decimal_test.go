package decimal

import (
	"testing"
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
}

func TestDecimal(t *testing.T) {
	for _, tc := range testCases {
		if tc.Result != tc.Input {
			t.Errorf("Error(%s): expected \n`%s`, \ngot \n`%s`", tc.name, tc.Result, tc.Input)
		}
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
}

func TestStringParse(t *testing.T) {
	for _, tc := range testParseCases {
		d, err := NewFromString(tc.Input)
		if err != nil {
			t.Fatal(err)
		}
		if tc.Result != d.StringFixedBank() {
			t.Errorf("Error(%s): expected \n`%s`, \ngot \n`%s`", tc.name, tc.Result, d.StringFixedBank())
		}
	}
}

func BenchmarkNewFromString(b *testing.B) {
	numbers := []string{"10.0", "245.6", "3", "2.456"}
	for n := 0; n < b.N; n++ {
		for _, numStr := range numbers {
			NewFromString(numStr)
		}
	}
}
