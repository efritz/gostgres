package joins

import (
	"slices"

	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

type hashJoinStrategy struct {
	n     *joinNode
	pairs []equalityPair
}

func (s *hashJoinStrategy) Name() string {
	return "hash"
}

func (s *hashJoinStrategy) Ordering() expressions.OrderExpression {
	return s.n.left.Ordering()
}

func (s *hashJoinStrategy) Scanner(ctx queries.Context) (scan.Scanner, error) {
	rightScanner, err := s.n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	h := map[uint64][]shared.Row{}
	if err := scan.VisitRows(rightScanner, func(row shared.Row) (bool, error) {
		keys, err := evaluatePair(ctx, s.pairs, rightOfPair, row)
		if err != nil {
			return false, err
		}

		key := shared.Hash(keys)
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

	return scan.ScannerFunc(func() (shared.Row, error) {
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
					return shared.NewRow(s.n.Fields(), append(slices.Clone(leftRow.Values), rightRow.Values...))
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

			// TODO - handle hash collision
			rightRows = h[shared.Hash(lKeys)]
		}
	}), nil
}
