package ledger

// TransactionDateFormat is the date format that is used to parse a ledger file. The default is "2006/01/02"
var TransactionDateFormat string

// DisplayPrecision specifies the number of decimal places to be displayed for balances.
var DisplayPrecision int

func init() {
	TransactionDateFormat = "2006/01/02"
	DisplayPrecision = 2
}
