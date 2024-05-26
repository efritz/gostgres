package table

import (
	"fmt"

	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/types"
	"golang.org/x/exp/slices"
)

type table struct {
	name        string
	fields      []impls.TableField
	rows        map[int64]rows.Row
	primaryKey  impls.BaseIndex
	indexes     []impls.BaseIndex
	constraints []impls.Constraint
}

var _ impls.Table = &table{}

func NewTable(name string, fields []impls.TableField) impls.Table {
	tableFields := []impls.TableField{
		impls.NewInternalTableField(name, rows.TIDName, types.TypeBigInteger),
	}
	for _, field := range fields {
		tableFields = append(tableFields, field.WithRelationName(name))
	}

	return &table{
		name:   name,
		fields: tableFields,
		rows:   map[int64]rows.Row{},
	}
}

func (t *table) Name() string {
	return t.name
}

func (t *table) Indexes() []impls.BaseIndex {
	if t.primaryKey != nil {
		return append([]impls.BaseIndex{t.primaryKey}, t.indexes...)
	}

	return t.indexes
}

func (t *table) Fields() []impls.TableField {
	return slices.Clone(t.fields)
}

func (t *table) Size() int {
	return len(t.rows)
}

func (t *table) TIDs() []int64 {
	tids := make([]int64, 0, len(t.rows))
	for tid := range t.rows {
		tids = append(tids, tid)
	}
	slices.Sort(tids)

	return tids
}

func (t *table) Row(tid int64) (rows.Row, bool) {
	row, ok := t.rows[tid]
	return row, ok
}

func (t *table) SetPrimaryKey(index impls.BaseIndex) error {
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

func (t *table) AddIndex(index impls.BaseIndex) error {
	for _, row := range t.rows {
		if err := index.Insert(row); err != nil {
			return err
		}
	}

	t.indexes = append(t.indexes, index)
	return nil
}

func (t *table) AddConstraint(ctx impls.Context, constraint impls.Constraint) error {
	for _, row := range t.rows {
		if err := constraint.Check(ctx, row); err != nil {
			return err
		}
	}

	t.constraints = append(t.constraints, constraint)
	return nil
}

var tid = int64(0)

func (t *table) Insert(ctx impls.Context, row rows.Row) (_ rows.Row, err error) {
	tid++
	id := tid

	var fields []fields.Field
	for _, field := range t.fields {
		fields = append(fields, field.Field)
	}

	newRow, err := rows.NewRow(fields, append([]any{id}, row.Values...))
	if err != nil {
		return rows.Row{}, err
	}

	for _, constraint := range t.constraints {
		if err := constraint.Check(ctx, newRow); err != nil {
			return rows.Row{}, err
		}
	}

	for _, index := range t.Indexes() {
		if err := index.Insert(newRow); err != nil {
			return rows.Row{}, err
		}
	}

	t.rows[id] = newRow
	return newRow, nil
}

func (t *table) Delete(row rows.Row) (rows.Row, bool, error) {
	tid, err := row.TID()
	if err != nil {
		return rows.Row{}, false, err
	}

	fullRow, ok := t.rows[tid]
	if !ok {
		return rows.Row{}, false, nil
	}

	for _, index := range t.indexes {
		if err := index.Delete(fullRow); err != nil {
			return rows.Row{}, false, err
		}
	}

	delete(t.rows, tid)
	return fullRow, true, nil
}
