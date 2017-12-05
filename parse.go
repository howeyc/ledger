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
func ParseLedger(ledgerReader io.Reader) (generalLedger []*Transaction, err error) {
	c, e := ParseLedgerAsync(ledgerReader)
	for {
		select {
		case trans := <-c:
			generalLedger = append(generalLedger, trans)
		case err := <-e:
			sort.Sort(sortTransactionsByDate{generalLedger})
			return generalLedger, err
		}
	}
}

// ParseLedgerAsync parses a ledger file and returns a Transaction and error channels .
//
func ParseLedgerAsync(ledgerReader io.Reader) (c chan *Transaction, e chan error) {
	c = make(chan *Transaction)
	e = make(chan error)

	go func() {

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
						e <- fmt.Errorf("%d: Unable to balance transaction, %s", lineCount, transErr)
					}
					c <- trans
					trans = nil
				}
			} else if trans == nil {
				lineSplit := strings.SplitN(line, " ", 2)
				if len(lineSplit) != 2 {
					e <- fmt.Errorf("%d: Unable to parse payee line: %s", lineCount, line)
				}
				dateString := lineSplit[0]
				transDate, dateErr := date.Parse(dateString)
				if dateErr != nil {
					e <- fmt.Errorf("%d: Unable to parse date: %s", lineCount, dateString)
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
				e <- fmt.Errorf("%d: Unable to balance transaction, %s", lineCount, transErr)
			}
			c <- trans
			trans = nil
		}
		e <- nil
	}()
	return c, e
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
	var emptyAccPtr *Account
	var emptyAccIndex int
	for accIndex, accChange := range input.AccountChanges {
		if accChange.Balance == nil {
			if emptyAccPtr != nil {
				return fmt.Errorf("More than one account change empty!")
			}
			emptyAccPtr = &accChange
			emptyAccIndex = accIndex
		} else {
			balance = balance.Add(balance, accChange.Balance)
		}
	}
	if balance.Sign() != 0 {
		if emptyAccPtr == nil {
			return fmt.Errorf("No empty account change to place extra balance!")
		}
		input.AccountChanges[emptyAccIndex].Balance = balance.Neg(balance)
	}
	return nil
}
