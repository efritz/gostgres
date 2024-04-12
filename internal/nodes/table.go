package nodes

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
)

type Table struct {
	name   string
	fields []shared.Field
	rows   map[int]shared.Row

	// TODO
	// indexes []Index
}

func NewTable(name string, fields []shared.Field) *Table {
	tidField := shared.NewField(name, "tid", shared.TypeKindNumeric, true)

	return &Table{
		name:   name,
		fields: append([]shared.Field{tidField}, fields...),
		rows:   map[int]shared.Row{},
	}
}

func (t *Table) Fields() []shared.Field {
	return copyFields(t.fields)
}

var tid = 0

func (t *Table) Insert(row shared.Row) (_ shared.Row, err error) {
	tid++
	id := tid

	newRow, err := shared.NewRow(t.fields, append([]interface{}{id}, row.Values...))
	if err != nil {
		return shared.Row{}, err
	}

	t.rows[id] = newRow
	return newRow, nil
}

func (t *Table) Delete(row shared.Row) (shared.Row, bool, error) {
	tid, ok := extractTID(row)
	if !ok {
		return shared.Row{}, false, fmt.Errorf("no tid in row")
	}

	fullRow, ok := t.rows[tid]
	if !ok {
		return shared.Row{}, false, nil
	}

	delete(t.rows, tid)
	return fullRow, true, nil
}

func extractTID(row shared.Row) (int, bool) {
	if len(row.Fields) == 0 || row.Fields[0].Name != "tid" {
		return 0, false
	}

	tid, ok := row.Values[0].(int)
	return tid, ok
}
