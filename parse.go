package ledger

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/alfredxing/calc/compute"
	date "github.com/joyt/godate"
)

// ParseLedger parses a ledger file and returns a list of Transactions.
func ParseLedger(ledgerReader io.Reader) (generalLedger []*Transaction, err error) {
	parseLedger(ledgerReader, func(t *Transaction, e error) (stop bool) {
		if e != nil {
			err = e
			stop = true
			return
		}

		generalLedger = append(generalLedger, t)
		return
	})

	return
}

// ParseLedgerAsync parses a ledger file and returns a Transaction and error channels .
func ParseLedgerAsync(ledgerReader io.Reader) (c chan *Transaction, e chan error) {
	c = make(chan *Transaction)
	e = make(chan error)

	go func() {
		parseLedger(ledgerReader, func(t *Transaction, err error) (stop bool) {
			if err != nil {
				e <- err
			} else {
				c <- t
			}
			return
		})

		e <- nil
	}()
	return c, e
}

// Calculation expressions are enclosed in parantheses
var calcExpr = regexp.MustCompile(`(?s) \((.*)\)`)

type parser struct {
	scanner    *bufio.Scanner
	filename   string
	lineCount  int
	comments   []string
	dateLayout string
}

func parseLedger(ledgerReader io.Reader, callback func(t *Transaction, err error) (stop bool)) {
	var lp parser
	lp.scanner = bufio.NewScanner(ledgerReader)

	var line string
	for lp.scanner.Scan() {
		line = lp.scanner.Text()

		// update filename/line if sentinel comment is found
		if strings.HasPrefix(line, markerPrefix) {
			lp.filename, lp.lineCount = parseMarker(line)
			continue
		}

		// remove heading and tailing space from the line
		trimmedLine := strings.TrimSpace(line)
		lp.lineCount++

		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			lp.comments = append(lp.comments, trimmedLine[commentIdx:])
			trimmedLine = trimmedLine[:commentIdx]
			trimmedLine = strings.TrimSpace(trimmedLine)
		}

		// Skip empty lines
		if len(trimmedLine) == 0 {
			continue
		}

		before, after, split := strings.Cut(trimmedLine, " ")
		if !split {
			if callback(nil, fmt.Errorf("%s:%d: Unable to parse transaction: %w", lp.filename, lp.lineCount,
				fmt.Errorf("Unable to parse payee line: %s", line))) {
				return
			}
			continue
		}
		switch before {
		case "account":
			lp.parseAccount(after)
		default:
			trans, transErr := lp.parseTransaction(before, after)
			if transErr != nil {
				if callback(nil, fmt.Errorf("%s:%d: Unable to parse transaction: %w", lp.filename, lp.lineCount, transErr)) {
					return
				}
				continue
			}
			callback(trans, nil)
		}
	}
}

func (lp *parser) parseAccount(accName string) (accountName string, err error) {
	accountName = accName

	var line string
	for lp.scanner.Scan() {
		// Read until blank line (ignore all sub-directives)
		line = lp.scanner.Text()
		// remove heading and tailing space from the line
		trimmedLine := strings.TrimSpace(line)
		lp.lineCount++

		// skip comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			trimmedLine = trimmedLine[:commentIdx]
		}

		// stop slurping up sub-directives on empty line
		if len(trimmedLine) == 0 {
			return
		}
	}

	return
}

func (lp *parser) parseDate(dateString string) (transDate time.Time, err error) {
	// try curent date layout
	transDate, err = time.Parse(lp.dateLayout, dateString)
	if err != nil {
		// try to find new date layout
		transDate, lp.dateLayout, err = date.ParseAndGetLayout(dateString)
		if err != nil {
			err = fmt.Errorf("Unable to parse date(%s): %w", dateString, err)
		}
	}
	return
}

func (lp *parser) parseTransaction(dateString, payeeString string) (trans *Transaction, err error) {
	transDate, derr := lp.parseDate(dateString)
	if derr != nil {
		return nil, derr
	}
	trans = &Transaction{Payee: payeeString, Date: transDate}

	var line string
	for lp.scanner.Scan() {
		line = lp.scanner.Text()
		// remove heading and tailing space from the line
		trimmedLine := strings.TrimSpace(line)
		lp.lineCount++

		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			lp.comments = append(lp.comments, trimmedLine[commentIdx:])
			trimmedLine = trimmedLine[:commentIdx]
			trimmedLine = strings.TrimSpace(trimmedLine)
			if len(trimmedLine) == 0 {
				continue
			}
		}

		if len(trimmedLine) == 0 {
			break
		}

		// Check for expr
		trimmedLine = calcExpr.ReplaceAllStringFunc(trimmedLine, func(s string) string {
			f, _ := compute.Evaluate(s)
			return fmt.Sprintf("%f", f)
		})

		var accChange Account
		accChange.Name = trimmedLine
		if i := strings.LastIndexFunc(trimmedLine, unicode.IsSpace); i >= 0 {
			acc := strings.TrimSpace(trimmedLine[:i])
			amt := trimmedLine[i+1:]
			if ratbal, valid := new(big.Rat).SetString(amt); valid {
				accChange.Name = acc
				accChange.Balance = ratbal
			}
		}
		trans.AccountChanges = append(trans.AccountChanges, accChange)
	}

	transErr := balanceTransaction(trans)
	if transErr != nil {
		err = fmt.Errorf("Unable to balance transaction: %w", transErr)
		return
	}
	trans.Comments = lp.comments
	lp.comments = nil
	return
}

// Takes a transaction and balances it. This is mainly to fill in the empty part
// with the remaining balance.
func balanceTransaction(input *Transaction) error {
	balance := new(big.Rat)
	var emptyFound bool
	var emptyAccIndex int
	if len(input.AccountChanges) < 2 {
		return fmt.Errorf("need at least two postings")
	}
	for accIndex, accChange := range input.AccountChanges {
		if accChange.Balance == nil {
			if emptyFound {
				return fmt.Errorf("more than one account empty")
			}
			emptyAccIndex = accIndex
			emptyFound = true
		} else {
			balance = balance.Add(balance, accChange.Balance)
		}
	}
	if balance.Sign() != 0 {
		if !emptyFound {
			return fmt.Errorf("no empty account to place extra balance")
		}
	}
	if emptyFound {
		input.AccountChanges[emptyAccIndex].Balance = balance.Neg(balance)
	}

	return nil
}
