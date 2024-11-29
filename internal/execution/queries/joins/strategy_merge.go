package joins

import (
	"slices"

	"github.com/efritz/gostgres/internal/shared/fields"
	"github.com/efritz/gostgres/internal/shared/impls"
	"github.com/efritz/gostgres/internal/shared/ordering"
	"github.com/efritz/gostgres/internal/shared/rows"
	"github.com/efritz/gostgres/internal/shared/scan"
)

type logicalMergeJoinStrategy struct {
	n     *logicalJoinNode
	pairs []equalityPair
}

func (s *logicalMergeJoinStrategy) Ordering() impls.OrderExpression {
	// TODO - can add right fields as well?
	return s.n.left.Ordering()
}

func (s *logicalMergeJoinStrategy) Build(n *joinNode) joinStrategy {
	return &mergeJoinStrategy{
		n:      n,
		pairs:  s.pairs,
		fields: n.fields,
	}
}

//
//

type mergeJoinStrategy struct {
	n      *joinNode
	pairs  []equalityPair
	fields []fields.Field
}

func (s *mergeJoinStrategy) Name() string {
	return "merge"
}

func (s *mergeJoinStrategy) Scanner(ctx impls.ExecutionContext) (scan.RowScanner, error) {
	ctx.Log("Building Merge Join Strategy scanner")

	leftScanner, err := s.n.left.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	rightScanner, err := s.n.right.Scanner(ctx)
	if err != nil {
		return nil, err
	}

	markRestorer, ok := rightScanner.(scan.MarkRestorer)
	if !ok {
		panic("right scanner is not a MarkRestorer")
	}

	return &mergeJoinScanner{
		ctx:          ctx,
		strategy:     s,
		leftScanner:  leftScanner,
		rightScanner: rightScanner,
		markRestorer: markRestorer,
	}, nil
}

type mergeJoinScanner struct {
	ctx          impls.ExecutionContext
	strategy     *mergeJoinStrategy
	leftScanner  scan.RowScanner
	rightScanner scan.RowScanner
	markRestorer scan.MarkRestorer

	// state
	leftRow  *rows.Row
	rightRow *rows.Row
	mark     *rows.Row
	matching bool
}

func (s *mergeJoinScanner) Scan() (rows.Row, error) {
	s.ctx.Log("Scanning Merge Join Strategy")
	if s.leftRow == nil {
		// Get next row from left relation. This is necessary only on the first
		// iteration of the scan. After that, the left row is always non-nil.
		if err := s.advanceLeft(); err != nil {
			return rows.Row{}, err
		}
	}

	for {
		if s.rightRow == nil {
			// Get next row from right relation. This is necessary on the first
			// iteration of the scan, as well as immediately after we have found
			// a matching pair of rows and need to find the next one.
			if err := s.advanceRight(); err != nil {
				if err != scan.ErrNoRows {
					return rows.Row{}, err
				}

				if !s.matching {
					// When we exhaust the right relation, we only want to halt
					// when we will no longer restore a previous mark position.
					return rows.Row{}, err
				}
			}
		}

		// If we are had not yet found a pair of matching rows, we need to continue
		// to advance the smaller of the two relations until we find a pair that match.
		// When we find a match, we'll set the mark to the current right row and drop
		// to the next phase of the join logic.
		if !s.matching {
			if err := s.findNextMatch(); err != nil {
				return rows.Row{}, err
			}

			s.matching = true
			s.mark = s.rightRow
			s.markRestorer.Mark()
		}

		if s.rightRow != nil {
			// Check if the current pair of rows match. If they are, we'll emit the
			// pair and note that we need to advance the right relation on the next
			// iteration of the scan.
			if ot, err := s.compareRows(*s.leftRow, *s.rightRow); err != nil {
				return rows.Row{}, err
			} else if ot == ordering.OrderTypeEqual {
				row, err := rows.NewRow(s.strategy.fields, append(slices.Clone(s.leftRow.Values), s.rightRow.Values...))
				if err != nil {
					return rows.Row{}, err
				}

				s.rightRow = nil
				return row, nil
			}
		}

		// The current pair of rows did not match. We need to advance the left relation
		// to the next row. If the left relation matches our previous mark, we need to
		// restore the right relation to the mark and begin scanning from there. This is
		// necessary in the presence of duplicate values in the left relation.

		if err := s.advanceLeft(); err != nil {
			return rows.Row{}, err
		}

		if ot, err := s.compareRows(*s.leftRow, *s.mark); err != nil {
			return rows.Row{}, err
		} else if ot == ordering.OrderTypeEqual {
			s.rightRow = nil
			s.markRestorer.Restore()
		} else {
			s.matching = false
		}
	}
}

func (s *mergeJoinScanner) findNextMatch() error {
	for {
		ot, err := s.compareRows(*s.leftRow, *s.rightRow)
		if err != nil {
			return err
		}

		switch ot {
		case ordering.OrderTypeEqual:
			return nil

		case ordering.OrderTypeBefore:
			if err := s.advanceLeft(); err != nil {
				return err
			}

		case ordering.OrderTypeAfter:
			if err := s.advanceRight(); err != nil {
				return err
			}
		}
	}
}

func (s *mergeJoinScanner) advanceLeft() error {
	return scanIntoTarget(s.leftScanner, &s.leftRow)
}

func (s *mergeJoinScanner) advanceRight() error {
	return scanIntoTarget(s.rightScanner, &s.rightRow)
}

func scanIntoTarget(scanner scan.RowScanner, target **rows.Row) error {
	row, err := scanner.Scan()
	if err != nil {
		*target = nil
		return err
	}

	*target = &row
	return nil
}

func (s *mergeJoinScanner) compareRows(leftRow, rightRow rows.Row) (ordering.OrderType, error) {
	lValues, err := evaluatePair(s.ctx, s.strategy.pairs, leftOfPair, leftRow)
	if err != nil {
		return ordering.OrderTypeIncomparable, err
	}

	rValues, err := evaluatePair(s.ctx, s.strategy.pairs, rightOfPair, rightRow)
	if err != nil {
		return ordering.OrderTypeIncomparable, err
	}

	return ordering.CompareValueSlices(lValues, rValues), nil
}
