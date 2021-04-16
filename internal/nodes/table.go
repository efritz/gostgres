package nodes

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Table struct {
	rows shared.Rows
}

func NewTable(rows shared.Rows) *Table {
	return &Table{
		rows: rows,
	}
}

func (t *Table) Fields() []shared.Field {
	return copyFields(t.rows.Fields)
}

func (t *Table) Insert(row shared.Row) (err error) {
	t.rows, err = t.rows.AddValues(row.Values)
	return err
}
