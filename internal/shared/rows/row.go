package rows

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/fields"
)

type Row struct {
	Fields []fields.Field
	Values []any
}

func NewRow(fields []fields.Field, values []any) (_ Row, err error) {
	fields, values, err = refineTypes(fields, values)
	if err != nil {
		return Row{}, err
	}

	return Row{Fields: fields, Values: values}, nil
}

func (r Row) TID() (tid int64, _ error) {
	for i, field := range r.Fields {
		if !field.IsTID() {
			continue
		}

		if tid != 0 {
			return 0, fmt.Errorf("ambiguous tid in row")
		}

		var ok bool
		tid, ok = r.Values[i].(int64)
		if !ok {
			panic("tid is not an int64")
		}
	}

	if tid == 0 {
		return 0, fmt.Errorf("no tid in row")
	}

	return tid, nil
}

func (r Row) DropInternalFields() Row {
	var fields []fields.Field
	var values []any

	for i, field := range r.Fields {
		if field.Internal() {
			continue
		}

		fields = append(fields, field)
		values = append(values, r.Values[i])
	}

	return Row{Fields: fields, Values: values}
}

func (r Row) IsolateTID(relationName string) (Row, error) {
	var fields []fields.Field
	var values []any

	for i, field := range r.Fields {
		if !field.IsTID() {
			continue
		}

		fields = append(fields, field)
		values = append(values, r.Values[i])
	}

	return Row{Fields: fields, Values: values}, nil
}

func CombineRows(rows ...Row) Row {
	var fields []fields.Field
	for _, row := range rows {
		fields = append(fields, row.Fields...)
	}

	var values []any
	for _, row := range rows {
		values = append(values, row.Values...)
	}

	return Row{Fields: fields, Values: values}
}
