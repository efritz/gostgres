package rows

import (
	"github.com/efritz/gostgres/internal/shared/fields"
)

type Rows struct {
	Fields []fields.Field
	Values [][]any
}

func NewRows(fields []fields.Field) (Rows, error) {
	return NewRowsWithValues(fields, nil)
}

func NewRowsWithValues(fields []fields.Field, values [][]any) (_ Rows, err error) {
	for i, rowValues := range values {
		fields, rowValues, err = refineTypes(fields, rowValues)
		if err != nil {
			return Rows{}, err
		}

		values[i] = rowValues
	}

	return Rows{Fields: fields, Values: values}, nil
}

func (rows Rows) AddValues(values []any) (Rows, error) {
	return NewRowsWithValues(rows.Fields, append(rows.Values, values))
}

func (rows Rows) Size() int {
	return len(rows.Values)
}

func (rows Rows) Row(index int) Row {
	row, _ := NewRow(rows.Fields, rows.Values[index])
	return row
}
