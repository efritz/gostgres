package relations

import (
	"fmt"

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

func (r *Table) Fields() []shared.Field {
	return copyFields(r.rows.Fields)
}

func (t *Table) Insert(row shared.Row) error {
	if len(t.rows.Fields) != len(row.Fields) {
		// TODO - check for field types, ordering
		return fmt.Errorf("unexpected number of columns")
	}

	t.rows = t.rows.AddValues(row.Values)
	return nil
}
