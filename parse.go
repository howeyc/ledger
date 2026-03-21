package ledger

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/araddon/dateparse"
	"github.com/expr-lang/expr"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
)

// ParseLedgerFile parses a ledger file and returns a list of Transactions.
func ParseLedgerFile(filename string) (generalLedger []*Transaction, err error) {
	ifile, ierr := os.Open(filename)
	if ierr != nil {
		return nil, ierr
	}
	defer ifile.Close()
	return ParseLedger(filename, ifile)
}

// ParseLedger parses a ledger file and returns a list of Transactions.
func ParseLedger(name string, ledgerReader io.Reader) (generalLedger []*Transaction, err error) {
	blocks, err := parseBlocks(name, ledgerReader)
	if err != nil {
		return nil, err
	}

	return lo.MapErr(blocks, func(b block, _ int) (*Transaction, error) {
		trans, transErr := b.parseTransaction()
		if transErr != nil {
			return nil, fmt.Errorf("%s:%d: unable to parse transaction: %w", b.filename, b.lineNum, transErr)
		}
		return trans, nil
	})
}

type parser struct {
	scanner *linescanner

	comments   []string
	dateLayout string

	strPrevDate string
	prevDateErr error
	prevDate    time.Time
}

func parseBlocks(filename string, ledgerReader io.Reader) ([]block, error) {
	var lp parser
	lp.scanner = newLineScanner(filename, ledgerReader)

	blocks := []block{}
	comments := []string{}
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
				comments = append(comments, currentComment)
			}
			continue
		}

		before, after, split := strings.Cut(trimmedLine, " ")
		if !split {
			return nil, fmt.Errorf(
				"%s:%d: unable to parse transaction: %w",
				lp.scanner.Name(),
				lp.scanner.LineNumber(),
				fmt.Errorf("unable to parse payee line: %s", trimmedLine),
			)
		}
		switch before {
		case "account":
			lp.skipAccount()
		case "include":
			paths, _ := filepath.Glob(filepath.Join(filepath.Dir(lp.scanner.Name()), after))
			if len(paths) < 1 {
				return nil, fmt.Errorf(
					"%s:%d: unable to include file(%s): %w", lp.scanner.Name(), lp.scanner.LineNumber(), after, errors.New("not found"))
			}

			b, err := lo.FlatMapErr(paths, func(path string, _ int) ([]block, error) {
				f, _ := os.Open(path)
				defer f.Close()
				return parseBlocks(path, f)
			})
			if err != nil {
				return nil, err
			}
			blocks = append(blocks, b...)
		default:
			transDate, derr := lp.parseDate(before)
			if derr != nil {
				return nil, fmt.Errorf("%s:%d: unable to parse transaction: %w", lp.scanner.Name(), lp.scanner.LineNumber(), derr)
			}

			blocks = append(blocks, lp.parseBlock(transDate, after, currentComment, comments))
			comments = []string{}
		}
	}

	return blocks, nil
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

	// Use dateparse to handle flexible date formats
	transDate, err = dateparse.ParseAny(dateString)
	if err != nil {
		err = fmt.Errorf("unable to parse date(%s): %w", dateString, err)
	}

	// maybe next date is same
	lp.strPrevDate = dateString
	lp.prevDate = transDate
	lp.prevDateErr = err

	return
}

func (a *Account) parsePosting(trimmedLine string, comment string) (err error) {
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
	a.Comment = comment

	if m[3] != "" {
		program, err := expr.Compile(m[3])
		if err != nil {
			return err
		}
		out, err := expr.Run(program, nil)
		if err != nil {
			return err
		}

		var f float64
		switch v := out.(type) {
		case int:
			f = float64(v)
		case int64:
			f = float64(v)
		case float32:
			f = float64(v)
		case float64:
			f = v
		default:
			return fmt.Errorf("expression did not evaluate to a number: %T", out)
		}

		a.Balance = decimal.NewFromFloat(f)
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

type block struct {
	transDate    time.Time
	payeeString  string
	payeeComment string
	comments     []string
	lines        []string
	filename     string
	lineNum      int
}

func (lp *parser) parseBlock(transDate time.Time, payeeString, payeeComment string, comments []string) block {
	lines := []string{}
	for lp.scanner.Scan() {
		trimmedLine := lp.scanner.Text()
		lines = append(lines, trimmedLine)
		if len(trimmedLine) == 0 {
			break
		}
	}

	return block{
		transDate:    transDate,
		payeeString:  payeeString,
		payeeComment: payeeComment,
		comments:     comments,
		lines:        lines,
		filename:     lp.scanner.Name(),
		lineNum:      lp.scanner.LineNumber(),
	}
}

func (b *block) parseTransaction() (trans *Transaction, err error) {
	trans = &Transaction{}
	for _, trimmedLine := range b.lines {
		postingComment := ""
		// handle comments
		if commentIdx := strings.Index(trimmedLine, ";"); commentIdx >= 0 {
			currentComment := trimmedLine[commentIdx:]
			trimmedLine = trimmedLine[:commentIdx]
			trimmedLine = strings.TrimSpace(trimmedLine)
			if len(trimmedLine) == 0 {
				b.comments = append(b.comments, currentComment)
				continue
			}
			postingComment = currentComment
		}

		if len(trimmedLine) == 0 {
			break
		}

		posting := Account{}
		posting.parsePosting(trimmedLine, postingComment)
		trans.AccountChanges = append(trans.AccountChanges, posting)
	}

	trans.Payee = b.payeeString
	trans.Date = b.transDate
	trans.PayeeComment = b.payeeComment
	if len(b.comments) > 0 {
		trans.Comments = b.comments
	}

	if err = trans.IsBalanced(); err != nil {
		return nil, err
	}

	return
}
