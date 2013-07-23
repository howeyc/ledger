// time.go
package main

import (
	"time"
)

type Quarter int

const (
	Q0 = iota
	Q1
	Q2
	Q3
	Q4
)

func getQuarter(t time.Time) Quarter {
	switch t.Month() {
	case time.January, time.February, time.March:
		return Q1
	case time.April, time.May, time.June:
		return Q2
	case time.July, time.August, time.September:
		return Q3
	case time.October, time.November, time.December:
		return Q4
	default:
		return Q0
	}
}
