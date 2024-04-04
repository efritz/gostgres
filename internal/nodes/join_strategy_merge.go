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
		panic("WOAH MAN.") // TODO
	}

	var leftRow *shared.Row
	var rightRow *shared.Row
	var mark *shared.Row

	// TODO - convert this into a resumable generator
	var rows []shared.Row
	if _, err := func() (interface{}, error) {
		row1, err := leftScanner.Scan()
		if err != nil {
			return nil, err
		}
		leftRow = &row1

		row2, err := rightScanner.Scan()
		if err != nil {
			return nil, err
		}
		rightRow = &row2

		for {
			for {
				ot, err := s.compareRows(*leftRow, *rightRow)
				if err != nil {
					return nil, err
				}
				if ot == shared.OrderTypeEqual {
					break
				}
				if ot == shared.OrderTypeBefore {
					row, err := leftScanner.Scan()
					if err != nil {
						return nil, err
					}
					leftRow = &row
				}
				if ot == shared.OrderTypeAfter {
					row, err := rightScanner.Scan()
					if err != nil {
						return nil, err
					}
					rightRow = &row
				}
			}

			mark = rightRow
			markRestorer.Mark()

			for {
				for {
					ot, err := s.compareRows(*leftRow, *rightRow)
					if err != nil {
						return nil, err
					}
					if ot != shared.OrderTypeEqual {
						break
					}

					row, err := shared.NewRow(s.n.Fields(), append(copyValues(leftRow.Values), rightRow.Values...))
					if err != nil {
						return nil, err
					}
					rows = append(rows, row)

					row3, err := rightScanner.Scan()
					if err != nil {
						if err != ErrNoRows {
							return nil, err
						}

						break
					} else {
						rightRow = &row3
					}
				}

				row4, err := leftScanner.Scan()
				if err != nil {
					return nil, err
				}
				leftRow = &row4

				ot, err := s.compareRows(*leftRow, *mark)
				if err != nil {
					return nil, err
				}
				if ot == shared.OrderTypeEqual {
					markRestorer.Restore()

					row5, err := rightScanner.Scan()
					if err != nil {
						return nil, err
					}
					rightRow = &row5
				} else {
					break
				}
			}
		}
	}(); err != nil {
		if err != ErrNoRows {
			return nil, err
		}
	}

	i := 0
	return ScannerFunc(func() (shared.Row, error) {
		if i < len(rows) {
			row := rows[i]
			i++
			return row, nil
		}

		return shared.Row{}, ErrNoRows
	}), nil
}

func (s *mergeJoinStrategy) compareRows(leftRow, rightRow shared.Row) (shared.OrderType, error) {
	leftKey, err := s.left.ValueFrom(leftRow)
	if err != nil {
		return 0, err
	}

	rightKey, err := s.right.ValueFrom(rightRow)
	if err != nil {
		return 0, err
	}

	return shared.CompareValues(leftKey, rightKey), nil
}
