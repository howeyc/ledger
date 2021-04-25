package ledger

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"

	date "github.com/joyt/godate"
	"github.com/marcmak/calc/calc"
)

const (
	whitespace = " \t"
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

var accountToAmountSpace = regexp.MustCompile(" {2,}|\t+")

func parseLedger(ledgerReader io.Reader, callback func(t *Transaction, err error) (stop bool)) {
	scanner := bufio.NewScanner(ledgerReader)
	var line string
	var filename string
	var lineCount int
	var comments []string

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
		trimmedLine := strings.Trim(line, whitespace)
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
			trans, lines, transErr := parseTransaction(scanner)
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
	trimmedLine := strings.Trim(line, whitespace)

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
		trimmedLine = strings.Trim(line, whitespace)
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

func parseTransaction(scanner *bufio.Scanner) (trans *Transaction, lines int, err error) {
	var comments []string

	line := scanner.Text()
	trimmedLine := strings.Trim(line, whitespace)
	// handle comments (comment saved in calling function)
	if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
		trimmedLine = trimmedLine[:commentIdx]
	}
	trimmedLine = strings.Trim(trimmedLine, whitespace)

	// Parse Date-Payee line
	lineSplit := strings.SplitN(trimmedLine, " ", 2)
	if len(lineSplit) != 2 {
		err = fmt.Errorf("Unable to parse payee line: %s", line)
		return
	}
	dateString := lineSplit[0]
	transDate, dateErr := date.Parse(dateString)
	if dateErr != nil {
		err = fmt.Errorf("Unable to parse date: %s", dateString)
		return
	}
	payeeString := lineSplit[1]
	trans = &Transaction{Payee: payeeString, Date: transDate}

	for scanner.Scan() {
		line = scanner.Text()
		// remove heading and tailing space from the line
		trimmedLine = strings.Trim(line, whitespace)
		lines++

		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			comments = append(comments, trimmedLine[commentIdx:])
			trimmedLine = trimmedLine[:commentIdx]
			if len(trimmedLine) == 0 {
				continue
			}
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
			var accChange Account
			lineSplit := accountToAmountSpace.Split(trimmedLine, -1)
			var nonEmptyWords []string
			for _, word := range lineSplit {
				if len(word) > 0 {
					nonEmptyWords = append(nonEmptyWords, word)
				}
			}
			lastIndex := len(nonEmptyWords) - 1
			balErr, rationalNum := getBalance(strings.Trim(nonEmptyWords[lastIndex], whitespace))
			if !balErr {
				// Assuming no balance and whole line is account name
				accChange.Name = strings.Join(nonEmptyWords, " ")
			} else {
				accChange.Name = strings.Join(nonEmptyWords[:lastIndex], " ")
				accChange.Balance = rationalNum
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

func getBalance(balance string) (bool, *big.Rat) {
	rationalNum := new(big.Rat)
	if strings.Contains(balance, "(") {
		rationalNum.SetFloat64(calc.Solve(balance))
		return true, rationalNum
	}
	_, isValid := rationalNum.SetString(balance)
	return isValid, rationalNum
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
