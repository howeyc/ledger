package pdr

import (
	"testing"
	"time"
)

// 2019-11-25
var baseTime = time.Unix(1574687238, 0).UTC()

var testCases = []struct {
	Input      string
	Start, End string
}{
	{"current month", "2019-11-01", "2019-12-01"},
	{"month to date", "2019-11-01", "2019-12-01"},
	{"last month", "2019-10-01", "2019-11-01"},
	{"previous month", "2019-10-01", "2019-11-01"},
	{"next month", "2019-12-01", "2020-01-01"},

	{"current year", "2019-01-01", "2020-01-01"},
	{"year to date", "2019-01-01", "2020-01-01"},
	{"ytd", "2019-01-01", "2020-01-01"},
	{"previous year", "2018-01-01", "2019-01-01"},
	{"last 3 years", "2017-01-01", "2020-01-01"},
	{"next year", "2020-01-01", "2021-01-01"},
	{"next two years", "2019-01-01", "2021-01-01"},
	{"next 5 years", "2019-01-01", "2024-01-01"},
	{"next three months", "2019-11-01", "2020-02-01"},

	{"current quarter", "2019-10-01", "2020-01-01"},
	{"next quarter", "2020-01-01", "2020-04-01"},
	{"next two quarters", "2019-10-01", "2020-04-01"},

	{"last week", "2019-11-17", "2019-11-24"},
	{"last 2 weeks", "2019-11-17", "2019-12-01"},
	{"next 4 weeks", "2019-11-24", "2019-12-22"},

	{"last 2 months", "2019-10-01", "2019-12-01"},
	{"last quarter", "2019-07-01", "2019-10-01"},
	{"last two quarters", "2019-07-01", "2020-01-01"},
	{"last three quarters", "2019-04-01", "2020-01-01"},

	// Adding max duration to baseTime
	{"all time", "0001-01-01", "2312-03-06"},
	{"forever", "0001-01-01", "2312-03-06"},
}

func TestParse(t *testing.T) {
	for _, c := range testCases {
		s, e, err := ParseRange(c.Input, baseTime)
		gotStart := s.Format(time.DateOnly)
		gotEnd := e.Format(time.DateOnly)
		if gotStart != c.Start {
			t.Fatalf("input %v, expected start: %v, got: %v", c.Input, c.Start, gotStart)
		}
		if gotEnd != c.End {
			t.Fatalf("input %v, expected end: %v, got: %v", c.Input, c.End, gotEnd)
		}
		if err != nil {
			t.Fatalf("input %v, unexpected error: %v", c.Input, err)
		}
	}
}
