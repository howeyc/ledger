package ledger

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"
	"strings"

	"github.com/joyt/godate"
	"github.com/marcmak/calc/calc"
)

const (
	whitespace = " \t"
)

// ParseLedger parses a ledger file and returns a list of Transactions.
//
// Transactions are sorted by date.
func ParseLedger(ledgerReader io.Reader, ledgerFileName string) (generalLedger []*Transaction, err error) {
	var trans *Transaction
	scanner := bufio.NewScanner(ledgerReader)
	var line string
	var lineCount int

	accountToAmountSpace := regexp.MustCompile(" {2,}|\t+")
	for scanner.Scan() {
		line = scanner.Text()
		// remove heading and tailing space from the line
		trimmedLine := strings.Trim(line, whitespace)
		lineCount++
		if strings.HasPrefix(trimmedLine, ";") {
			// nop
		} else if len(trimmedLine) == 0 {
			if trans != nil {
				transErr := balanceTransaction(trans)
				if transErr != nil {
					return generalLedger, fmt.Errorf("%s:%d: Unable to balance transaction, %s", ledgerFileName, lineCount, transErr)
				}
				generalLedger = append(generalLedger, trans)
				trans = nil
			}
		} else if trans == nil {
			lineSplit := strings.SplitN(line, " ", 2)
			if len(lineSplit) != 2 {
				return generalLedger, fmt.Errorf("%s:%d: Unable to parse payee line: %s", ledgerFileName, lineCount, line)
			}
			dateString := lineSplit[0]
			transDate, dateErr := date.Parse(dateString)
			if dateErr != nil {
				return generalLedger, fmt.Errorf("%d: Unable to parse date: %s", lineCount, dateString)
			}
			payeeString := lineSplit[1]
			trans = &Transaction{Payee: payeeString, Date: transDate}
		} else {
			var accChange Account
			lineSplit := accountToAmountSpace.Split(trimmedLine, -1)
			nonEmptyWords := []string{}
			for _, word := range lineSplit {
				if len(word) > 0 {
					nonEmptyWords = append(nonEmptyWords, word)
				}
			}
			lastIndex := len(nonEmptyWords) - 1
			balErr, rationalNum := getBalance(strings.Trim(nonEmptyWords[lastIndex], whitespace))
			if balErr == false {
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
			return generalLedger, fmt.Errorf("%d: Unable to balance transaction, %s", lineCount, transErr)
		}
		generalLedger = append(generalLedger, trans)
		trans = nil
	}
	sort.Sort(sortTransactionsByDate{generalLedger})
	return generalLedger, scanner.Err()
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
				return fmt.Errorf("More than one account change empty!")
			}
			emptyAccIndex = accIndex
			emptyFound = true
		} else {
			balance = balance.Add(balance, accChange.Balance)
		}
	}
	if balance.Sign() != 0 {
		if !emptyFound {
			return fmt.Errorf("No empty account change to place extra balance!")
		}
	}
	if emptyFound {
		input.AccountChanges[emptyAccIndex].Balance = balance.Neg(balance)
	}
	return nil
}
