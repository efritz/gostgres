package access

import (
	"fmt"

	"github.com/efritz/gostgres/internal/catalog/table"
	"github.com/efritz/gostgres/internal/execution"
	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
)

type tableAccessStrategy struct {
	table *table.Table
}

var _ accessStrategy = &tableAccessStrategy{}

func NewTableAccessStrategy(table *table.Table) accessStrategy {
	return &tableAccessStrategy{table: table}
}

func (s *tableAccessStrategy) Serialize(w serialization.IndentWriter) {
	w.WritefLine("table scan of %s", s.table.Name())
}

func (s *tableAccessStrategy) Filter() expressions.Expression {
	return nil
}

func (s *tableAccessStrategy) Ordering() expressions.OrderExpression {
	return nil
}

func (s *tableAccessStrategy) Scanner(ctx execution.Context) (scan.Scanner, error) {
	tids := s.table.TIDs()

	i := 0

	return scan.ScannerFunc(func() (shared.Row, error) {
		if i >= len(tids) {
			return shared.Row{}, scan.ErrNoRows
		}

		tid := tids[i]
		i++

		row, ok := s.table.Row(tid)
		if !ok {
			return shared.Row{}, fmt.Errorf("missing row %d", tid)
		}

		return row, nil
	}), nil
}
