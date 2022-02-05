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
//
// Transactions are sorted by date.
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
//
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

func parseLedger(ledgerReader io.Reader, callback func(t *Transaction, err error) (stop bool)) {
	scanner := bufio.NewScanner(ledgerReader)
	var line string
	var filename string
	var lineCount int
	var comments []string

	// default date layout
	dateLayout := "2006/01/02"

	errorMsg := func(msg string) (stop bool) {
		return callback(nil, fmt.Errorf("%s:%d: %s", filename, lineCount, msg))
	}

	for scanner.Scan() {
		line = scanner.Text()

		// update filename/line if sentinel comment is found
		if strings.HasPrefix(line, markerPrefix) {
			filename, lineCount = parseMarker(line)
			continue
		}

		// remove heading and tailing space from the line
		trimmedLine := strings.TrimSpace(line)
		lineCount++

		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			comments = append(comments, trimmedLine[commentIdx:])
			trimmedLine = trimmedLine[:commentIdx]
			if len(trimmedLine) == 0 {
				continue
			}
		}

		// Skip empty lines
		if len(trimmedLine) == 0 {
			continue
		}

		lineSplit := strings.SplitN(trimmedLine, " ", 2)
		if len(lineSplit) != 2 {
			if errorMsg("Unable to parse payee line: " + line) {
				return
			}
			continue
		}
		commandDirective := lineSplit[0]
		switch commandDirective {
		case "account":
			_, lines, _ := parseAccount(scanner)
			lineCount += lines
		default:
			trans, lines, layout, transErr := parseTransaction(dateLayout, scanner)
			dateLayout = layout
			lineCount += lines
			if transErr != nil {
				if errorMsg(fmt.Errorf("Unable to parse transaction: %w", transErr).Error()) {
					return
				}
				continue
			}
			trans.Comments = append(comments, trans.Comments...)
			callback(trans, nil)
			comments = nil
		}
	}
}

func parseAccount(scanner *bufio.Scanner) (accountName string, lines int, err error) {
	line := scanner.Text()
	// remove heading and tailing space from the line
	trimmedLine := strings.TrimSpace(line)

	lineSplit := strings.SplitN(trimmedLine, " ", 2)
	if len(lineSplit) != 2 {
		err = fmt.Errorf("Unable to parse account line: %s", line)
		return
	}
	accountName = lineSplit[1]

	for scanner.Scan() {
		// Read until blank line (ignore all sub-directives)
		line = scanner.Text()
		// remove heading and tailing space from the line
		trimmedLine = strings.TrimSpace(line)
		lines++

		// skip comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			trimmedLine = trimmedLine[:commentIdx]
		}

		// continue slurping up sub-directives until empty line
		if len(trimmedLine) == 0 {
			return
		}
	}

	return
}

func parseTransaction(currentDateLayout string, scanner *bufio.Scanner) (trans *Transaction, lines int, layout string, err error) {
	var comments []string

	line := scanner.Text()
	trimmedLine := strings.TrimSpace(line)
	// handle comments (comment saved in calling function)
	if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
		trimmedLine = trimmedLine[:commentIdx]
	}
	trimmedLine = strings.TrimSpace(trimmedLine)

	// Parse Date-Payee line
	lineSplit := strings.SplitN(trimmedLine, " ", 2)
	if len(lineSplit) != 2 {
		err = fmt.Errorf("Unable to parse payee line: %s", line)
		return
	}
	dateString := lineSplit[0]

	// attempt currentDateLayout, hopefully file is consistent
	layout = currentDateLayout
	transDate, dateErr := time.Parse(layout, dateString)
	if dateErr != nil {
		// try to find new date layout
		transDate, layout, dateErr = date.ParseAndGetLayout(dateString)
	}
	if dateErr != nil {
		err = fmt.Errorf("Unable to parse date: %s", dateString)
		return
	}
	payeeString := lineSplit[1]
	trans = &Transaction{Payee: payeeString, Date: transDate}

	for scanner.Scan() {
		line = scanner.Text()
		// remove heading and tailing space from the line
		trimmedLine = strings.TrimSpace(line)
		lines++

		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			comments = append(comments, trimmedLine[commentIdx:])
			trimmedLine = trimmedLine[:commentIdx]
			if len(trimmedLine) == 0 {
				continue
			}
			trimmedLine = strings.TrimSpace(trimmedLine)
		}

		if len(trimmedLine) == 0 {
			if trans != nil {
				transErr := balanceTransaction(trans)
				if transErr != nil {
					err = fmt.Errorf("Unable to balance transaction: %w", transErr)
					return
				}
				trans.Comments = comments
				return
			}
		} else {
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
	}
	// If the file does not end on empty line, we must attempt to balance last
	// transaction of the file.
	if trans != nil {
		transErr := balanceTransaction(trans)
		if transErr != nil {
			err = fmt.Errorf("Unable to balance transaction: %w", transErr)
			return
		}
		trans.Comments = comments
	}
	return
}

// Takes a transaction and balances it. This is mainly to fill in the empty part
// with the remaining balance.
func balanceTransaction(input *Transaction) error {
	balance := new(big.Rat)
	var emptyFound bool
	var emptyAccIndex int
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
