package shared

import "fmt"

type Row struct {
	Fields []Field
	Values []interface{}
}

func NewRow(fields []Field, values []interface{}) Row {
	return Row{
		Fields: fields,
		Values: values,
	}
}

type Rows struct {
	Fields []Field
	Values [][]interface{}
}

func NewRows(fields []Field) Rows {
	return NewRowsWithValues(fields, nil)
}

func NewRowsWithValues(fields []Field, values [][]interface{}) Rows {
	return Rows{
		Fields: fields,
		Values: values,
	}
}

func (rows Rows) AddValues(values []interface{}) (Rows, error) {
	if len(rows.Fields) != len(values) {
		// TODO - check for field types, ordering
		return rows, fmt.Errorf("unexpected number of columns")
	}

	return Rows{
		Fields: rows.Fields,
		Values: append(rows.Values, values),
	}, nil
}

func (rows Rows) Row(index int) Row {
	return Row{
		Fields: rows.Fields,
		Values: rows.Values[index],
	}
}
