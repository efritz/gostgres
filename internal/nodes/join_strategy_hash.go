package nodes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type hashJoinStrategy struct {
	n     *joinNode
	left  expressions.Expression
	right expressions.Expression
}

func (s *hashJoinStrategy) Name() string {
	return "hash"
}

func (s *hashJoinStrategy) Ordering() OrderExpression {
	return s.n.left.Ordering()
}

func (s *hashJoinStrategy) Scanner(ctx ScanContext) (Scanner, error) {
	rightScanner, err := s.n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	h := map[interface{}]shared.Row{}
	if err := VisitRows(rightScanner, func(row shared.Row) (bool, error) {
		key, err := ctx.Evaluate(s.right, row)
		if err != nil {
			return false, err
		}

		h[key] = row
		return true, nil
	}); err != nil {
		return nil, err
	}

	leftScanner, err := s.n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	return ScannerFunc(func() (shared.Row, error) {
		for {
			leftRow, err := leftScanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			key, err := ctx.Evaluate(s.left, leftRow)
			if err != nil {
				return shared.Row{}, err
			}

			rightRow, ok := h[key]
			if !ok {
				continue
			}

			return shared.NewRow(s.n.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))
		}
	}), nil
}
