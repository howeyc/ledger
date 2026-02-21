package iif

import (
	"encoding/csv"
	"errors"
	"io"
	"strings"
)

var (
	ErrInvalidHeaderLine     = errors.New("iif: invalid header line")
	ErrMismatchedColumns     = errors.New("iif: mismatched number of columns")
	ErrMismatchedRecords     = errors.New("iif: row does not match expected header")
	ErrUnknownRecordType     = errors.New("iif: unknown record type")
	ErrUnexpectedSectionType = errors.New("iif: unexpected record type for current section")
	ErrEmptyHeader           = errors.New("iif: empty header")
)

type RecordType string

type Header struct {
	Type   RecordType
	Fields []string
}

type Record struct {
	Type   RecordType
	Fields map[string]string
}

type Block struct {
	Records [][]Record
	Headers []Header
}

type File struct {
	Blocks []Block
}

type Decoder struct {
	r        *csv.Reader
	err      error
	IsHeader bool
	Type     RecordType
	Fields   []string
}

func NewDecoder(r io.Reader) *Decoder {
	reader := csv.NewReader(r)
	reader.Comma = '\t'
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = false
	reader.FieldsPerRecord = -1
	d := Decoder{r: reader}
	d.Next()
	return &d
}

func (d *Decoder) Next() {
	line, err := d.r.Read()
	d.err = err
	if err == nil {
		d.IsHeader = strings.HasPrefix(line[0], "!")
		if d.IsHeader {
			d.Type = RecordType(line[0][1:])
		} else {
			d.Type = RecordType(line[0])
		}
		d.Fields = line[1:]
	}
}

func (d *Decoder) Error() error {
	if d.err != io.EOF {
		return d.err
	}
	return nil
}

func (d *Decoder) Done() bool {
	return d.err != nil
}

func (f *File) Load(d *Decoder) error {
	for !d.Done() {
		if d.Error() != nil {
			return d.Error()
		}
		b := Block{}
		err := b.Load(d)
		if err != nil {
			return err
		}
		f.Blocks = append(f.Blocks, b)
	}
	return nil
}

func (h Header) MapFields(fields []string) map[string]string {
	m := make(map[string]string, len(fields))
	for i, f := range h.Fields {
		if i >= len(fields) {
			break
		}
		m[f] = fields[i]
	}
	return m
}

func (b *Block) Load(d *Decoder) error {
	if d.Done() {
		return d.Error()
	}
	// Parse Headers
	for !d.Done() && d.IsHeader {
		b.Headers = append(
			b.Headers,
			Header{
				Type:   RecordType(d.Type),
				Fields: trimLine(d.Fields),
			},
		)
		d.Next()
	}
	if d.Error() != nil {
		return d.Error()
	}

	// Parse Records
	for !d.Done() && !d.IsHeader {
		r := []Record{}
		// At least one record per header
		if len(b.Headers) == 0 {
			return ErrEmptyHeader
		}
		for _, h := range b.Headers {
			if d.Done() {
				return d.Error()
			}
			if d.Done() || d.Type != h.Type {
				return ErrMismatchedRecords
			}

			for !d.Done() && !d.IsHeader && d.Type == h.Type {
				r = append(r, Record{
					Type:   d.Type,
					Fields: h.MapFields(d.Fields),
				})
				d.Next()
			}
			if len(r) == 0 {
				return ErrMismatchedRecords
			}
		}
		b.Records = append(b.Records, r)
	}
	return nil
}

func trimLine(records []string) []string {
	for i, r := range records {
		if r == "" {
			return records[:i]
		}
	}
	return records
}

func (d *Decoder) Decode() (*File, error) {
	f := File{}
	err := f.Load(d)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return &f, nil
}

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

func (e *Encoder) Encode(f *File) error {
	return errors.New("iif encoding not implemented")
}
