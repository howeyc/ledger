package ledger

import (
	"bufio"
	"io"
	"os"
	"unsafe"
)

type linescanner struct {
	scanner *bufio.Scanner
	unsafe  bool

	filename  string
	lineCount int
}

// NewLineScanner creates a wrapper around bufio.Scanner with pre-allocated
// buffer. Significantly reduces memory allocations and reduces runtime.
func newLineScanner(filename string, r io.Reader) *linescanner {
	lp := &linescanner{}
	lp.scanner = bufio.NewScanner(r)
	if fs, fserr := os.Stat(filename); fserr == nil {
		lp.scanner.Buffer(make([]byte, int(fs.Size())), int(fs.Size()))
		lp.unsafe = true
	}
	lp.filename = filename

	return lp
}

func (lp *linescanner) Scan() bool {
	return lp.scanner.Scan()
}

func (lp *linescanner) Text() string {
	var line string
	if lp.unsafe {
		if lbytes := lp.scanner.Bytes(); len(lbytes) > 0 {
			line = unsafe.String(unsafe.SliceData(lbytes), len(lbytes))
		} else {
			line = ""
		}
	} else {
		line = lp.scanner.Text()
	}
	lp.lineCount++
	return line
}

func (lp *linescanner) LineNumber() int {
	return lp.lineCount
}

func (lp *linescanner) Name() string {
	return lp.filename
}
