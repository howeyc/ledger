package ledger

import (
	"bytes"
	"io"
	"os"
)

// NewLedgerReader reads a file and includes any files with include directives
// and returns the whole combined ledger as a buffer for parsing.
//
// Deprecated: use ParseLedgerFile
func NewLedgerReader(filename string) (io.Reader, error) {
	var buf bytes.Buffer

	ifile, ierr := os.Open(filename)
	if ierr != nil {
		return &buf, ierr
	}
	io.Copy(&buf, ifile)
	ifile.Close()

	return &buf, nil
}
