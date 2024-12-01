package join

import (
	"slices"

	"github.com/efritz/gostgres/internal/execution/queries/nodes"
	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
	"github.com/efritz/gostgres/internal/shared/utils"
)

type hashJoinStrategy struct {
	left   nodes.Node
	right  nodes.Node
	pairs  []EqualityPair
	fields []fields.Field
}

func NewHashJoinStrategy(left nodes.Node, right nodes.Node, pairs []EqualityPair, fields []fields.Field) nodes.JoinStrategy {
	return &hashJoinStrategy{
		left:   left,
		right:  right,
		pairs:  pairs,
		fields: fields,
	}
}

func (s *hashJoinStrategy) Name() string {
	return "hash"
}

func (s *hashJoinStrategy) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Hash Join Strategy scanner")

	rightScanner, err := s.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	h := map[uint64][]rows.Row{}
	if err := scan.VisitRows(rightScanner, func(row rows.Row) (bool, error) {
		keys, err := evaluatePair(ctx, s.pairs, rightOfPair, row)
		if err != nil {
			return false, err
		}

		key := utils.Hash(keys)
		h[key] = append(h[key], row)
		return true, nil
	}); err != nil {
		return nil, err
	}

	leftScanner, err := s.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	var leftRow rows.Row
	var rightRows []rows.Row

	return scan.RowScannerFunc(func() (rows.Row, error) {
		ctx.Log("Scanning Hash Join Strategy")

		for {
			for len(rightRows) > 0 {
				rightRow := rightRows[0]
				rightRows = rightRows[1:]

				lKeys, err := evaluatePair(ctx, s.pairs, leftOfPair, leftRow)
				if err != nil {
					return rows.Row{}, err
				}

				rKeys, err := evaluatePair(ctx, s.pairs, rightOfPair, rightRow)
				if err != nil {
					return rows.Row{}, err
				}

				if ordering.CompareValueSlices(lKeys, rKeys) == ordering.OrderTypeEqual {
					return rows.NewRow(s.fields, append(slices.Clone(leftRow.Values), rightRow.Values...))
				}
			}

			leftRow, err = leftScanner.Scan()
			if err != nil {
				return rows.Row{}, err
			}

			lKeys, err := evaluatePair(ctx, s.pairs, leftOfPair, leftRow)
			if err != nil {
				return rows.Row{}, err
			}

			// TODO - handle hash collision
			rightRows = h[utils.Hash(lKeys)]
		}
	}), nil
}
