package nodes

import "github.com/efritz/gostgres/internal/shared"

type hashJoinStrategy struct {
	n          *joinNode
	leftField  shared.Field
	rightField shared.Field
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
		i, err := shared.FindMatchingFieldIndex(s.rightField, row.Fields)
		if err != nil {
			return false, err
		}
		key := row.Values[i]

		h[key] = row
		return true, nil
	}); err != nil {
		return err
	}

	return s.n.left.Scan(func(row shared.Row) (bool, error) {
		i, err := shared.FindMatchingFieldIndex(s.leftField, row.Fields)
		if err != nil {
			return false, err
		}
		key := row.Values[i]

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
