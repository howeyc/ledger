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
}

func TestDecimal(t *testing.T) {
	for _, tc := range testCases {
		if tc.Result != tc.Input {
			t.Errorf("Error(%s): expected \n`%s`, \ngot \n`%s`", tc.name, tc.Result, tc.Input)
		}
	}
}
