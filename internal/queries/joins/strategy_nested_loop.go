package joins

import (
	"slices"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type nestedLoopJoinStrategy struct {
	n *joinNode
}

func (s *nestedLoopJoinStrategy) Name() string {
	return "nested loop"
}

func (s *nestedLoopJoinStrategy) Ordering() expressions.OrderExpression {
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

func (s *nestedLoopJoinStrategy) Scanner(ctx queries.Context) (scan.Scanner, error) {
	leftScanner, err := s.n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var (
		leftRow      *shared.Row
		rightScanner scan.Scanner
	)

	return scan.ScannerFunc(func() (shared.Row, error) {
		for {
			if leftRow == nil {
				row, err := leftScanner.Scan()
				if err != nil {
					return shared.Row{}, err
				}
				leftRow = &row

				scanner, err := s.n.right.Scanner(ctx.WithOuterRow(row))
				if err != nil {
					return shared.Row{}, nil
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

				return shared.Row{}, err
			}

			row, err := shared.NewRow(s.n.Fields(), append(slices.Clone(leftRow.Values), rightRow.Values...))
			if err != nil {
				return shared.Row{}, err
			}

			if s.n.filter != nil {
				if ok, err := shared.ValueAs[bool](queries.Evaluate(ctx, s.n.filter, row)); err != nil {
					return shared.Row{}, err
				} else if ok == nil || !*ok {
					continue
				}
			}

			return row, nil
		}
	}), nil
}
