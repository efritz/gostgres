package access

import (
	"fmt"
	"io"
	"strings"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/table"
)

type tableAccessStrategy struct {
	table *table.Table
}

var _ accessStrategy = &tableAccessStrategy{}

func NewTableAccessStrategy(table *table.Table) accessStrategy {
	return &tableAccessStrategy{table: table}
}

func (s *tableAccessStrategy) Serialize(w io.Writer, indentationLevel int) {
	io.WriteString(w, fmt.Sprintf("%stable scan of %s\n", indent(indentationLevel), s.table.Name()))
}

func (s *tableAccessStrategy) Filter() expressions.Expression {
	return nil
}

func (s *tableAccessStrategy) Ordering() expressions.OrderExpression {
	return nil
}

func (s *tableAccessStrategy) Scanner(ctx scan.ScanContext) (scan.Scanner, error) {
	tids := s.table.TIDs()

	rows, err := shared.NewRows(s.table.Fields())
	if err != nil {
		return nil, err
	}
	for _, tid := range tids {
		row, ok := s.table.Row(tid)
		if !ok {
			return nil, fmt.Errorf("missing row %d", tid)
		}

		rows, err = rows.AddValues(row.Values)
		if err != nil {
			return nil, err
		}
	}

	i := 0

	return scan.ScannerFunc(func() (shared.Row, error) {
		for i < rows.Size() {
			row := rows.Row(i)
			i++
			return row, nil
		}

		return shared.Row{}, scan.ErrNoRows
	}), nil
}

// TODO - deduplicate
func indent(level int) string {
	return strings.Repeat(" ", level*4)
}
