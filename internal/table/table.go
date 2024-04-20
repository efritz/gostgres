package table

import (
	"fmt"
	"sort"

	"github.com/efritz/gostgres/internal/indexes"
	"github.com/efritz/gostgres/internal/shared"
)

type Table struct {
	name    string
	fields  []shared.Field
	rows    map[int]shared.Row
	indexes []indexes.BaseIndex
}

func NewTable(name string, fields []shared.Field) *Table {
	tidField := shared.NewField(name, "tid", shared.TypeKindNumeric, true)

	return &Table{
		name:   name,
		fields: append([]shared.Field{tidField}, fields...),
		rows:   map[int]shared.Row{},
	}
}

func (t *Table) Name() string {
	return t.name
}

func (t *Table) Indexes() []indexes.BaseIndex {
	return t.indexes
}

func (t *Table) Fields() []shared.Field {
	return copyFields(t.fields)
}

func (t *Table) Size() int {
	return len(t.rows)
}

func (t *Table) TIDs() []int {
	tids := make([]int, 0, len(t.rows))
	for tid := range t.rows {
		tids = append(tids, tid)
	}
	sort.Ints(tids)

	return tids
}

func (t *Table) Row(tid int) (shared.Row, bool) {
	row, ok := t.rows[tid]
	return row, ok
}

func (t *Table) AddIndex(index indexes.BaseIndex) error {
	for _, row := range t.rows {
		if err := index.Insert(row); err != nil {
			return err
		}
	}

	t.indexes = append(t.indexes, index)
	return nil
}

var tid = 0

func (t *Table) Insert(row shared.Row) (_ shared.Row, err error) {
	tid++
	id := tid

	newRow, err := shared.NewRow(t.fields, append([]any{id}, row.Values...))
	if err != nil {
		return shared.Row{}, err
	}

	for _, index := range t.indexes {
		if err := index.Insert(newRow); err != nil {
			return shared.Row{}, err
		}
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

	for _, index := range t.indexes {
		if err := index.Delete(fullRow); err != nil {
			return shared.Row{}, false, err
		}
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

// TODO - deduplicate

func copyFields(fields []shared.Field) []shared.Field {
	c := make([]shared.Field, len(fields))
	copy(c, fields)
	return c
}
