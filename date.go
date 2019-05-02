package ledger

import "time"

// TransactionsInDateRange returns a new array of transactions that are in the date range
// specified by start and end. The returned list contains transactions on the same day as start
// but does not include any transactions on the day of end.
func TransactionsInDateRange(trans []*Transaction, start, end time.Time) []*Transaction {
	var newlist []*Transaction

	start = start.Add(-1 * time.Second)

	for _, tran := range trans {
		if tran.Date.After(start) && tran.Date.Before(end) {
			newlist = append(newlist, tran)
		}
	}

	return newlist
}

// Period is used to specify the length of a date range or frequency
type Period string

// Periods suppored by ledger
const (
	PeriodWeek     Period = "Weekly"
	Period2Week    Period = "BiWeekly"
	PeriodMonth    Period = "Monthly"
	Period2Month   Period = "BiMonthly"
	PeriodQuarter  Period = "Quarterly"
	PeriodSemiYear Period = "SemiYearly"
	PeriodYear     Period = "Yearly"
)

func getDateBoundaries(per Period, start, end time.Time) []time.Time {
	var incDays, incMonth, incYear int
	var periodStart time.Time

	switch per {
	case PeriodWeek:
		incDays = 7
		for periodStart = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC); periodStart.Weekday() != time.Sunday; {
			periodStart = periodStart.AddDate(0, 0, -1)
		}
	case Period2Week:
		incDays = 14
		for periodStart = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC); periodStart.Weekday() != time.Sunday; {
			periodStart = periodStart.AddDate(0, 0, -1)
		}
	case PeriodMonth:
		incMonth = 1
		periodStart = time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	case Period2Month:
		incMonth = 2
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
	case PeriodSemiYear:
		incMonth = 6
		switch start.Month() {
		case time.January, time.February, time.March, time.April, time.May, time.June:
			periodStart = time.Date(start.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
		default:
			periodStart = time.Date(start.Year(), time.July, 1, 0, 0, 0, 0, time.UTC)
		}
	case PeriodYear:
		incYear = 1
		periodStart = time.Date(start.Year(), time.January, 1, 0, 0, 0, 0, time.UTC)
	}

	boundaries := []time.Time{periodStart}
	for periodStart.Before(end) || periodStart.Equal(end) {
		periodStart = periodStart.AddDate(incYear, incMonth, incDays)
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

// RangeTransactions contains the transactions and the start and end time of the date range
type RangeTransactions struct {
	Start, End   time.Time
	Transactions []*Transaction
}

// startEndTime will return the start and end Times of a list of transactions
func startEndTime(trans []*Transaction) (start, end time.Time) {
	if len(trans) < 1 {
		return
	}

	start = trans[0].Date
	end = trans[0].Date

	for _, t := range trans {
		if end.Before(t.Date) {
			end = t.Date
		}
		if start.After(t.Date) {
			start = t.Date
		}
	}

	return
}

// TransactionsByPeriod will return the transactions for each period.
func TransactionsByPeriod(trans []*Transaction, per Period) []*RangeTransactions {
	var results []*RangeTransactions
	if len(trans) < 1 {
		return results
	}

	tStart, tEnd := startEndTime(trans)

	boundaries := getDateBoundaries(per, tStart, tEnd)

	bStart := boundaries[0]
	for _, boundary := range boundaries[1:] {
		bEnd := boundary

		bTrans := TransactionsInDateRange(trans, bStart, bEnd)
		// End date should be the last day (inclusive, so subtract 1 day)
		results = append(results, &RangeTransactions{Start: bStart, End: bEnd.AddDate(0, 0, -1), Transactions: bTrans})

		bStart = bEnd
	}

	return results
}

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

	tStart, tEnd := startEndTime(trans)

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
