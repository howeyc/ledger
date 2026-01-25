package ledger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alfredxing/calc/compute"
	"github.com/howeyc/ledger/decimal"
	date "github.com/joyt/godate"
)

// ParseLedgerFile parses a ledger file and returns a list of Transactions.
func ParseLedgerFile(filename string) (generalLedger []*Transaction, err error) {
	ifile, ierr := os.Open(filename)
	if ierr != nil {
		return nil, ierr
	}
	defer ifile.Close()
	var mu sync.Mutex
	parseLedger(filename, ifile, func(t []*Transaction, e error) (stop bool) {
		if e != nil {
			err = e
			stop = true
			return
		}

		mu.Lock()
		generalLedger = append(generalLedger, t...)
		mu.Unlock()
		return
	})

	return
}

// ParseLedger parses a ledger file and returns a list of Transactions.
func ParseLedger(ledgerReader io.Reader) (generalLedger []*Transaction, err error) {
	parseLedger("", ledgerReader, func(t []*Transaction, e error) (stop bool) {
		if e != nil {
			err = e
			stop = true
			return
		}

		generalLedger = append(generalLedger, t...)
		return
	})

	return
}

// ParseLedgerAsync parses a ledger file and returns a Transaction and error channels .
func ParseLedgerAsync(ledgerReader io.Reader) (c chan *Transaction, e chan error) {
	c = make(chan *Transaction)
	e = make(chan error)

	go func() {
		parseLedger("", ledgerReader, func(tlist []*Transaction, err error) (stop bool) {
			if err != nil {
				e <- err
			} else {
				for _, t := range tlist {
					c <- t
				}
			}
			return
		})

		e <- nil
		close(c)
		close(e)
	}()
	return c, e
}

type parser struct {
	scanner *linescanner

	comments   []string
	dateLayout string

	strPrevDate string
	prevDateErr error
	prevDate    time.Time

	transactions []Transaction
	ctIdx        int
	postings     []Account
	cpIdx        int
}

const preAllocSize = 100000
const preAllocWarn = 10

func (p *parser) init() {
	p.transactions = make([]Transaction, preAllocSize)
	p.postings = make([]Account, preAllocSize*3)
	p.ctIdx = 0
	p.cpIdx = 0
}

func (p *parser) grow() {
	if len(p.transactions)-p.ctIdx < preAllocWarn ||
		len(p.postings)-p.cpIdx < (preAllocWarn*3) {
		p.init()
	}
}

func parseLedger(filename string, ledgerReader io.Reader, callback func(t []*Transaction, err error) (stop bool)) (stop bool) {
	var lp parser
	lp.init()
	lp.scanner = newLineScanner(filename, ledgerReader)

	var tlist []*Transaction

	for lp.scanner.Scan() {
		// remove heading and tailing space from the line
		trimmedLine := strings.TrimSpace(lp.scanner.Text())

		var currentComment string
		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			currentComment = trimmedLine[commentIdx:]
			trimmedLine = trimmedLine[:commentIdx]
			trimmedLine = strings.TrimSpace(trimmedLine)
		}

		// Skip empty lines
		if len(trimmedLine) == 0 {
			if len(currentComment) > 0 {
				lp.comments = append(lp.comments, currentComment)
			}
			continue
		}

		before, after, split := strings.Cut(trimmedLine, " ")
		if !split {
			if callback(nil, fmt.Errorf("%s:%d: unable to parse transaction: %w", lp.scanner.Name(), lp.scanner.LineNumber(),
				fmt.Errorf("unable to parse payee line: %s", trimmedLine))) {
				return true
			}
			if len(currentComment) > 0 {
				lp.comments = append(lp.comments, currentComment)
			}
			continue
		}
		switch before {
		case "account":
			lp.skipAccount()
		case "include":
			paths, _ := filepath.Glob(filepath.Join(filepath.Dir(lp.scanner.Name()), after))
			if len(paths) < 1 {
				callback(nil, fmt.Errorf("%s:%d: unable to include file(%s): %w", lp.scanner.Name(), lp.scanner.LineNumber(), after, errors.New("not found")))
				return true
			}
			var wg sync.WaitGroup
			for _, incpath := range paths {
				wg.Add(1)
				go func(ipath string) {
					ifile, _ := os.Open(ipath)
					defer ifile.Close()
					if parseLedger(ipath, ifile, callback) {
						stop = true
					}
					wg.Done()
				}(incpath)
			}
			wg.Wait()
			if stop {
				return stop
			}
		default:
			trans, transErr := lp.parseTransaction(before, after, currentComment)
			if transErr != nil {
				if callback(nil, fmt.Errorf("%s:%d: unable to parse transaction: %w", lp.scanner.Name(), lp.scanner.LineNumber(), transErr)) {
					return true
				}
				continue
			}
			tlist = append(tlist, trans)
		}
	}
	callback(tlist, nil)
	return false
}

func (lp *parser) skipAccount() {
	for lp.scanner.Scan() {
		// Read until blank line (ignore all sub-directives)
		if len(lp.scanner.Text()) == 0 {
			return
		}
	}
}

func (lp *parser) parseDate(dateString string) (transDate time.Time, err error) {
	// seen before, skip parse
	if lp.strPrevDate == dateString {
		return lp.prevDate, lp.prevDateErr
	}

	// try current date layout
	transDate, err = time.Parse(lp.dateLayout, dateString)
	if err != nil {
		// try to find new date layout
		transDate, lp.dateLayout, err = date.ParseAndGetLayout(dateString)
		if err != nil {
			err = fmt.Errorf("unable to parse date(%s): %w", dateString, err)
		}
	}

	// maybe next date is same
	lp.strPrevDate = dateString
	lp.prevDate = transDate
	lp.prevDateErr = err

	return
}

func (a *Account) parsePosting(trimmedLine string) (err error) {
	trimmedLine = strings.TrimSpace(trimmedLine)

	// Regex groups:
	// 1: account name
	// 2: amount (number or parenthesized expression)
	// 3: @@ converted amount
	// 4: @ conversion rate
	re := regexp.MustCompile(
		`^(?P<name>.+?)` +
			`(?:(?:\s{2,}|\t)` +
			`(?:(?P<currency>[A-Z\$]+)\s+)?` +
			`(?P<amount>[\-]?\d+(?:\.\d+)?|\([0-9+\-*\/. ]+\))` +
			`(?:\s*(?:@@\s*` +
			`(?P<converted>[\-]?\d+(?:\.\d+)?)|@\s*` +
			`(?P<factor>[\-]?\d+(?:\.\d+)?)))?)?\s*$`,
	)

	m := re.FindStringSubmatch(trimmedLine)
	if m == nil {
		return fmt.Errorf("invalid posting: %q", trimmedLine)
	}

	a.Name = m[1]
	a.Currency = m[2]

	if m[3] != "" {
		bal, err := compute.Evaluate(m[3])
		if err != nil {
			return err
		}
		a.Balance = decimal.NewFromFloat(bal)
	}

	// @@ explicit converted amount
	if m[4] != "" {
		conv, err := decimal.NewFromString(m[4])
		if err != nil {
			return err
		}
		a.Converted = &conv
	}

	// @ rate-based conversion
	if m[5] != "" {
		rate, err := decimal.NewFromString(m[5])
		if err != nil {
			return err
		}
		a.ConversionFactor = &rate
	}
	return
}

func (lp *parser) parseTransaction(dateString, payeeString, payeeComment string) (trans *Transaction, err error) {
	transDate, derr := lp.parseDate(dateString)
	if derr != nil {
		return nil, derr
	}

	transBal := decimal.Zero
	var numEmpty int
	var emptyAccIndex int
	var accIndex int

	for lp.scanner.Scan() {
		trimmedLine := lp.scanner.Text()

		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			currentComment := trimmedLine[commentIdx:]
			trimmedLine = trimmedLine[:commentIdx]
			trimmedLine = strings.TrimSpace(trimmedLine)
			if len(trimmedLine) == 0 {
				lp.comments = append(lp.comments, currentComment)
				continue
			}
			lp.postings[lp.cpIdx+accIndex].Comment = currentComment
		}

		if len(trimmedLine) == 0 {
			break
		}

		_ = lp.postings[lp.cpIdx+accIndex].parsePosting(trimmedLine)

		if lp.postings[lp.cpIdx+accIndex].Balance.IsZero() {
			numEmpty++
			emptyAccIndex = accIndex
		}

		if lp.postings[lp.cpIdx+accIndex].Converted != nil {
			transBal = transBal.Add(lp.postings[lp.cpIdx+accIndex].Converted.Neg())
		} else if lp.postings[lp.cpIdx+accIndex].ConversionFactor != nil {
			transBal = transBal.Add(lp.postings[lp.cpIdx+accIndex].Balance.Mul(
				*lp.postings[lp.cpIdx+accIndex].ConversionFactor,
			))
		} else {
			transBal = transBal.Add(lp.postings[lp.cpIdx+accIndex].Balance)
		}
		accIndex++
	}

	if accIndex < 2 {
		err = errors.New("need at least two postings")
		return
	}

	if !transBal.IsZero() {
		switch numEmpty {
		case 0:
			return nil, errors.New("unable to balance transaction: no empty account to place extra balance")
		case 1:
			// If there is a single empty account, then it is obvious where to
			// place the remaining balance.
			lp.postings[lp.cpIdx+emptyAccIndex].Balance = transBal.Neg()
		default:
			return nil, errors.New("unable to balance transaction: more than one account empty")
		}
	}

	lp.transactions[lp.ctIdx].Payee = payeeString
	lp.transactions[lp.ctIdx].Date = transDate
	lp.transactions[lp.ctIdx].PayeeComment = payeeComment
	lp.transactions[lp.ctIdx].AccountChanges = lp.postings[lp.cpIdx : lp.cpIdx+accIndex]
	lp.transactions[lp.ctIdx].Comments = lp.comments

	trans = &lp.transactions[lp.ctIdx]

	lp.comments = nil
	lp.cpIdx += accIndex
	lp.ctIdx++

	lp.grow()

	return
}
