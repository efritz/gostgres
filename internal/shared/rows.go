package shared

import "fmt"

type Row struct {
	Fields []Field
	Values []interface{}
}

func NewRow(fields []Field, values []interface{}) (_ Row, err error) {
	fields, err = refineFieldTypes(fields, values)
	if err != nil {
		return Row{}, err
	}

	return Row{Fields: fields, Values: values}, nil
}

func CombineRows(rows ...Row) Row {
	var fields []Field
	for _, row := range rows {
		fields = append(fields, row.Fields...)
	}

	var values []interface{}
	for _, row := range rows {
		values = append(values, row.Values...)
	}

	return Row{Fields: fields, Values: values}
}

type Rows struct {
	Fields []Field
	Values [][]interface{}
}

func NewRows(fields []Field) (Rows, error) {
	return NewRowsWithValues(fields, nil)
}

func NewRowsWithValues(fields []Field, values [][]interface{}) (_ Rows, err error) {
	for _, values := range values {
		fields, err = refineFieldTypes(fields, values)
		if err != nil {
			return Rows{}, err
		}
	}

	return Rows{Fields: fields, Values: values}, nil
}

func (rows Rows) AddValues(values []interface{}) (Rows, error) {
	return NewRowsWithValues(rows.Fields, append(rows.Values, values))
}

func (rows Rows) Size() int {
	return len(rows.Values)
}

func (rows Rows) Row(index int) Row {
	row, _ := NewRow(rows.Fields, rows.Values[index])
	return row
}

func refineFieldTypes(fields []Field, values []interface{}) ([]Field, error) {
	if len(fields) != len(values) {
		return nil, fmt.Errorf("unexpected number of columns")
	}

	refined := make([]Field, 0, len(fields))
	for i, field := range fields {
		refinedType := refineType(field.TypeKind, values[i])
		if refinedType == TypeKindInvalid {
			return nil, fmt.Errorf("type error (%v is not %s)", values[i], field.TypeKind)
		}

		refined = append(refined, NewField(field.RelationName, field.Name, refinedType, field.Internal))
	}

	return refined, nil
}
