package iif

import (
	"fmt"
	"reflect"
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	Tr     Trns  `type:"TRNS"`
	Splits []Spl `type:"SPL"`
}

type Trns struct {
	TransactionType string          `iif:"TRNSTYPE"`
	Date            time.Time       `iif:"DATE"`
	Account         string          `iif:"ACCNT"`
	Name            string          `iif:"NAME"`
	Class           string          `iif:"CLASS"`
	Amount          decimal.Decimal `iif:"AMOUNT"`
	Memo            string          `iif:"MEMO"`
}

type Spl struct {
	TransactionType string          `iif:"TRNSTYPE"`
	Date            time.Time       `iif:"DATE"`
	Account         string          `iif:"ACCNT"`
	Name            string          `iif:"NAME"`
	Class           string          `iif:"CLASS"`
	Amount          decimal.Decimal `iif:"AMOUNT"`
	Memo            string          `iif:"MEMO"`
}

func DeserializeTransactions(b Block) ([]Transaction, error) {
	var out []Transaction

	for _, recGroup := range b.Records {
		if len(recGroup) == 0 {
			continue
		}

		var tx Transaction
		if err := DeserializeRecordGroup(&tx, recGroup); err != nil {
			return nil, err
		}
		out = append(out, tx)
	}

	return out, nil
}

func DeserializeRecordGroup(tx any, recs []Record) error {
	for _, r := range recs {
		if err := applyRecord(tx, r); err != nil {
			return err
		}
	}
	return nil
}

func applyRecord(tx any, r Record) error {
	txVal := reflect.ValueOf(tx).Elem()
	txType := txVal.Type()

	for i := 0; i < txType.NumField(); i++ {
		field := txType.Field(i)
		tag := field.Tag.Get("type")
		if tag == "" || string(r.Type) != tag {
			continue
		}

		fv := txVal.Field(i)

		if fv.Kind() == reflect.Slice {
			elemType := fv.Type().Elem()
			elemPtr := reflect.New(elemType).Elem()

			if err := populateStructFromRecord(elemPtr, r); err != nil {
				return err
			}

			fv.Set(reflect.Append(fv, elemPtr))
			return nil
		}
		if fv.Kind() == reflect.Struct {
			if err := populateStructFromRecord(fv, r); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func populateStructFromRecord(v reflect.Value, r Record) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("populateStructFromRecord: expected struct, got %s", v.Kind())
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag := sf.Tag.Get("iif")
		if tag == "" {
			continue
		}

		raw, ok := r.Fields[tag]
		if !ok {
			continue
		}

		fv := v.Field(i)
		if !fv.CanSet() {
			continue
		}

		if err := setFieldValueFromString(fv, raw); err != nil {
			return fmt.Errorf("field %s: %w", sf.Name, err)
		}
	}

	return nil
}

// setFieldValueFromString converts the string representation from a Record
// into the appropriate Go type and assigns it to fv.
func setFieldValueFromString(fv reflect.Value, s string) error {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(s)
		return nil
	case reflect.Struct:
		// Handle known struct types (time.Time, decimal.Decimal, etc.)
		switch fv.Type() {
		case reflect.TypeOf(time.Time{}):
			// IIF date formats can vary; here we assume the standard
			// QuickBooks IIF date "MM/DD/YYYY". Adjust if needed.
			if s == "" {
				return nil
			}
			t, err := time.Parse("1/2/2006", s)
			if err != nil {
				return err
			}
			fv.Set(reflect.ValueOf(t))
			return nil
		case reflect.TypeOf(decimal.Decimal{}):
			if s == "" {
				fv.Set(reflect.ValueOf(decimal.Zero))
				return nil
			}
			d, err := decimal.NewFromString(s)
			if err != nil {
				return err
			}
			fv.Set(reflect.ValueOf(d))
			return nil
		default:
			return fmt.Errorf("unsupported struct type %s", fv.Type())
		}
	default:
		return fmt.Errorf("unsupported kind %s", fv.Kind())
	}
}
