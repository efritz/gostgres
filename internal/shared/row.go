package shared

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

func NewRows(Fields []Field) Rows {
	return Rows{
		Fields: Fields,
		Values: nil,
	}
}

func (rows Rows) AddValues(values []interface{}) Rows {
	return Rows{
		Fields: rows.Fields,
		Values: append(rows.Values, values),
	}
}

func (rows Rows) Row(index int) Row {
	return Row{
		Fields: rows.Fields,
		Values: rows.Values[index],
	}
}

type Field struct {
	Name         string
	RelationName string
	// TODO - value types
}
