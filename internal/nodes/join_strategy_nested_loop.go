package nodes

import "github.com/efritz/gostgres/internal/shared"

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

func (s *nestedLoopJoinStrategy) Scan(visitor VisitorFunc) error {
	return s.n.left.Scan(func(leftRow shared.Row) (bool, error) {
		return true, s.n.right.Scan(func(rightRow shared.Row) (bool, error) {
			row, err := shared.NewRow(s.n.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))
			if err != nil {
				return false, err
			}

			if s.n.filter != nil {
				if ok, err := shared.EnsureBool(s.n.filter.ValueFrom(row)); err != nil {
					return false, err
				} else if !ok {
					return true, nil
				}
			}

			return visitor(row)
		})
	})
}
