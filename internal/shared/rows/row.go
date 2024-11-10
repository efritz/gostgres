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

func (r Row) TID() (int64, error) {
	for i, field := range r.Fields {
		if field.IsTID() {
			v, ok := r.Values[i].(int64)
			if !ok {
				panic("tid is not an int64")
			}

			return v, nil
		}
	}

	return 0, fmt.Errorf("no tid in row")
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
