package indexes

import (
	"github.com/efritz/gostgres/internal/scan"
	"github.com/efritz/gostgres/internal/shared"
)

func (i *btreeIndex) Scanner(ctx scan.ScanContext, opts BtreeIndexScanOptions) (tidScanner, error) {
	stack := []*btreeNode{}
	current := i.root

	checkBounds := func(nodeValues []any, bounds [][]scanBound, expected shared.OrderType) bool {
		for j, bounds := range bounds[:min(len(bounds), len(nodeValues))] {
			for _, bound := range bounds {
				value, err := ctx.Evaluate(bound.expression, shared.Row{})
				if err != nil {
					return false
				}

				orderType := shared.CompareValues(nodeValues[j], value)
				if !(orderType == expected || (orderType == shared.OrderTypeEqual && bound.inclusive)) {
					return false
				}
			}
		}

		return true
	}

	withinLowerBound := func(values []any) bool { return checkBounds(values, opts.lowerBounds, shared.OrderTypeAfter) }
	withinUpperBound := func(values []any) bool { return checkBounds(values, opts.upperBounds, shared.OrderTypeBefore) }

	return tidScannerFunc(func() (int, error) {
		for current != nil || len(stack) > 0 {
			for current != nil {
				stack = append(stack, current)

				if opts.scanDirection != ScanDirectionBackward && withinLowerBound(current.values) {
					current = current.left
				} else if opts.scanDirection == ScanDirectionBackward == withinUpperBound(current.values) {
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

			if opts.scanDirection == ScanDirectionForward && withinUpperBound(node.values) {
				current = node.right
			} else if opts.scanDirection == ScanDirectionBackward && withinLowerBound(node.values) {
				current = node.left
			} else {
				current = nil
			}

			if withinLowerBound(node.values) && withinUpperBound(node.values) {
				return node.tid, nil
			}
		}

		return 0, scan.ErrNoRows
	}), nil
}

func (i *btreeIndex) extractTIDAndValuesFromRow(row shared.Row) (int, []any, error) {
	tid, err := row.TID()
	if err != nil {
		return 0, nil, err
	}

	values := []any{}
	for _, expression := range i.expressions {
		value, err := expression.Expression.ValueFrom(row)
		if err != nil {
			return 0, nil, err
		}

		values = append(values, value)
	}

	return tid, values, nil
}
