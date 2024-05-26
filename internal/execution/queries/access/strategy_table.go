package access

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/engine/serialization"
	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
)

type tableAccessStrategy struct {
	table impls.Table
}

var _ accessStrategy = &tableAccessStrategy{}

func NewTableAccessStrategy(table impls.Table) accessStrategy {
	return &tableAccessStrategy{table: table}
}

func (s *tableAccessStrategy) Serialize(w serialization.IndentWriter) {
	w.WritefLine("table scan of %s", s.table.Name())
}

func (s *tableAccessStrategy) Filter() impls.Expression {
	return nil
}

func (s *tableAccessStrategy) Ordering() impls.OrderExpression {
	return nil
}

func (s *tableAccessStrategy) Scanner(ctx impls.Context) (scan.Scanner, error) {
	tids := s.table.TIDs()

	i := 0

	return scan.ScannerFunc(func() (rows.Row, error) {
		if i >= len(tids) {
			return rows.Row{}, scan.ErrNoRows
		}

		tid := tids[i]
		i++

		row, ok := s.table.Row(tid)
		if !ok {
			return rows.Row{}, fmt.Errorf("missing row %d", tid)
		}

		return row, nil
	}), nil
}
