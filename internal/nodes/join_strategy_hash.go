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

func (s *hashJoinStrategy) Scan(visitor VisitorFunc) error {
	h := map[interface{}]shared.Row{}

	if err := s.n.right.Scan(func(row shared.Row) (bool, error) {
		key, err := s.right.ValueFrom(row)
		if err != nil {
			return false, err
		}

		h[key] = row
		return true, nil
	}); err != nil {
		return err
	}

	return s.n.left.Scan(func(row shared.Row) (bool, error) {
		key, err := s.left.ValueFrom(row)
		if err != nil {
			return false, err
		}

		if rightRow, ok := h[key]; ok {
			row, err := shared.NewRow(s.n.Fields(), append(copyValues(row.Values), rightRow.Values...))
			if err != nil {
				return false, err
			}

			return visitor(row)
		}

		return true, nil
	})
}
