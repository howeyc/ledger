package main

import (
	"bufio"
	"fmt"
	"io"
	"math/big"
	"strings"
	"time"
)

type Account struct {
	Name    string
	Balance *big.Rat
}

type Transaction struct {
	Payee          string
	Date           time.Time
	AccountChanges []Account
}

func parseLedger(ledgerReader io.Reader) (generalLedger []Transaction, err error) {
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
				generalLedger = append(generalLedger, *trans)
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
			accChange.Name = strings.Join(nonEmptyWords[:lastIndex], " ")
			rationalNum := new(big.Rat)
			_, balErr := rationalNum.SetString(nonEmptyWords[lastIndex])
			if balErr == false {
				// Assuming no balance and whole line is account name
				accChange.Name = strings.Join(nonEmptyWords, " ")
				//	return generalLedger, fmt.Errorf("%d: Unable to parse value: %s", lineCount, nonEmptyWords[lastIndex])
			} else {
				accChange.Name = strings.Join(nonEmptyWords[:lastIndex], " ")
				accChange.Balance = rationalNum
			}
			trans.AccountChanges = append(trans.AccountChanges, accChange)
		}
	}
	return generalLedger, scanner.Err()
}

func printLedger(w io.Writer, generalLedger []Transaction) {
	for _, trans := range generalLedger {
		fmt.Fprintf(w, "%s %s\n", trans.Date.Format(TransactionDateFormat), trans.Payee)
		for _, accChange := range trans.AccountChanges {
			fmt.Fprintf(w, "    %s          %s\n", accChange.Name, accChange.Balance.FloatString(2))
		}
		fmt.Fprintln(w, "")
	}
}

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
