package shared

type Rows struct {
	Fields []Field
	Values [][]any
}

func NewRows(fields []Field) (Rows, error) {
	return NewRowsWithValues(fields, nil)
}

func NewRowsWithValues(fields []Field, values [][]any) (_ Rows, err error) {
	for _, values := range values {
		fields, err = refineFieldTypes(fields, values)
		if err != nil {
			return Rows{}, err
		}
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
