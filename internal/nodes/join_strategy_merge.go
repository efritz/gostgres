package nodes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/shared"
)

type mergeJoinStrategy struct {
	n     *joinNode
	left  expressions.Expression
	right expressions.Expression
}

func (s *mergeJoinStrategy) Name() string {
	return "merge"
}

func (s *mergeJoinStrategy) Ordering() OrderExpression {
	return nil // TODO - ordered on the left + right fields?
}

func (s *mergeJoinStrategy) Scanner() (Scanner, error) {
	leftScanner, err := s.n.left.Scanner()
	if err != nil {
		return nil, err
	}

	rightScanner, err := s.n.right.Scanner()
	if err != nil {
		return nil, err
	}

	markRestorer, ok := rightScanner.(MarkRestorer)
	if !ok {
		panic("right scanner is not a MarkRestorer")
	}

	return &mergeJoinScanner{
		s:            s,
		leftScanner:  leftScanner,
		rightScanner: rightScanner,
		markRestorer: markRestorer,
	}, nil
}

type mergeJoinScanner struct {
	s            *mergeJoinStrategy
	leftScanner  Scanner
	rightScanner Scanner
	markRestorer MarkRestorer

	// state
	leftRow  *shared.Row
	rightRow *shared.Row
	mark     *shared.Row
	matching bool
}

func (s *mergeJoinScanner) Scan() (shared.Row, error) {
	if s.leftRow == nil {
		// Get next row from left relation. This is necessary only on the first
		// iteration of the scan. After that, the left row is always non-nil.
		if err := s.advanceLeft(); err != nil {
			return shared.Row{}, err
		}
	}

	for {
		if s.rightRow == nil {
			// Get next row from right relation. This is necessary on the first
			// iteration of the scan, as well as immediately after we have found
			// a matching pair of rows and need to find the next one.
			if err := s.advanceRight(); err != nil {
				if err != ErrNoRows {
					return shared.Row{}, err
				}

				if !s.matching {
					// When we exhaust the right relation, we only want to halt
					// when we will no longer restore a previous mark position.
					return shared.Row{}, err
				}
			}
		}

		// If we are had not yet found a pair of matching rows, we need to continue
		// to advance the smaller of the two relations until we find a pair that match.
		// When we find a match, we'll set the mark to the current right row and drop
		// to the next phase of the join logic.
		if !s.matching {
			if err := s.findNextMatch(); err != nil {
				return shared.Row{}, err
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
				return shared.Row{}, err
			} else if ot == shared.OrderTypeEqual {
				row, err := shared.NewRow(s.s.n.Fields(), append(copyValues(s.leftRow.Values), s.rightRow.Values...))
				if err != nil {
					return shared.Row{}, err
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
			return shared.Row{}, err
		}

		if ot, err := s.compareRows(*s.leftRow, *s.mark); err != nil {
			return shared.Row{}, err
		} else if ot == shared.OrderTypeEqual {
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
		case shared.OrderTypeEqual:
			return nil

		case shared.OrderTypeBefore:
			if err := s.advanceLeft(); err != nil {
				return err
			}

		case shared.OrderTypeAfter:
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

func scanIntoTarget(scanner Scanner, target **shared.Row) error {
	row, err := scanner.Scan()
	if err != nil {
		*target = nil
		return err
	}

	*target = &row
	return nil
}

func (s *mergeJoinScanner) compareRows(leftRow, rightRow shared.Row) (shared.OrderType, error) {
	leftKey, err := s.s.left.ValueFrom(leftRow)
	if err != nil {
		return 0, err
	}

	rightKey, err := s.s.right.ValueFrom(rightRow)
	if err != nil {
		return 0, err
	}

	return shared.CompareValues(leftKey, rightKey), nil
}
