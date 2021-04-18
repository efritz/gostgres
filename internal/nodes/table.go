package nodes

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Table struct {
	name string
	rows shared.Rows
}

func NewTable(name string, rows shared.Rows) (*Table, error) {
	fields := append([]shared.Field{
		shared.NewField(name, "tid", shared.TypeKindNumeric, true),
	}, rows.Fields...)

	newRows, err := shared.NewRows(fields)
	if err != nil {
		return nil, err
	}

	table := &Table{
		name: name,
		rows: newRows,
	}
	for i := range rows.Values {
		if _, err := table.Insert(rows.Row(i)); err != nil {
			return nil, err
		}
	}

	return table, nil
}

func (t *Table) Fields() []shared.Field {
	return copyFields(t.rows.Fields)
}

var tid = 0

func (t *Table) Insert(row shared.Row) (_ shared.Row, err error) {
	tid++
	t.rows, err = t.rows.AddValues(append([]interface{}{tid}, row.Values...))
	if err != nil {
		return shared.Row{}, err
	}

	return t.rows.Row(len(t.rows.Values) - 1), nil
}
