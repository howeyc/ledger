package cmd

import (
	"strings"

	"github.com/howeyc/ledger"
)

func strToPeriod(p string) ledger.Period {
	pl := strings.ToLower(p)

	switch pl {
	case "daily":
		return ledger.PeriodDay
	case "weekly":
		return ledger.PeriodWeek
	case "biweekly", "bi-weekly":
		return ledger.Period2Week
	case "monthly":
		return ledger.PeriodMonth
	case "bimonthly", "bi-monthly":
		return ledger.Period2Month
	case "quarterly":
		return ledger.PeriodQuarter
	case "semiyearly", "semi-yearly":
		return ledger.PeriodSemiYear
	case "yearly":
		return ledger.PeriodYear
	}

	return ledger.Period("")
}
