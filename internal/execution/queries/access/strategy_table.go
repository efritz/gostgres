package access

import (
	"fmt"

	"github.com/efritz/gostgres/internal/execution/scan"
	"github.com/efritz/gostgres/internal/serialization"
	"github.com/efritz/gostgres/internal/shared"
	"github.com/efritz/gostgres/internal/types"
)

type tableAccessStrategy struct {
	table types.Table
}

var _ accessStrategy = &tableAccessStrategy{}

func NewTableAccessStrategy(table types.Table) accessStrategy {
	return &tableAccessStrategy{table: table}
}

func (s *tableAccessStrategy) Serialize(w serialization.IndentWriter) {
	w.WritefLine("table scan of %s", s.table.Name())
}

func (s *tableAccessStrategy) Filter() types.Expression {
	return nil
}

func (s *tableAccessStrategy) Ordering() types.OrderExpression {
	return nil
}

func (s *tableAccessStrategy) Scanner(ctx types.Context) (scan.Scanner, error) {
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
