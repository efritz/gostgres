package table

import (
	"sort"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
	"golang.org/x/exp/slices"
)

type Table struct {
	name    string
	fields  []shared.Field
	rows    map[int]shared.Row
	indexes []Index
}

type Index interface {
	Unwrap() Index
	Filter() expressions.Expression
	Insert(row shared.Row) error
	Delete(row shared.Row) error
}

func NewTable(name string, fields []shared.Field) *Table {
	tableFields := []shared.Field{
		shared.NewInternalField(name, shared.TIDName, shared.TypeNumeric),
	}
	for _, field := range fields {
		tableFields = append(tableFields, field.WithRelationName(name))
	}

	return &Table{
		name:   name,
		fields: tableFields,
		rows:   map[int]shared.Row{},
	}
}

func (t *Table) Name() string {
	return t.name
}

func (t *Table) Indexes() []Index {
	return t.indexes
}

func (t *Table) Fields() []shared.Field {
	return slices.Clone(t.fields)
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

func (t *Table) AddIndex(index Index) error {
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
	tid, err := row.TID()
	if err != nil {
		return shared.Row{}, false, err
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