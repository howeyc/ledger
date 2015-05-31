package ledger

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"sort"
	"strings"
	"time"
)

// Parses a ledger file and returns a list of Transactions.
//
// Transactions are sorted by date.
func ParseLedger(ledgerReader io.Reader) (generalLedger []*Transaction, err error) {
	var trans *Transaction
	scanner := bufio.NewScanner(ledgerReader)
	var line string
	var lineCount int
	for scanner.Scan() {
		line = scanner.Text()
		lineCount++
		if strings.HasPrefix(line, ";") {
			// nop
		} else if len(line) == 0 {
			if trans != nil {
				transErr := balanceTransaction(trans)
				if transErr != nil {
					return generalLedger, fmt.Errorf("%d: Unable to balance transaction, %s", lineCount, transErr)
				}
				generalLedger = append(generalLedger, trans)
				trans = nil
			}
		} else if trans == nil {
			lineSplit := strings.SplitN(line, " ", 2)
			if len(lineSplit) != 2 {
				return generalLedger, fmt.Errorf("%d: Unable to parse payee line: %s", lineCount, line)
			}
			dateString := lineSplit[0]
			transDate, dateErr := time.Parse(TransactionDateFormat, dateString)
			if dateErr != nil {
				return generalLedger, fmt.Errorf("%d: Unable to parse date: %s", lineCount, dateString)
			}
			payeeString := lineSplit[1]
			trans = &Transaction{Payee: payeeString, Date: transDate}
		} else {
			var accChange Account
			lineSplit := strings.Split(line, " ")
			nonEmptyWords := []string{}
			for _, word := range lineSplit {
				if len(word) > 0 {
					nonEmptyWords = append(nonEmptyWords, word)
				}
			}
			lastIndex := len(nonEmptyWords) - 1
			rationalNum := new(big.Rat)
			_, balErr := rationalNum.SetString(nonEmptyWords[lastIndex])
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
