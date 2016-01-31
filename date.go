package ledger

import "time"

// TransactionsInDateRange returns a new array of transactions that are in the date range
// specified by start and end. The returned list contains transactions on the same day as start
// but does not include any transactions on the day of end.
func TransactionsInDateRange(trans []*Transaction, start, end time.Time) []*Transaction {
	var newlist []*Transaction

	start = start.Add(-1 * time.Second)

	for _, tran := range trans {
		if tran.Date.After(start) && tran.Date.Before((end)) {
			newlist = append(newlist, tran)
		}
	}

	return newlist
}

// Period is used to specify the length of a date range or frequency
type Period string

// Periods suppored by ledger
const (
	PeriodMonth   Period = "Monthly"
	PeriodQuarter Period = "Quarterly"
	PeriodYear    Period = "Yearly"
)

func getDateBoundaries(per Period, start, end time.Time) []time.Time {
	var incMonth, incYear int
	var periodStart time.Time

	switch per {
	case PeriodMonth:
		incMonth = 1
		periodStart = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	case PeriodQuarter:
		incMonth = 3
		switch start.Month() {
		case time.January, time.February, time.March:
			periodStart = time.Date(start.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		case time.April, time.May, time.June:
			periodStart = time.Date(start.Year(), time.April, 1, 0, 0, 0, 0, time.UTC)
		case time.July, time.August, time.September:
			periodStart = time.Date(start.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
		default:
			periodStart = time.Date(start.Year(), time.October, 1, 0, 0, 0, 0, time.UTC)
		}
	case PeriodYear:
		incYear = 1
		periodStart = time.Date(start.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	boundaries := []time.Time{periodStart}
	for periodStart.Before(end) {
		periodStart = periodStart.AddDate(incYear, incMonth, 0)
		boundaries = append(boundaries, periodStart)
	}

	return boundaries
}

// RangeType is used to specify how the data is "split" into sections
type RangeType string

const (
	// RangeSnapshot will have each section be the running total at the time of the snapshot
	RangeSnapshot RangeType = "Snapshot"

	// RangePartition will have each section be the accumulated value of the transactions within that partition's date range
	RangePartition RangeType = "Partition"
)

// RangeBalance contains the account balances and the start and end time of the date range
type RangeBalance struct {
	Start, End time.Time
	Balances   []*Account
}

// BalancesByPeriod will return the account balances for each period.
func BalancesByPeriod(trans []*Transaction, per Period, rType RangeType) []*RangeBalance {
	var results []*RangeBalance
	if len(trans) < 1 {
		return results
	}

	tStart := trans[0].Date
	tEnd := trans[len(trans)-1].Date

	boundaries := getDateBoundaries(per, tStart, tEnd)

	bStart := boundaries[0]
	for _, boundary := range boundaries[1:] {
		bEnd := boundary

		bTrans := TransactionsInDateRange(trans, bStart, bEnd)
		// End date should be the last day (inclusive, so subtract 1 day)
		results = append(results, &RangeBalance{Start: bStart, End: bEnd.AddDate(0, 0, -1), Balances: GetBalances(bTrans, []string{})})

		if rType == RangePartition {
			bStart = bEnd
		}
	}

	return results
}
