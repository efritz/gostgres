package indexes

import (
	"github.com/efritz/gostgres/internal/expressions"
	"github.com/efritz/gostgres/internal/queries"
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

func (i *btreeIndex) Scanner(ctx queries.Context, opts BtreeIndexScanOptions) (tidScanner, error) {
	stack := []*btreeNode{}
	current := i.root

	checkBounds := func(nodeValues []any, bounds [][]scanBound, expected shared.OrderType) (bool, error) {
		for j, bounds := range bounds[:min(len(bounds), len(nodeValues))] {
			for _, bound := range bounds {
				value, err := ctx.Evaluate(bound.expression, shared.Row{})
				if err != nil {
					return false, err
				}

				orderType := shared.CompareValues(nodeValues[j], value)
				if !(orderType == expected || (orderType == shared.OrderTypeEqual && bound.inclusive)) {
					return false, nil
				}
			}
		}

		return true, nil
	}

	withinLowerBound := func(values []any) (bool, error) { return checkBounds(values, opts.lowerBounds, shared.OrderTypeAfter) }
	withinUpperBound := func(values []any) (bool, error) { return checkBounds(values, opts.upperBounds, shared.OrderTypeBefore) }

	return tidScannerFunc(func() (int64, error) {
		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)

				lowerOk, err := withinLowerBound(current.values)
				if err != nil {
					return 0, err
				}
				upperOk, err := withinUpperBound(current.values)
				if err != nil {
					return 0, err
				}

				if opts.scanDirection == ScanDirectionForward && lowerOk {
					current = current.left
				} else if opts.scanDirection == ScanDirectionBackward && upperOk {
					current = current.right
				} else {
					current = nil
				}
			}

			if len(stack) == 0 {
				break
			}

			idx := len(stack) - 1
			node := stack[idx]
			stack = stack[:idx]

			upperOk, err := withinUpperBound(node.values)
			if err != nil {
				return 0, err
			}
			lowerOk, err := withinLowerBound(node.values)
			if err != nil {
				return 0, err
			}

			if opts.scanDirection == ScanDirectionForward && upperOk {
				current = node.right
			} else if opts.scanDirection == ScanDirectionBackward && lowerOk {
				current = node.left
			} else {
				current = nil
			}

			if upperOk && lowerOk {
				return node.tid, nil
			}
		}

		return 0, scan.ErrNoRows
	}), nil
}

func (i *btreeIndex) extractTIDAndValuesFromRow(row shared.Row) (int64, []any, error) {
	tid, err := row.TID()
	if err != nil {
		return 0, nil, err
	}

	values := []any{}
	for _, expression := range i.expressions {
		value, err := expression.Expression.ValueFrom(expressions.EmptyContext, row)
		if err != nil {
			return 0, nil, err
		}

		values = append(values, value)
	}

	return tid, values, nil
}
