package shared

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
