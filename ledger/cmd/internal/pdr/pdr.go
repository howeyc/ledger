//go:generate peg -inline -switch grammer.peg

// Package pdr parses date range as string
// Uses pointlander/peg
package pdr

import (
	"strings"
	"time"
)

// ParseRange parses a human readable specified time range into two dates containing that range.
// start is included in the range, end is just beyond the range. So the returned dates/times are
// such that the range is start <= RANGE < end.
//
// Also, ranges with a numeric factor are returned as if the current date is a part of the range
// specified. For instance, the "last two months" is the previous month and the current month.
// However, range without numeric factor excludes current month. Specifying "last month" returns
// just the range for that month.
func ParseRange(s string, baseTime time.Time) (start, end time.Time, err error) {
	p := &parser{
		Buffer:      strings.ToLower(s),
		currentTime: baseTime,
	}

	p.Init()

	if err := p.Parse(); err != nil {
		return time.Time{}, time.Time{}, err
	}

	p.Execute()
	return p.start, p.end, nil
}
