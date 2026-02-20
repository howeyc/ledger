package qif

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Non-investment QIF transaction, based on the "Non-investment transaction format"
// from the GnuCash documentation. Only a subset of fields is modeled for now.
type Transaction struct {
	// Header/type line, e.g. "!Type:Cash"
	Type string `qif:"header"`

	// Core transaction fields
	Date   string `qif:"D"` // D - Date
	Amount string `qif:"T"` // T - Amount
	Num    string `qif:"N"` // N - Number (check/reference)
	Payee  string `qif:"P"` // P - Payee/description
	Memo   string `qif:"M"` // M - Memo
	Addr   string `qif:"A"` // A - Address (multi-line; kept concatenated with '\n')
	Cleared string `qif:"C"` // C - Cleared status
	Category string `qif:"L"` // L - Category (or transfer/class)

	// Split fields – repeated groups, flattened for now to first occurrence
	SplitCategory string `qif:"S"` // S - Category in split
	SplitMemo     string `qif:"E"` // E - Memo in split
	SplitAmount   string `qif:"$"` // $ - Dollar amount of split

	// RawLines contains the raw QIF lines (without trailing newline) that
	// composed this transaction, excluding the header and trailing '^'.
	RawLines []string `qif:"-"`
}

// Decoder reads QIF data from an input stream.
type Decoder struct {
	r *bufio.Reader
}

// NewDecoder returns a new QIF decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
	}
}

// Decode reads QIF data from the underlying reader and returns all parsed
// non-investment transactions. For now this is a convenience wrapper around
// a streaming decode; it reads the whole file.
func (d *Decoder) Decode() ([]*Transaction, error) {
	var (
		transactions []*Transaction
		currentType  string
	)

	for {
		line, err := d.readLine()
		if err == io.EOF {
			// No partial transaction handling – QIF files should end with '^'
			return transactions, nil
		}
		if err != nil {
			return nil, err
		}

		if len(line) == 0 {
			continue
		}

		// Header / account-type line: !Type:Cash, !Type:Bank, ...
		if strings.HasPrefix(line, "!Type:") {
			currentType = strings.TrimSpace(line[len("!Type:"):])
			continue
		}

		// A transaction must start with 'D' (date) according to the spec.
		if line[0] == 'D' {
			tx, err := d.decodeTransaction(currentType, line)
			if err != nil {
				return nil, err
			}
			transactions = append(transactions, tx)
			continue
		}

		// Lines outside of transactions are currently ignored.
	}

}

// decodeTransaction parses a single transaction, given that the first line
// (already read) is a 'D' date line. It continues reading until the '^' end
// marker has been consumed.
func (d *Decoder) decodeTransaction(txType string, firstLine string) (*Transaction, error) {
	tx := &Transaction{
		Type: txType,
	}

	assignField(tx, firstLine)

	for {
		line, err := d.readLine()
		if err != nil {
			if err == io.EOF {
				return nil, fmt.Errorf("unexpected EOF while reading transaction")
			}
			return nil, err
		}
		if len(line) == 0 {
			// empty lines inside a transaction are preserved in RawLines but
			// don't correspond to any field.
			tx.RawLines = append(tx.RawLines, line)
			continue
		}
		if line[0] == '^' {
			// end of transaction
			return tx, nil
		}

		assignField(tx, line)
	}
}

// assignField updates tx based on a single QIF field line.
// It also appends the raw line (minus trailing newline) to RawLines.
func assignField(tx *Transaction, line string) {
	if len(line) == 0 {
		return
	}
	// Store raw line
	tx.RawLines = append(tx.RawLines, line)

	prefix := line[0]
	value := line[1:]

	switch prefix {
	case 'D':
		tx.Date = value
	case 'T':
		tx.Amount = value
	case 'U':
		// Higher precision amount; if present, prefer it over T.
		tx.Amount = value
	case 'N':
		tx.Num = value
	case 'P':
		tx.Payee = value
	case 'M':
		if tx.Memo == "" {
			tx.Memo = value
		} else {
			// Multiple memo lines – concatenate with newline.
			tx.Memo += "\n" + value
		}
	case 'A':
		if tx.Addr == "" {
			tx.Addr = value
		} else {
			tx.Addr += "\n" + value
		}
	case 'C':
		tx.Cleared = value
	case 'L':
		tx.Category = value
	case 'S':
		// For now we keep only first split; real-world usage may need a slice.
		if tx.SplitCategory == "" {
			tx.SplitCategory = value
		}
	case 'E':
		if tx.SplitMemo == "" {
			tx.SplitMemo = value
		}
	case '$':
		if tx.SplitAmount == "" {
			tx.SplitAmount = value
		}
	}
}

// readLine reads a single logical line without the trailing '\n' or '\r\n'.
func (d *Decoder) readLine() (string, error) {
	line, err := d.r.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	// Trim CRLF and LF.
	line = strings.TrimRight(line, "\r\n")
	if err == io.EOF && len(line) == 0 {
		return "", io.EOF
	}
	return line, err
}

// ParseQIF is a convenience helper that parses all transactions from a QIF
// stream and returns them.
func ParseQIF(reader io.Reader) ([]*Transaction, error) {
	return NewDecoder(reader).Decode()
}
