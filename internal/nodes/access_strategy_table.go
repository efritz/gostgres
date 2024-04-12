package nodes

import (
	"fmt"
	"io"
	"sort"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type tableAccessStrategy struct {
	table *Table
}

var _ accessStrategy = &tableAccessStrategy{}

func NewTableAccessStrategy(table *Table) accessStrategy {
	return &tableAccessStrategy{table: table}
}

func (s *tableAccessStrategy) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%stable scan of %s\n", indent(indentationLevel), s.table.name))
}

func (s *tableAccessStrategy) Filter() expressions.Expression {
	return nil
}

func (s *tableAccessStrategy) Ordering() OrderExpression {
	return nil
}

func (s *tableAccessStrategy) Scanner() (Scanner, error) {
	tids := make([]int, 0, len(s.table.rows))
	for tid := range s.table.rows {
		tids = append(tids, tid)
	}
	sort.Ints(tids)

	rows, err := shared.NewRows(s.table.Fields())
	if err != nil {
		return nil, err
	}
	for _, tid := range tids {
		rows, err = rows.AddValues(s.table.rows[tid].Values)
		if err != nil {
			return nil, err
		}
	}

	i := 0

	return ScannerFunc(func() (shared.Row, error) {
		for i < rows.Size() {
			row := rows.Row(i)
			i++
			return row, nil
		}

		return shared.Row{}, ErrNoRows
	}), nil
}
