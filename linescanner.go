package ledger

import (
	"bufio"
	"io"
)

type linescanner struct {
	scanner   *bufio.Scanner
	filename  string
	lineCount int
}

// NewLineScanner creates a wrapper around bufio.Scanner with pre-allocated
// buffer. Significantly reduces memory allocations and reduces runtime.
func newLineScanner(filename string, r io.Reader) *linescanner {
	lp := &linescanner{}
	lp.scanner = bufio.NewScanner(r)
	lp.filename = filename

	return lp
}

func (lp *linescanner) Scan() bool {
	return lp.scanner.Scan()
}

func (lp *linescanner) Text() string {
	var line string
	line = lp.scanner.Text()
	lp.lineCount++
	return line
}

func (lp *linescanner) LineNumber() int {
	return lp.lineCount
}

func (lp *linescanner) Name() string {
	return lp.filename
}
