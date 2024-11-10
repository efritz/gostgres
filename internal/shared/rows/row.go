package rows

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/fields"
)

type Row struct {
	Fields []fields.ResolvedField
	Values []any
}

func NewRow(fields []fields.ResolvedField, values []any) (_ Row, err error) {
	fields, values, err = refineTypes(fields, values)
	if err != nil {
		return Row{}, err
	}

	return Row{Fields: fields, Values: values}, nil
}

func (r Row) TID() (int64, error) {
	for i, field := range r.Fields {
		if field.IsTID() {
			return r.Values[i].(int64), nil
		}
	}

	return 0, fmt.Errorf("no tid in row")
}

func CombineRows(rows ...Row) Row {
	var fields []fields.ResolvedField
	for _, row := range rows {
		fields = append(fields, row.Fields...)
	}

	var values []any
	for _, row := range rows {
		values = append(values, row.Values...)
	}

	return Row{Fields: fields, Values: values}
}
