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
	return nil // TODO - ordered on the left field?
}

func (s *hashJoinStrategy) Scanner() (Scanner, error) {
	// TODO - can share with grouping?
	h := map[interface{}]shared.Row{}
	rightScanner, err := s.n.right.Scanner()
	if err != nil {
		return nil, err
	}

	if err := VisitRows(rightScanner, func(row shared.Row) (bool, error) {
		key, err := s.right.ValueFrom(row)
		if err != nil {
			return false, err
		}

		h[key] = row
		return true, nil
	}); err != nil {
		return nil, err
	}

	leftScanner, err := s.n.left.Scanner()
	if err != nil {
		return nil, err
	}

	return ScannerFunc(func() (shared.Row, error) {
		for {
			leftRow, err := leftScanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			key, err := s.left.ValueFrom(leftRow)
			if err != nil {
				return shared.Row{}, err
			}

			if rightRow, ok := h[key]; ok {
				return shared.NewRow(s.n.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))
			}
		}
	}), nil
}
