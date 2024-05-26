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

const TIDName = "tid"

func (r Row) TID() (int64, error) {
	if len(r.Fields) == 0 || r.Fields[0].Name() != TIDName {
		return 0, nil
	}

	tid, ok := r.Values[0].(int64)
	if !ok {
		return 0, fmt.Errorf("no tid in row")
	}

	return tid, nil
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
