package nodes

import (
	"github.com/efritz/gostgres/internal/shared"
)

type nestedLoopJoinStrategy struct {
	n *joinNode
}

func (s *nestedLoopJoinStrategy) Name() string {
	return "nested loop"
}

func (s *nestedLoopJoinStrategy) Ordering() OrderExpression {
	leftOrdering := s.n.left.Ordering()
	if leftOrdering == nil {
		return nil
	}

	rightOrdering := s.n.right.Ordering()
	if rightOrdering == nil {
		return leftOrdering
	}

	return NewOrderExpression(append(leftOrdering.Expressions(), rightOrdering.Expressions()...))
}

func (s *nestedLoopJoinStrategy) Scanner(ctx ScanContext) (Scanner, error) {
	leftScanner, err := s.n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var (
		leftRow      *shared.Row
		rightScanner Scanner
	)

	return ScannerFunc(func() (shared.Row, error) {
		for {
			if leftRow == nil {
				row, err := leftScanner.Scan()
				if err != nil {
					return shared.Row{}, err
				}
				leftRow = &row

				scanner, err := s.n.right.Scanner(ScanContext{
					OuterRow: row,
				})
				if err != nil {
					return shared.Row{}, nil
				}
				rightScanner = scanner
			}

			rightRow, err := rightScanner.Scan()
			if err != nil {
				if err == ErrNoRows {
					leftRow = nil
					rightScanner = nil
					continue
				}

				return shared.Row{}, err
			}

			row, err := shared.NewRow(s.n.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))
			if err != nil {
				return shared.Row{}, err
			}

			if s.n.filter != nil {
				if ok, err := shared.EnsureBool(ctx.Evaluate(s.n.filter, row)); err != nil {
					return shared.Row{}, err
				} else if !ok {
					continue
				}
			}

			return row, nil
		}
	}), nil
}
