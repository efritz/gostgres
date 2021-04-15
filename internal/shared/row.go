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
	RelationName string
	Name         string
	// TODO - value types
}

func FindMatchingFieldIndex(needle Field, haystack []Field) (int, error) {
	unqualifiedIndexes := make([]int, 0, len(haystack))
	for index, field := range haystack {
		if field.Name != needle.Name {
			continue
		}

		if field.RelationName == needle.RelationName {
			return index, nil
		}

		unqualifiedIndexes = append(unqualifiedIndexes, index)
	}

	if needle.RelationName == "" {
		if len(unqualifiedIndexes) == 1 {
			return unqualifiedIndexes[0], nil
		}
		if len(unqualifiedIndexes) > 1 {
			return 0, fmt.Errorf("ambiguous field %s", needle.Name)
		}
	}

	return 0, fmt.Errorf("unknown field %s", needle.Name)
}
