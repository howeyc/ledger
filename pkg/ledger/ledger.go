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

var TransactionDateFormat string
var DisplayPrecision int

func init() {
	TransactionDateFormat = "2006/01/02"
	DisplayPrecision = 2
}

type Account struct {
	Name    string
	Balance *big.Rat
}

type sortAccounts []*Account

func (s sortAccounts) Len() int      { return len(s) }
func (s sortAccounts) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortAccountsByName struct{ sortAccounts }

func (s sortAccountsByName) Less(i, j int) bool {
	return s.sortAccounts[i].Name < s.sortAccounts[j].Name
}

type Transaction struct {
	Payee          string
	Date           time.Time
	AccountChanges []Account
}

type sortTransactions []*Transaction

func (s sortTransactions) Len() int      { return len(s) }
func (s sortTransactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type sortTransactionsByDate struct{ sortTransactions }

func (s sortTransactionsByDate) Less(i, j int) bool {
	return s.sortTransactions[i].Date.Before(s.sortTransactions[j].Date)
}

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
	sort.Sort(sortTransactionsByDate{generalLedger})
	return generalLedger, scanner.Err()
}

// Given a list of transactions and filter strings, returns account balances of
// all accounts that have any filter as a substring of the account name. Also
// returns balances for each account level depth as a separate record.
//
// Accounts are sorted by name.
func GetBalances(generalLedger []*Transaction, filterArr []string) []*Account {
	balances := make(map[string]*big.Rat)
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
			if inFilter {
				accHier := strings.Split(accChange.Name, ":")
				accDepth := len(accHier)
				for currDepth := accDepth; currDepth > 0; currDepth-- {
					currAccName := strings.Join(accHier[:currDepth], ":")
					if ratNum, ok := balances[currAccName]; !ok {
						ratNum = new(big.Rat)
						ratNum.SetString(accChange.Balance.RatString())
						balances[currAccName] = ratNum
					} else {
						ratNum.Add(ratNum, accChange.Balance)
					}
				}
			}
		}
	}

	accList := make([]*Account, len(balances))
	count := 0
	for accName, accBalance := range balances {
		account := &Account{Name: accName, Balance: accBalance}
		accList[count] = account
		count++
	}

	sort.Sort(sortAccountsByName{accList})
	return accList
}

// Prints out account balances formated to a windows of a width of columns.
// Only shows accounts with names less than or equal to the given depth.
func PrintBalances(accountList []*Account, printZeroBalances bool, depth, columns int) {
	overallBalance := new(big.Rat)
	for _, account := range accountList {
		accDepth := len(strings.Split(account.Name, ":"))
		if accDepth == 1 {
			overallBalance.Add(overallBalance, account.Balance)
		}
		if (printZeroBalances || account.Balance.Sign() != 0) && (depth < 0 || accDepth <= depth) {
			outBalanceString := account.Balance.FloatString(DisplayPrecision)
			spaceCount := columns - len(account.Name) - len(outBalanceString)
			fmt.Printf("%s%s%s\n", account.Name, strings.Repeat(" ", spaceCount), outBalanceString)
		}
	}
	fmt.Println(strings.Repeat("-", columns))
	outBalanceString := overallBalance.FloatString(DisplayPrecision)
	spaceCount := columns - len(outBalanceString)
	fmt.Printf("%s%s\n", strings.Repeat(" ", spaceCount), outBalanceString)
}

// Prints all transactions as a formatted ledger file.
func PrintLedger(w io.Writer, generalLedger []*Transaction, columns int) {
	for _, trans := range generalLedger {
		fmt.Fprintf(w, "%s %s\n", trans.Date.Format(TransactionDateFormat), trans.Payee)
		for _, accChange := range trans.AccountChanges {
			outBalanceString := accChange.Balance.FloatString(DisplayPrecision)
			spaceCount := columns - 4 - len(accChange.Name) - len(outBalanceString)
			fmt.Fprintf(w, "    %s%s%s\n", accChange.Name, strings.Repeat(" ", spaceCount), outBalanceString)
		}
		fmt.Fprintln(w, "")
	}
}

// Prints each transaction that matches the given filters.
func PrintRegister(generalLedger []*Transaction, filterArr []string, columns int) {
	runningBalance := new(big.Rat)
	for _, trans := range generalLedger {
		for _, accChange := range trans.AccountChanges {
			inFilter := len(filterArr) == 0
			for _, filter := range filterArr {
				if strings.Contains(accChange.Name, filter) {
					inFilter = true
				}
			}
			if inFilter {
				runningBalance.Add(runningBalance, accChange.Balance)
				writtenBytes, _ := fmt.Printf("%s %s", trans.Date.Format(TransactionDateFormat), trans.Payee)
				outBalanceString := accChange.Balance.FloatString(DisplayPrecision)
				outRunningBalanceString := runningBalance.FloatString(DisplayPrecision)
				spaceCount := columns - writtenBytes - 2 - len(outBalanceString) - len(outRunningBalanceString)
				if spaceCount < 0 {
					spaceCount = 0
				}
				fmt.Printf("%s%s %s", strings.Repeat(" ", spaceCount), outBalanceString, outRunningBalanceString)
				fmt.Println("")
			}
		}
	}
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
