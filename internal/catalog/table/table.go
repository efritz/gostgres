package table

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
	"golang.org/x/exp/slices"
)

type Table struct {
	name        string
	fields      []types.TableField
	rows        map[int64]shared.Row
	primaryKey  types.BaseIndex
	indexes     []types.BaseIndex
	constraints []types.Constraint
}

var _ types.Table = &Table{}

func NewTable(name string, fields []types.TableField) types.Table {
	tableFields := []types.TableField{
		types.NewInternalTableField(name, shared.TIDName, shared.TypeBigInteger),
	}
	for _, field := range fields {
		tableFields = append(tableFields, field.WithRelationName(name))
	}

	return &Table{
		name:   name,
		fields: tableFields,
		rows:   map[int64]shared.Row{},
	}
}

func (t *Table) Name() string {
	return t.name
}

func (t *Table) Indexes() []types.BaseIndex {
	if t.primaryKey != nil {
		return append([]types.BaseIndex{t.primaryKey}, t.indexes...)
	}

	return t.indexes
}

func (t *Table) Fields() []types.TableField {
	return slices.Clone(t.fields)
}

func (t *Table) Size() int {
	return len(t.rows)
}

func (t *Table) TIDs() []int64 {
	tids := make([]int64, 0, len(t.rows))
	for tid := range t.rows {
		tids = append(tids, tid)
	}
	slices.Sort(tids)

	return tids
}

func (t *Table) Row(tid int64) (shared.Row, bool) {
	row, ok := t.rows[tid]
	return row, ok
}

func (t *Table) SetPrimaryKey(index types.BaseIndex) error {
	if t.primaryKey != nil {
		return fmt.Errorf("primary key already set")
	}

	for _, row := range t.rows {
		if err := index.Insert(row); err != nil {
			return err
		}
	}

	t.primaryKey = index
	return nil
}

func (t *Table) AddIndex(index types.BaseIndex) error {
	for _, row := range t.rows {
		if err := index.Insert(row); err != nil {
			return err
		}
	}

	t.indexes = append(t.indexes, index)
	return nil
}

func (t *Table) AddConstraint(ctx types.Context, constraint types.Constraint) error {
	for _, row := range t.rows {
		if err := constraint.Check(ctx, row); err != nil {
			return err
		}
	}

	t.constraints = append(t.constraints, constraint)
	return nil
}

var tid = int64(0)

func (t *Table) Insert(ctx types.Context, row shared.Row) (_ shared.Row, err error) {
	tid++
	id := tid

	var fields []shared.Field
	for _, field := range t.fields {
		fields = append(fields, field.Field)
	}

	newRow, err := shared.NewRow(fields, append([]any{id}, row.Values...))
	if err != nil {
		return shared.Row{}, err
	}

	for _, constraint := range t.constraints {
		if err := constraint.Check(ctx, newRow); err != nil {
			return shared.Row{}, err
		}
	}

	for _, index := range t.Indexes() {
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
