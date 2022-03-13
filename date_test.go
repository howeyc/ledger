package ledger

import (
	"testing"
	"time"
)

type boundCase struct {
	period     Period
	start, end time.Time
	bounds     []time.Time
}

var boundCases = []boundCase{
	{
		PeriodYear,
		time.Date(2019, time.March, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2021, time.March, 23, 0, 0, 0, 0, time.UTC),
		[]time.Time{
			time.Date(2019, time.January, 01, 0, 0, 0, 0, time.UTC),
			time.Date(2020, time.January, 01, 0, 0, 0, 0, time.UTC),
			time.Date(2021, time.January, 01, 0, 0, 0, 0, time.UTC),
			time.Date(2022, time.January, 01, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		PeriodMonth,
		time.Date(2019, time.March, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2019, time.April, 23, 0, 0, 0, 0, time.UTC),
		[]time.Time{
			time.Date(2019, time.March, 01, 0, 0, 0, 0, time.UTC),
			time.Date(2019, time.April, 01, 0, 0, 0, 0, time.UTC),
			time.Date(2019, time.May, 01, 0, 0, 0, 0, time.UTC),
		},
	},
	{
		Period("Unknown"),
		time.Date(2019, time.March, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2019, time.April, 23, 0, 0, 0, 0, time.UTC),
		[]time.Time{
			time.Date(2019, time.March, 23, 0, 0, 0, 0, time.UTC),
			time.Date(2019, time.April, 23, 0, 0, 0, 0, time.UTC),
		},
	},
}

func TestDateBoundaries(t *testing.T) {
	for _, tc := range boundCases {
		bounds := getDateBoundaries(tc.period, tc.start, tc.end)
		if len(bounds) != len(tc.bounds) {
			t.Fatalf("Error(%s): expected `%d` bounds, got `%d` bounds", tc.period, len(tc.bounds), len(bounds))
		}
		for i, b := range bounds {
			if !b.Equal(tc.bounds[i]) {
				t.Errorf("Error(%s): expected [%d] = `%s` , got `%s`", tc.period, i, tc.bounds[i].Format(time.RFC3339), b.Format(time.RFC3339))
			}
		}
	}
}
