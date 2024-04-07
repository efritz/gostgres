package nodes

import (
	"github.com/efritz/gostgres/internal/shared"
)

type Table struct {
	name string
	rows shared.Rows

	// TODO
	// indexes []Index
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

// TODO - should be delete followed by insert
func (t *Table) Update(row shared.Row) (bool, error) {
	mergedRow, err := shared.NewRow(t.rows.Fields, row.Values)
	if err != nil {
		return false, err
	}

	for i, values := range t.rows.Values {
		if values[0] != row.Values[0] {
			continue
		}

		// t.Delete(t.rows.Row(i))
		// t.Insert(mergedRow)
		t.rows.Fields = mergedRow.Fields
		t.rows.Values[i] = mergedRow.Values
		return true, nil
	}

	return false, nil
}

func (t *Table) Delete(row shared.Row) (shared.Row, bool, error) {
	for i, values := range t.rows.Values {
		if values[0] != row.Values[0] {
			continue
		}

		copy(t.rows.Values[i:], t.rows.Values[i+1:])
		t.rows.Values = t.rows.Values[:len(t.rows.Values)-1]
		row, err := shared.NewRow(t.rows.Fields, values)
		return row, true, err
	}

	return shared.Row{}, false, nil
}
