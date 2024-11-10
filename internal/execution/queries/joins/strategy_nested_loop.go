package joins

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/expressions"
	"github.com/efritz/gostgres/internal/execution/queries"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/types"
)

type nestedLoopJoinStrategy struct {
	n *joinNode
}

func (s *nestedLoopJoinStrategy) Name() string {
	return "nested loop"
}

func (s *nestedLoopJoinStrategy) Ordering() impls.OrderExpression {
	leftOrdering := s.n.left.Ordering()
	if leftOrdering == nil {
		return nil
	}

	rightOrdering := s.n.right.Ordering()
	if rightOrdering == nil {
		return leftOrdering
	}

	return expressions.NewOrderExpression(append(leftOrdering.Expressions(), rightOrdering.Expressions()...))
}

func (s *nestedLoopJoinStrategy) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Nested Loop Join Strategy scanner")

	leftScanner, err := s.n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var (
		leftRow      *rows.Row
		rightScanner scan.RowScanner
	)

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Nested Loop Join Strategy")

		for {
			if leftRow == nil {
				row, err := leftScanner.Scan()
				if err != nil {
					return rows.Row{}, err
				}
				leftRow = &row

				scanner, err := s.n.right.Scanner(ctx.WithOuterRow(row))
				if err != nil {
					return rows.Row{}, nil
				}
				rightScanner = scanner
			}

			rightRow, err := rightScanner.Scan()
			if err != nil {
				if err == scan.ErrNoRows {
					leftRow = nil
					rightScanner = nil
					continue
				}

				return rows.Row{}, err
			}

			row, err := rows.NewRow(s.n.Fields(), append(slices.Clone(leftRow.Values), rightRow.Values...))
			if err != nil {
				return rows.Row{}, err
			}

			if s.n.filter != nil {
				if ok, err := types.ValueAs[bool](queries.Evaluate(ctx, s.n.filter, row)); err != nil {
					return rows.Row{}, err
				} else if ok == nil || !*ok {
					continue
				}
			}

			return row, nil
		}
	}), nil
}
