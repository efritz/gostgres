package nodes

import (
	"github.com/efritz/gostgres/internal/shared"
)

type hashJoinStrategy struct {
	n     *joinNode
	pairs []equalityPair
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

	h := map[uint64][]shared.Row{}
	if err := VisitRows(rightScanner, func(row shared.Row) (bool, error) {
		keys, err := evaluatePair(ctx, s.pairs, rightOfPair, row)
		if err != nil {
			return false, err
		}

		key := hash(keys)
		h[key] = append(h[key], row)
		return true, nil
	}); err != nil {
		return nil, err
	}

	leftScanner, err := s.n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var leftRow shared.Row
	var rightRows []shared.Row

	return ScannerFunc(func() (shared.Row, error) {
		for {
			for len(rightRows) > 0 {
				rightRow := rightRows[0]
				rightRows = rightRows[1:]

				lKeys, err := evaluatePair(ctx, s.pairs, leftOfPair, leftRow)
				if err != nil {
					return shared.Row{}, err
				}

				rKeys, err := evaluatePair(ctx, s.pairs, rightOfPair, rightRow)
				if err != nil {
					return shared.Row{}, err
				}

				if shared.CompareValueSlices(lKeys, rKeys) == shared.OrderTypeEqual {
					return shared.NewRow(s.n.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))
				}
			}

			leftRow, err = leftScanner.Scan()
			if err != nil {
				return shared.Row{}, err
			}

			lKeys, err := evaluatePair(ctx, s.pairs, leftOfPair, leftRow)
			if err != nil {
				return shared.Row{}, err
			}

			rightRows = h[hash(lKeys)]
		}
	}), nil
}
